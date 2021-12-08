package cmd

import (
	"fmt"
	"time"

	"github.com/SAP/jenkins-library/pkg/abap/aakaas"
	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func abapAddonAssemblyKitReserveNextPackages(config abapAddonAssemblyKitReserveNextPackagesOptions, telemetryData *telemetry.CustomData, cpe *abapAddonAssemblyKitReserveNextPackagesCommonPipelineEnvironment) {
	utils := aakaas.NewAakBundleWithTime(time.Duration(config.MaxRuntimeInMinutes), time.Duration(config.PollingIntervalInSeconds))
	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runAbapAddonAssemblyKitReserveNextPackages(&config, telemetryData, &utils, cpe)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapAddonAssemblyKitReserveNextPackages(config *abapAddonAssemblyKitReserveNextPackagesOptions, telemetryData *telemetry.CustomData, utils *aakaas.AakUtils, cpe *abapAddonAssemblyKitReserveNextPackagesCommonPipelineEnvironment) error {

	conn := new(abapbuild.Connector)
	if err := conn.InitAAKaaS(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, *utils); err != nil {
		return err
	}

	addonDescriptor := new(abaputils.AddonDescriptor)
	if err := addonDescriptor.InitFromJSONstring(config.AddonDescriptor); err != nil {
		return errors.Wrap(err, "Reading AddonDescriptor failed [Make sure abapAddonAssemblyKit...CheckCVs|CheckPV steps have been run before]")
	}

	packagesWithRepos, err := reservePackages(addonDescriptor.Repositories, *conn)
	if err != nil {
		return err
	}

	err = pollReserveNextPackages(packagesWithRepos, utils)
	if err != nil {
		return err
	}

	addonDescriptor.Repositories, err = checkAndCopyFieldsToRepositories(packagesWithRepos)
	if err != nil {
		return err
	}

	log.Entry().Info("Writing package names, types, status, namespace and predecessorCommitID to CommonPipelineEnvironment")
	cpe.abap.addonDescriptor = addonDescriptor.AsJSONstring()
	return nil
}

func checkAndCopyFieldsToRepositories(pckgWR []aakaas.PackageWithRepository) ([]abaputils.Repository, error) {
	var repos []abaputils.Repository

	log.Entry().Infof("%-30v | %-20v | %-6v | %-40v | %-40v", "Software Component", "Package", "Status", "CommitID (from addon.yml)", "PredecessorCommitID (from AAKaaS)")

	for i := range pckgWR {

		log.Entry().Infof("%-30v | %-20v | %-6v | %-40v | %-40v", pckgWR[i].Repo.Name, pckgWR[i].Package.PackageName, pckgWR[i].Package.Status, pckgWR[i].Repo.CommitID, pckgWR[i].Package.PredecessorCommitID)

		if pckgWR[i].Package.Status == aakaas.PackageStatusReleased {
			//Ensure for Packages with Status R that CommitID of package = the one from addon.yml, beware of short commitID in addon.yml
			addonYAMLcommitIDLength := len(pckgWR[i].Repo.CommitID)
			if len(pckgWR[i].Package.CommitID) < addonYAMLcommitIDLength {
				return repos, errors.New("Provided CommitIDs have wrong length: " + pckgWR[i].Repo.CommitID + "(addon.yml) longer than the one from AAKaaS " + pckgWR[i].Package.CommitID)
			}
			packageCommitIDsubsting := pckgWR[i].Package.CommitID[0:addonYAMLcommitIDLength]
			if pckgWR[i].Repo.CommitID != packageCommitIDsubsting {
				log.Entry().Error("package " + pckgWR[i].Package.PackageName + " was already build but with commit " + pckgWR[i].Package.CommitID + ", not with " + pckgWR[i].Repo.CommitID)
				log.Entry().Error("If you want to build a new package make sure to increase the dotted-version-string in addon.yml")
				log.Entry().Error("If you do NOT want to build a new package enter the commitID " + pckgWR[i].Package.CommitID + " for software component " + pckgWR[i].Repo.Name + " in addon.yml")
				return repos, errors.New("commit of released package does not match with addon.yml")
			}
		} else if pckgWR[i].Package.PredecessorCommitID != "" {
			//Check for newly reserved packages which are to be build that CommitID from addon.yml != PreviousCommitID [this will result in an error as no delta can be calculated]
			addonYAMLcommitIDLength := len(pckgWR[i].Repo.CommitID)
			if len(pckgWR[i].Package.PredecessorCommitID) < addonYAMLcommitIDLength {
				return repos, errors.New("Provided CommitIDs have wrong length: " + pckgWR[i].Repo.CommitID + "(addon.yml) longer than the one from AAKaaS " + pckgWR[i].Package.CommitID)
			}
			packagePredecessorCommitIDsubsting := pckgWR[i].Package.PredecessorCommitID[0:addonYAMLcommitIDLength]
			if pckgWR[i].Repo.CommitID == packagePredecessorCommitIDsubsting {
				return repos, errors.New("CommitID of package " + pckgWR[i].Package.PackageName + " is the same as the one of the predecessor package. Make sure to change both the dotted-version-string AND the commitID in addon.yml")
			}
		}

		pckgWR[i].Package.CopyFieldsToRepo(&pckgWR[i].Repo)
		repos = append(repos, pckgWR[i].Repo)
	}
	return repos, nil
}

func pollReserveNextPackages(pckgWR []aakaas.PackageWithRepository, utils *aakaas.AakUtils) error {
	pollingInterval := (*utils).GetPollingInterval()
	timeout := time.After((*utils).GetMaxRuntime())
	ticker := time.Tick(pollingInterval)
	for {
		select {
		case <-timeout:
			return errors.New("Timed out")
		case <-ticker:
			var allFinished bool = true
			for i := range pckgWR {
				err := pckgWR[i].Package.GetPackageAndNamespace()
				// if there is an error, reservation is not yet finished
				if err != nil {
					log.Entry().Infof("Reservation of %s is not yet finished, check again in %s", pckgWR[i].Package.PackageName, pollingInterval)
					allFinished = false
				} else {
					switch pckgWR[i].Package.Status {
					case aakaas.PackageStatusLocked:
						return fmt.Errorf("Package %s has invalid status 'locked'", pckgWR[i].Package.PackageName)
					case aakaas.PackageStatusCreationTriggered:
						log.Entry().Infof("Reservation of %s is still running with status 'creation triggered', check again in %s", pckgWR[i].Package.PackageName, pollingInterval)
						allFinished = false
					case aakaas.PackageStatusPlanned:
						log.Entry().Infof("Reservation of %s was successful with status 'planned'", pckgWR[i].Package.PackageName)
					case aakaas.PackageStatusReleased:
						log.Entry().Infof("Reservation of %s not needed, package is already in status 'released'", pckgWR[i].Package.PackageName)
					default:
						return fmt.Errorf("Package %s has unknown status '%s'", pckgWR[i].Package.PackageName, pckgWR[i].Package.Status)
					}
				}
			}
			if allFinished {
				log.Entry().Infof("Reservation of package(s) was successful")
				return nil
			}
		}
	}
}

func reservePackages(repositories []abaputils.Repository, conn abapbuild.Connector) ([]aakaas.PackageWithRepository, error) {
	var packagesWithRepos []aakaas.PackageWithRepository
	for i := range repositories {
		var p aakaas.Package
		p.InitPackage(repositories[i], conn)
		err := p.ReserveNext()
		if err != nil {
			return packagesWithRepos, err
		}
		pWR := aakaas.PackageWithRepository{
			Package: p,
			Repo:    repositories[i],
		}
		packagesWithRepos = append(packagesWithRepos, pWR)
	}
	return packagesWithRepos, nil
}
