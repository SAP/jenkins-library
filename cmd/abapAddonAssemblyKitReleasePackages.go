package cmd

import (
	"encoding/json"
	"time"

	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func abapAddonAssemblyKitReleasePackages(config abapAddonAssemblyKitReleasePackagesOptions, telemetryData *telemetry.CustomData, cpe *abapAddonAssemblyKitReleasePackagesCommonPipelineEnvironment) {
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
	err := runAbapAddonAssemblyKitReleasePackages(&config, telemetryData, &autils, &client, cpe)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapAddonAssemblyKitReleasePackages(config *abapAddonAssemblyKitReleasePackagesOptions, telemetryData *telemetry.CustomData, com abaputils.Communication, client piperhttp.Sender, cpe *abapAddonAssemblyKitReleasePackagesCommonPipelineEnvironment) error {
	conn := new(connector)
	conn.initAAK(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, &piperhttp.Client{})
	var addonDescriptor abaputils.AddonDescriptor
	json.Unmarshal([]byte(config.AddonDescriptor), &addonDescriptor)

	packagesWithReposLocked, packagesWithReposNotLocked := sortByStatus(addonDescriptor.Repositories, *conn)
	packagesWithReposLocked, err := releaseAndPoll(packagesWithReposLocked, 5, 30)
	if err != nil {
		return err
	}
	addonDescriptor.Repositories = sortingBack(packagesWithReposLocked, packagesWithReposNotLocked)
	log.Entry().Info("Writing package status to CommonPipelineEnvironment")
	backToCPE, _ := json.Marshal(addonDescriptor)
	cpe.abap.addonDescriptor = string(backToCPE)
	return nil
}

func sortingBack(packagesWithReposLocked []packageWithRepository, packagesWithReposNotLocked []packageWithRepository) []abaputils.Repository {
	var combinedRepos []abaputils.Repository
	for i := range packagesWithReposLocked {
		packagesWithReposLocked[i].p.changeStatus(&packagesWithReposLocked[i].repo)
		combinedRepos = append(combinedRepos, packagesWithReposLocked[i].repo)
	}
	for i := range packagesWithReposNotLocked {
		combinedRepos = append(combinedRepos, packagesWithReposNotLocked[i].repo)
	}
	return combinedRepos
}

func sortByStatus(repos []abaputils.Repository, conn connector) ([]packageWithRepository, []packageWithRepository) {
	var packagesWithReposLocked []packageWithRepository
	var packagesWithReposNotLocked []packageWithRepository
	for i := range repos {
		var p pckg
		p.init(repos[i], conn)
		pWR := packageWithRepository{
			p:    p,
			repo: repos[i],
		}
		if p.Status == "L" {
			packagesWithReposLocked = append(packagesWithReposLocked, pWR)
		} else {
			log.Entry().Infof("Package %s has status %s, cannot release this package", p.PackageName, p.Status)
			packagesWithReposNotLocked = append(packagesWithReposNotLocked, pWR)
		}
	}
	return packagesWithReposLocked, packagesWithReposNotLocked
}

func releaseAndPoll(pckgWR []packageWithRepository, maxRuntimeInMinutes time.Duration, pollIntervalsInSeconds time.Duration) ([]packageWithRepository, error) {
	timeout := time.After(maxRuntimeInMinutes * time.Minute)
	ticker := time.Tick(pollIntervalsInSeconds * time.Second)

	for {
		select {
		case <-timeout:
			return pckgWR, errors.New("Timed out")
		case <-ticker:
			var allFinished bool = true
			for i := range pckgWR {
				if pckgWR[i].p.Status != "R" {
					err := pckgWR[i].p.release()
					// if there is an error, release is not yet finished
					if err != nil {
						log.Entry().Infof("Release of %s is not yet finished, check again in %02d seconds", pckgWR[i].p.PackageName, pollIntervalsInSeconds)
						allFinished = false
					}
				}
			}
			if allFinished {
				log.Entry().Infof("Release of package(s) was succesful")
				return pckgWR, nil
			}
		}
	}
}

func (p *pckg) release() error {
	var body []byte
	var err error
	if p.PackageName == "" {
		return errors.New("Parameter missing. Please provide the name of the package which should be released")
	}
	log.Entry().Infof("Release package %s", p.PackageName)
	p.connector.getToken("/odata/aas_ocs_package")
	appendum := "/odata/aas_ocs_package/ReleasePackage?Name='" + p.PackageName + "'"
	body, err = p.connector.post(appendum, "")
	if err != nil {
		return err
	}
	var jPck jsonPackage
	json.Unmarshal(body, &jPck)
	p.Status = jPck.Package.Status
	return nil
}
