package cmd

import (
	"fmt"
	"time"

	"github.com/SAP/jenkins-library/pkg/abap/aakaas"
	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"

	"errors"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func abapAddonAssemblyKitReserveNextPackages(config abapAddonAssemblyKitReserveNextPackagesOptions, telemetryData *telemetry.CustomData, cpe *abapAddonAssemblyKitReserveNextPackagesCommonPipelineEnvironment) {
	utils := aakaas.NewAakBundleWithTime(time.Duration(config.MaxRuntimeInMinutes), time.Duration(config.PollingIntervalInSeconds))
	telemetryData.BuildTool = "AAKaaS"

	if err := runAbapAddonAssemblyKitReserveNextPackages(&config, &utils, cpe); err != nil {
		telemetryData.ErrorCode = err.Error()
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapAddonAssemblyKitReserveNextPackages(config *abapAddonAssemblyKitReserveNextPackagesOptions, utils *aakaas.AakUtils,
	cpe *abapAddonAssemblyKitReserveNextPackagesCommonPipelineEnvironment) error {

	log.Entry().Info("╔═════════════════════════════════════════╗")
	log.Entry().Info("║ abapAddonAssemblyKitReserveNextPackages ║")
	log.Entry().Info("╚═════════════════════════════════════════╝")

	log.Entry().Infof("... initializing connection to AAKaaS @ %v", config.AbapAddonAssemblyKitEndpoint)
	conn := new(abapbuild.Connector)
	if err := conn.InitAAKaaS(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, *utils, config.AbapAddonAssemblyKitOriginHash, config.AbapAddonAssemblyKitCertificateFile, config.AbapAddonAssemblyKitCertificatePass); err != nil {
		return err
	}

	log.Entry().Info("... reading AddonDescriptor (Software Component, Version) from CommonPipelineEnvironment")
	addonDescriptor := new(abaputils.AddonDescriptor)
	if err := addonDescriptor.InitFromJSONstring(config.AddonDescriptor); err != nil {
		return fmt.Errorf("Reading AddonDescriptor failed [Make sure abapAddonAssemblyKit...CheckCVs|CheckPV steps have been run before]: %w", err)
	}

	log.Entry().Info("╭────────────────────────────────┬──────────────────────╮")
	log.Entry().Info("│ Software Component             │ Version              │")
	log.Entry().Info("├────────────────────────────────┼──────────────────────┤")
	for i := range addonDescriptor.Repositories {
		log.Entry().Infof("│ %-30v │ %-20v │", addonDescriptor.Repositories[i].Name, addonDescriptor.Repositories[i].VersionYAML)
	}
	log.Entry().Info("╰────────────────────────────────┴──────────────────────╯")

	packagesWithRepos, err := reservePackages(addonDescriptor.Repositories, *conn)
	if err != nil {
		return err
	}

	log.Entry().Info("... checking for ongoing Reservations")
	if err = pollReserveNextPackages(packagesWithRepos, utils); err != nil {
		return err
	}

	log.Entry().Info("┏━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━┳━━━━━━━━┳━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓")
	log.Entry().Infof("┃ %-30v ┃ %-20v ┃ %-5v ┃ %-6v ┃ %-10v ┃ %-40v ┃ %-40v ┃", "Software Component", "Package Name", "Type", "Status", "Namespace", "CommitID (from addon.yml)", "PredecessorCommitID (from AAKaaS)")
	log.Entry().Info("┣━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╋━━━━━━━━━━━━━━━━━━━━━━╋━━━━━━━╋━━━━━━━━╋━━━━━━━━━━━━╋━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━╋━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┫")
	for i := range packagesWithRepos {
		log.Entry().Infof("┃ %-30v ┃ %-20v ┃ %-5v ┃ %-6v ┃ %-10v ┃ %-40v ┃ %-40v ┃", packagesWithRepos[i].Repo.Name, packagesWithRepos[i].Package.PackageName, packagesWithRepos[i].Package.Type, packagesWithRepos[i].Package.Status, packagesWithRepos[i].Package.Namespace, packagesWithRepos[i].Repo.CommitID, packagesWithRepos[i].Package.PredecessorCommitID)
	}
	log.Entry().Info("┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┻━━━━━━━━━━━━━━━━━━━━━━┻━━━━━━━┻━━━━━━━━┻━━━━━━━━━━━━┻━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┻━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛")

	log.Entry().Info("... checking and processing provided and received data")
	if addonDescriptor.Repositories, err = checkAndCopyFieldsToRepositories(packagesWithRepos); err != nil {
		return err
	}

	log.Entry().Info("... writing AddonDescriptor (package name, type, status, namespace and predecessorCommitID) back to CommonPipelineEnvironment")
	cpe.abap.addonDescriptor = addonDescriptor.AsJSONstring()
	return nil
}

func checkAndCopyFieldsToRepositories(pckgWR []aakaas.PackageWithRepository) ([]abaputils.Repository, error) {
	var repos []abaputils.Repository
	var checkFailure error = nil
	for i := range pckgWR {
		if pckgWR[i].Package.Status != aakaas.PackageStatusReleased && pckgWR[i].Package.Namespace == "" {
			checkFailure = errors.New("AAKaaS returned a response with empty Namespace which indicates a configuration error")
		}

		checkFailure = checkCommitID(pckgWR, i, checkFailure)

		pckgWR[i].Package.CopyFieldsToRepo(&pckgWR[i].Repo)
		repos = append(repos, pckgWR[i].Repo)
	}
	return repos, checkFailure
}

func checkCommitID(pckgWR []aakaas.PackageWithRepository, i int, checkFailure error) error {
	if !pckgWR[i].Repo.UseClassicCTS {
		if pckgWR[i].Package.Status == aakaas.PackageStatusReleased {
			checkFailure = checkCommitIDSameAsGiven(pckgWR, i, checkFailure)
		} else if pckgWR[i].Package.PredecessorCommitID != "" {
			checkFailure = checkCommitIDNotSameAsPrevious(pckgWR, i, checkFailure)
		}
	}
	return checkFailure
}

func checkCommitIDSameAsGiven(pckgWR []aakaas.PackageWithRepository, i int, checkFailure error) error {
	//Ensure for Packages with Status R that CommitID of package = the one from addon.yml, beware of short commitID in addon.yml (and AAKaaS had due to a bug some time also short commid IDs)
	AAKaaSCommitId := pckgWR[i].Package.CommitID
	AddonYAMLCommitId := pckgWR[i].Repo.CommitID

	var commitIdLength int
	//determine shortes commitID length
	commitIdLength = min(len(AAKaaSCommitId), len(AddonYAMLCommitId))

	//shorten both to common length
	AAKaaSCommitId = AAKaaSCommitId[0:commitIdLength]
	AddonYAMLCommitId = AddonYAMLCommitId[0:commitIdLength]

	if AddonYAMLCommitId != AAKaaSCommitId {
		log.Entry().Error("package " + pckgWR[i].Package.PackageName + " was already build but with commit " + pckgWR[i].Package.CommitID + ", not with " + pckgWR[i].Repo.CommitID)
		log.Entry().Error("If you want to build a new package make sure to increase the dotted-version-string in addon.yml - current value: " + pckgWR[i].Package.VersionYAML)
		log.Entry().Error("If you do NOT want to build a new package enter the commitID " + pckgWR[i].Package.CommitID + " for software component " + pckgWR[i].Repo.Name + " in addon.yml")
		checkFailure = errors.New("commit of already released package does not match with addon.yml")
		log.Entry().WithError(checkFailure).Error(" => Check failure: to be corrected in addon.yml prior next execution")
	}
	return checkFailure
}

func checkCommitIDNotSameAsPrevious(pckgWR []aakaas.PackageWithRepository, i int, checkFailure error) error {
	//Check for newly reserved packages which are to be build that CommitID from addon.yml != PreviousCommitID [this will result in an error as no delta can be calculated]
	AAKaaSPreviousCommitId := pckgWR[i].Package.PredecessorCommitID
	AddonYAMLCommitId := pckgWR[i].Repo.CommitID

	var commitIdLength int
	//determine shortes commitID length
	commitIdLength = min(len(AAKaaSPreviousCommitId), len(AddonYAMLCommitId))

	AAKaaSPreviousCommitId = AAKaaSPreviousCommitId[0:commitIdLength]
	AddonYAMLCommitId = AddonYAMLCommitId[0:commitIdLength]

	if AddonYAMLCommitId == AAKaaSPreviousCommitId {
		checkFailure = errors.New("CommitID of package " + pckgWR[i].Package.PackageName + " is the same as the one of the predecessor package. Make sure to change both the dotted-version-string AND the commitID in addon.yml")
		log.Entry().WithError(checkFailure).Error(" => Check failure: to be corrected in addon.yml prior next execution")
	}

	return checkFailure
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
				if pckgWR[i].Package.Status == aakaas.PackageStatusReleased {
					continue
				}
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
				log.Entry().Infof(" => Reservations of package(s) finished successfully")
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
