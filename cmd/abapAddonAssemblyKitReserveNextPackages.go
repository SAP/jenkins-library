package cmd

import (
	"encoding/json"
	"fmt"
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

func abapAddonAssemblyKitReserveNextPackages(config abapAddonAssemblyKitReserveNextPackagesOptions, telemetryData *telemetry.CustomData, cpe *abapAddonAssemblyKitReserveNextPackagesCommonPipelineEnvironment) {
	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	client := piperhttp.Client{}
	maxRuntimeInMinutes := time.Duration(5 * time.Minute)
	pollIntervalsInSeconds := time.Duration(30 * time.Second)
	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runAbapAddonAssemblyKitReserveNextPackages(&config, telemetryData, &client, cpe, maxRuntimeInMinutes, pollIntervalsInSeconds)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapAddonAssemblyKitReserveNextPackages(config *abapAddonAssemblyKitReserveNextPackagesOptions, telemetryData *telemetry.CustomData, client piperhttp.Sender,
	cpe *abapAddonAssemblyKitReserveNextPackagesCommonPipelineEnvironment, maxRuntimeInMinutes time.Duration, pollIntervalsInSeconds time.Duration) error {
	conn := new(abapbuild.Connector)
	conn.InitAAKaaS(config.AbapAddonAssemblyKitEndpoint, config.Username, config.Password, client)

	var addonDescriptor abaputils.AddonDescriptor
	json.Unmarshal([]byte(config.AddonDescriptor), &addonDescriptor)

	packagesWithRepos, err := reservePackages(addonDescriptor.Repositories, *conn)
	if err != nil {
		return err
	}

	err = pollReserveNextPackages(packagesWithRepos, maxRuntimeInMinutes, pollIntervalsInSeconds)
	if err != nil {
		return err
	}
	addonDescriptor.Repositories = copyFieldsToRepositories(packagesWithRepos)
	log.Entry().Info("Writing package names, types, status, namespace and predecessorCommitID to CommonPipelineEnvironment")
	backToCPE, _ := json.Marshal(addonDescriptor)
	cpe.abap.addonDescriptor = string(backToCPE)
	return nil
}

func copyFieldsToRepositories(pckgWR []aakaas.PackageWithRepository) []abaputils.Repository {
	var repos []abaputils.Repository
	for i := range pckgWR {
		pckgWR[i].Package.CopyFieldsToRepo(&pckgWR[i].Repo)
		repos = append(repos, pckgWR[i].Repo)
	}
	return repos
}

func pollReserveNextPackages(pckgWR []aakaas.PackageWithRepository, maxRuntimeInMinutes time.Duration, pollIntervalsInSeconds time.Duration) error {
	timeout := time.After(maxRuntimeInMinutes)
	ticker := time.Tick(pollIntervalsInSeconds)
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
					log.Entry().Infof("Reservation of %s is not yet finished, check again in %s", pckgWR[i].Package.PackageName, pollIntervalsInSeconds)
					allFinished = false
				} else {
					switch pckgWR[i].Package.Status {
					case aakaas.PackageStatusLocked:
						return fmt.Errorf("Package %s has invalid status 'locked'", pckgWR[i].Package.PackageName)
					case aakaas.PackageStatusCreationTriggered:
						log.Entry().Infof("Reservation of %s is still running with status 'creation triggered', check again in %s", pckgWR[i].Package.PackageName, pollIntervalsInSeconds)
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
