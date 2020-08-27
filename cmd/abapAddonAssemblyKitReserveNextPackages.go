package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
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

	packagesWithRepos, err := reservePackages(addonDescriptor.Repositories, *conn)
	if err != nil {
		return err
	}
	//TODO zeiten anpassen
	err = pollReserveNextPackages(packagesWithRepos, 30, 5)
	addonDescriptor.Repositories = addFieldsToRepository(packagesWithRepos)
	log.Entry().Info("Writing package names, types, status, namespace and predecessorCommitID to CommonPipelineEnvironment")
	backToCPE, _ := json.Marshal(addonDescriptor)
	cpe.abap.addonDescriptor = string(backToCPE)
	return nil
}
func addFieldsToRepository(pckgWR []packageWithRepository) []abaputils.Repository {
	var repos []abaputils.Repository
	for i := range pckgWR {
		pckgWR[i].p.addFields(&pckgWR[i].repo)
		repos = append(repos, pckgWR[i].repo)
	}
	return repos
}

func pollReserveNextPackages(pckgWR []packageWithRepository, maxRuntimeInMinutes time.Duration, pollIntervalsInSeconds time.Duration) error {
	timeout := time.After(maxRuntimeInMinutes * time.Minute)
	ticker := time.Tick(pollIntervalsInSeconds * time.Second)
	for {
		select {
		case <-timeout:
			return errors.New("Timed out")
		case <-ticker:
			var allFinished bool = true
			for i := range pckgWR {
				err := pckgWR[i].p.get()
				// if there is an error, reservation is not yet finished
				if err != nil {
					log.Entry().Infof("Reservation of %s is not yet finished, check again in %02d seconds", pckgWR[i].p, pollIntervalsInSeconds)
					allFinished = false
				} else {
					switch pckgWR[i].p.Status {
					case "L":
						return fmt.Errorf("Package %s has invalid status L", pckgWR[i].p)
					case "C":
						log.Entry().Infof("Reservation of %s is not yet finished, check again in %02d seconds", pckgWR[i].p, pollIntervalsInSeconds)
						allFinished = false
					}
				}
			}
			if allFinished {
				log.Entry().Infof("Reservation of package(s) was succesful")
				return nil
			}
		}
	}
}

func reservePackages(repositories []abaputils.Repository, conn connector) ([]packageWithRepository, error) {
	var packagesWithRepos []packageWithRepository
	for i := range repositories {
		var p pckg
		p.init(repositories[i], conn)
		err := p.reserveNext()
		if err != nil {
			return packagesWithRepos, err
		}
		pWR := packageWithRepository{
			p:    p,
			repo: repositories[i],
		}
		packagesWithRepos = append(packagesWithRepos, pWR)
	}
	return packagesWithRepos, nil
}

func (p *pckg) init(repo abaputils.Repository, conn connector) {
	p.connector = conn
	p.ComponentName = repo.Name
	p.VersionYAML = repo.VersionYAML
	p.PackageName = repo.PackageName
	p.Status = repo.Status
}

func (p *pckg) addFields(initialRepo *abaputils.Repository) {
	initialRepo.PackageName = p.PackageName
	initialRepo.PackageType = p.Type
	initialRepo.PredecessorCommitID = p.PredecessorCommitID
	initialRepo.Status = p.Status
	initialRepo.Namespace = p.Namespace
	log.Entry().Infof("Package name %s, type %s, status %s, namespace %s, predecessorCommitID %s", p.PackageName, p.Type, p.Status, p.Namespace, p.PredecessorCommitID)
}

func (p *pckg) reserveNext() error {
	if p.ComponentName == "" || p.VersionYAML == "" {
		return errors.New("Parameters missing. Please provide the name and version of the component")
	}
	log.Entry().Infof("Reserve package for %s version %s", p.ComponentName, p.VersionYAML)
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
	log.Entry().Infof("Reservation of package %s started", p.PackageName)
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

type packageWithRepository struct {
	p    pckg
	repo abaputils.Repository
}
