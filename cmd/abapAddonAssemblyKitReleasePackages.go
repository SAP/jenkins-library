package cmd

import (
	"encoding/json"
	"time"

	"github.com/SAP/jenkins-library/pkg/abap/aakaas"
	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
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

	client := piperhttp.Client{}

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runAbapAddonAssemblyKitReleasePackages(&config, telemetryData, &client, cpe, time.Duration(config.MaxRuntimeInMinutes)*time.Minute, time.Duration(config.PollingIntervalInSeconds)*time.Second)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapAddonAssemblyKitReleasePackages(config *abapAddonAssemblyKitReleasePackagesOptions, telemetryData *telemetry.CustomData, client piperhttp.Sender,
	cpe *abapAddonAssemblyKitReleasePackagesCommonPipelineEnvironment, maxRuntime time.Duration, pollingInterval time.Duration) error {
	conn := new(abapbuild.Connector)
	if err := conn.InitAAKaaS(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, client); err != nil {
		return err
	}
	var addonDescriptor abaputils.AddonDescriptor
	json.Unmarshal([]byte(config.AddonDescriptor), &addonDescriptor)

	err := checkInput(addonDescriptor.Repositories)
	if err != nil {
		return err
	}
	packagesWithReposLocked, packagesWithReposNotLocked := sortByStatus(addonDescriptor.Repositories, *conn)
	packagesWithReposLocked, err = releaseAndPoll(packagesWithReposLocked, maxRuntime, pollingInterval)
	if err != nil {
		return err
	}
	addonDescriptor.Repositories = sortingBack(packagesWithReposLocked, packagesWithReposNotLocked)
	log.Entry().Info("Writing package status to CommonPipelineEnvironment")
	cpe.abap.addonDescriptor = string(addonDescriptor.AsJSON())
	return nil
}

func sortingBack(packagesWithReposLocked []aakaas.PackageWithRepository, packagesWithReposNotLocked []aakaas.PackageWithRepository) []abaputils.Repository {
	var combinedRepos []abaputils.Repository
	for i := range packagesWithReposLocked {
		packagesWithReposLocked[i].Package.ChangeStatus(&packagesWithReposLocked[i].Repo)
		combinedRepos = append(combinedRepos, packagesWithReposLocked[i].Repo)
	}
	for i := range packagesWithReposNotLocked {
		combinedRepos = append(combinedRepos, packagesWithReposNotLocked[i].Repo)
	}
	return combinedRepos
}

func checkInput(repos []abaputils.Repository) error {
	for i := range repos {
		if repos[i].PackageName == "" {
			return errors.New("Parameter missing. Please provide the name of the package which should be released")
		}
	}
	return nil
}

func sortByStatus(repos []abaputils.Repository, conn abapbuild.Connector) ([]aakaas.PackageWithRepository, []aakaas.PackageWithRepository) {
	var packagesWithReposLocked []aakaas.PackageWithRepository
	var packagesWithReposNotLocked []aakaas.PackageWithRepository
	for i := range repos {
		var pack aakaas.Package
		pack.InitPackage(repos[i], conn)
		pWR := aakaas.PackageWithRepository{
			Package: pack,
			Repo:    repos[i],
		}
		if pack.Status == "L" {
			packagesWithReposLocked = append(packagesWithReposLocked, pWR)
		} else {
			log.Entry().Infof("Package %s has status %s, cannot release this package", pack.PackageName, pack.Status)
			packagesWithReposNotLocked = append(packagesWithReposNotLocked, pWR)
		}
	}
	return packagesWithReposLocked, packagesWithReposNotLocked
}

func releaseAndPoll(pckgWR []aakaas.PackageWithRepository, maxRuntime time.Duration, pollingInterval time.Duration) ([]aakaas.PackageWithRepository, error) {
	timeout := time.After(maxRuntime)
	ticker := time.Tick(pollingInterval)

	for {
		select {
		case <-timeout:
			return pckgWR, errors.New("Timed out")
		case <-ticker:
			var allFinished bool = true
			for i := range pckgWR {
				if pckgWR[i].Package.Status != aakaas.PackageStatusReleased {
					err := pckgWR[i].Package.Release()
					// if there is an error, release is not yet finished
					if err != nil {
						log.Entry().Infof("Release of %s is not yet finished, check again in %s", pckgWR[i].Package.PackageName, pollingInterval)
						allFinished = false
					}
				}
			}
			if allFinished {
				log.Entry().Infof("Release of package(s) was successful")
				return pckgWR, nil
			}
		}
	}
}
