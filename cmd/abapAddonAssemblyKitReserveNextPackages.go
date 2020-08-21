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

	var addonDescriptor abaputils.AddonDescriptor
	json.Unmarshal([]byte(config.AddonDescriptor), &addonDescriptor)

	for i, repo := range addonDescriptor.Repositories {
		var p pckg
		p.init(repo, *conn)
		// TODO soll danach gepollt werden? glaub nicht..
		err := p.reserveNext()
		if err != nil {
			return err
		}
		// TODO kann gelöscht werden nachdem Dirk die Änderungen gemacht hat
		err = p.get()
		if err != nil {
			return err
		}
		// TODO status L => Fehler, da es nicht auftreten sollte
		addonDescriptor.Repositories[i] = p.addFields(repo)
	}
	backToCPE, _ := json.Marshal(addonDescriptor)
	cpe.abap.addonDescriptor = string(backToCPE)
	return nil
}

// TODO noch mehr übertragen?
func (p *pckg) init(repo abaputils.Repository, conn connector) {
	p.connector = conn
	p.ComponentName = repo.Name
	p.VersionYAML = repo.Version
	p.PackageName = repo.PackageName
}

// TODO genug?
func (p *pckg) addFields(initialRepo abaputils.Repository) abaputils.Repository {
	var repo abaputils.Repository
	repo = initialRepo
	repo.PackageName = p.PackageName
	repo.PackageType = p.Type
	repo.PredecessorCommitID = p.PredecessorCommitID
	repo.Status = p.Status
	repo.Namespace = p.Namespace
	return repo
}

func (p *pckg) reserveNext() error {
	p.connector.getToken("/odata/aas_ocs_package")
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
