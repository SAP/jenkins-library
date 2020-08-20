package cmd

import (
	"encoding/json"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func abapAddonAssemblyKitReserveNextPackages(config abapAddonAssemblyKitReserveNextPackagesOptions, telemetryData *telemetry.CustomData, cpe *abapAddonAssemblyKitReserveNextPackagesCommonPipelineEnvironment) {
	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	var autils = abaputils.AbapUtils{
		Exec: &c,
	}
	client := piperhttp.Client{}

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runAbapAddonAssemblyKitReserveNextPackages(&config, telemetryData, &autils, &client, cpe)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapAddonAssemblyKitReserveNextPackages(config *abapAddonAssemblyKitReserveNextPackagesOptions, telemetryData *telemetry.CustomData, com abaputils.Communication, client piperhttp.Sender, cpe *abapAddonAssemblyKitReserveNextPackagesCommonPipelineEnvironment) error {
	conn := new(connector)
	conn.initAAK(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, &piperhttp.Client{})
	var repos []abaputils.Repository
	json.Unmarshal([]byte(config.Repositories), &repos)

	// //TODO soll hier gepollt werden auf den status? ist bei dem ersten reserve next package und dann get das get leer?
	// //TODO laut Dirk soll wenn das Paket nicht status P hat der Assembly schritt nicht ausgeführt werden => soll ich diese info per commonEnv weitergeben und der assembly step dann
	// // nicht ausgeführt werden? oder soll das irgendwie in dieser großen pipeline auseinander gesteuert werden
	var reposBackToCPE []abaputils.Repository
	for _, repo := range repos {
		var p pckg
		p.init(repo, *conn)
		err := p.reserveNext()
		if err != nil {
			return err
		}
		err = p.get()
		if err != nil {
			return err
		}
		repoBack := p.addFields(repo)
		reposBackToCPE = append(reposBackToCPE, repoBack)
	}
	backToCPE, _ := json.Marshal(reposBackToCPE)
	cpe.abap.repositories = string(backToCPE)
	return nil
}

// TODO noch mehr übertragen?
func (p *pckg) init(repo abaputils.Repository, conn connector) {
	p.connector = conn
	p.ComponentName = repo.Name
	p.VersionYAML = repo.Version
	p.PackageName = repo.PackageName
}

// TODO change name
func (p *pckg) addFields(repo2 abaputils.Repository) abaputils.Repository {
	var repo abaputils.Repository
	repo = repo2
	repo.PackageName = p.PackageName
	repo.PackageType = p.Type
	repo.PredecessorCommitID = p.PredecessorCommitID
	repo.Status = p.Status
	repo.Namespace = p.Namespace
	return repo
}

type jsonPackage struct {
	DeterminePackage struct {
		Package *pckg `json:"DeterminePackageForScv"`
	} `json:"d"`
}

type jsonPackageFromGet struct {
	Package *pckg `json:"d"`
}

type pckg struct {
	connector
	ComponentName       string
	PackageName         string `json:"Name"`
	VersionYAML         string
	Type                string `json:"Type"`
	PredecessorCommitID string `json:"PredecessorCommitId"`
	Status              string `json:"Status"`
	Namespace           string `json:"Namespace"`
}

func (p *pckg) reserveNext() error {
	p.connector.getToken()
	appendum := "/odata/aas_ocs_package/DeterminePackageForScv?Name='" + p.ComponentName + "'&Version='" + p.VersionYAML + "'"
	body, err := p.connector.post(appendum, "")
	if err != nil {
		return err
	}
	var jPck jsonPackage
	json.Unmarshal(body, &jPck)
	p.PackageName = jPck.DeterminePackage.Package.PackageName
	p.Type = jPck.DeterminePackage.Package.Type
	p.PredecessorCommitID = jPck.DeterminePackage.Package.PredecessorCommitID
	p.Status = jPck.DeterminePackage.Package.Status
	p.Namespace = jPck.DeterminePackage.Package.Namespace
	return nil
}

func (p *pckg) get() error {
	appendum := "/odata/aas_ocs_package/OcsPackageSet('" + p.PackageName + "')"
	body, err := p.connector.get(appendum)
	if err != nil {
		return err
	}
	var jPck jsonPackageFromGet
	json.Unmarshal(body, &jPck)
	p.Status = jPck.Package.Status
	p.Namespace = jPck.Package.Namespace
	return nil
}

//TODO das sollte irgendwie zum reuse gepackt werden und auch mit dem post von dem assembly step zusammengeführt werden
// func (conn connector) post2(appendum string, importBody string) ([]byte, error) {
// 	url := conn.Baseurl + appendum
// 	var response *http.Response
// 	var err error
// 	if importBody == "" {
// 		response, err = conn.Client.SendRequest("POST", url, nil, conn.Header, nil)
// 	} else {
// 		response, err = conn.Client.SendRequest("POST", url, bytes.NewBuffer([]byte(importBody)), conn.Header, nil)
// 	}
// 	if err != nil {
// 		if response == nil {
// 			return nil, errors.Wrap(err, "Post failed")
// 		}
// 		defer response.Body.Close()
// 		errorbody, _ := ioutil.ReadAll(response.Body)
// 		return errorbody, errors.Wrapf(err, "Post failed: %v", string(errorbody))
// 	}
// 	defer response.Body.Close()
// 	body, err := ioutil.ReadAll(response.Body)
// 	return body, err
// }
