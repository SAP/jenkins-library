package cmd

import (
	"encoding/json"
	"time"

	"github.com/SAP/jenkins-library/pkg/abap/aakaas"
	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func abapAddonAssemblyKitReleasePackages(config abapAddonAssemblyKitReleasePackagesOptions, telemetryData *telemetry.CustomData, cpe *abapAddonAssemblyKitReleasePackagesCommonPipelineEnvironment) {
	utils := aakaas.NewAakBundleWithTime(time.Duration(config.MaxRuntimeInMinutes), time.Duration(config.PollingIntervalInSeconds))
	telemetryData.BuildTool = "AAKaaS"

	if err := runAbapAddonAssemblyKitReleasePackages(&config, &utils, cpe); err != nil {
		telemetryData.ErrorCode = err.Error()
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapAddonAssemblyKitReleasePackages(config *abapAddonAssemblyKitReleasePackagesOptions, utils *aakaas.AakUtils,
	cpe *abapAddonAssemblyKitReleasePackagesCommonPipelineEnvironment) error {
	conn := new(abapbuild.Connector)
	if err := conn.InitAAKaaS(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, *utils, config.AbapAddonAssemblyKitOriginHash, config.AbapAddonAssemblyKitCertificateFile, config.AbapAddonAssemblyKitCertificatePass); err != nil {
		return err
	}
	var addonDescriptor abaputils.AddonDescriptor
	if err := json.Unmarshal([]byte(config.AddonDescriptor), &addonDescriptor); err != nil {
		return err
	}

	err := checkInput(addonDescriptor.Repositories)
	if err != nil {
		return err
	}
	packagesWithReposLocked, packagesWithReposNotLocked := sortByStatus(addonDescriptor.Repositories, *conn)
	packagesWithReposLocked, err = releaseAndPoll(packagesWithReposLocked, utils)
	if err != nil {
		var numberOfReleasedPackages int
		for i := range packagesWithReposLocked {
			if packagesWithReposLocked[i].Package.Status == aakaas.PackageStatusReleased {
				numberOfReleasedPackages += 1
			}
		}
		if numberOfReleasedPackages == 0 {
			return errors.Wrap(err, "Release of all packages failed/timed out - Aborting as abapEnvironmentAssembleConfirm step is not needed")
		} else {
			log.Entry().WithError(err).Warn("Release of at least one package failed/timed out - Continuing anyway to enable abapEnvironmentAssembleConfirm to run")
		}
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
		if pack.Status == aakaas.PackageStatusLocked {
			packagesWithReposLocked = append(packagesWithReposLocked, pWR)
		} else {
			log.Entry().Infof("Package %s has status %s, cannot release this package", pack.PackageName, pack.Status)
			packagesWithReposNotLocked = append(packagesWithReposNotLocked, pWR)
		}
	}
	return packagesWithReposLocked, packagesWithReposNotLocked
}

func releaseAndPoll(pckgWR []aakaas.PackageWithRepository, utils *aakaas.AakUtils) ([]aakaas.PackageWithRepository, error) {
	timeout := time.After((*utils).GetMaxRuntime())
	ticker := time.Tick((*utils).GetPollingInterval())

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
						log.Entry().Error(err)
						log.Entry().Infof("Release of %s is not yet finished, check again in %s", pckgWR[i].Package.PackageName, (*utils).GetPollingInterval())
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
