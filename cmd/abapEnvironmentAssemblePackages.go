package cmd

import (
	"encoding/json"
	"path"
	"path/filepath"
	"time"

	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

type buildWithRepository struct {
	build abapbuild.Build
	repo  abaputils.Repository
}

func abapEnvironmentAssemblePackages(config abapEnvironmentAssemblePackagesOptions, telemetryData *telemetry.CustomData, cpe *abapEnvironmentAssemblePackagesCommonPipelineEnvironment) {
	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	var autils = abaputils.AbapUtils{
		Exec: &c,
	}

	client := piperhttp.Client{}
	err := runAbapEnvironmentAssemblePackages(&config, telemetryData, &autils, &client, cpe)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapEnvironmentAssemblePackages(config *abapEnvironmentAssemblePackagesOptions, telemetryData *telemetry.CustomData, com abaputils.Communication, client abapbuild.HTTPSendLoader, cpe *abapEnvironmentAssemblePackagesCommonPipelineEnvironment) error {
	conn := new(abapbuild.Connector)
	var connConfig abapbuild.ConnectorConfiguration
	connConfig.CfAPIEndpoint = config.CfAPIEndpoint
	connConfig.CfOrg = config.CfOrg
	connConfig.CfSpace = config.CfSpace
	connConfig.CfServiceInstance = config.CfServiceInstance
	connConfig.CfServiceKeyName = config.CfServiceKeyName
	connConfig.Host = config.Host
	connConfig.Username = config.Username
	connConfig.Password = config.Password
	connConfig.AddonDescriptor = config.AddonDescriptor
	connConfig.MaxRuntimeInMinutes = config.MaxRuntimeInMinutes

	err := conn.InitBuildFramework(connConfig, com, client)
	if err != nil {
		return errors.Wrap(err, "Connector initialization failed")
	}

	var addonDescriptor abaputils.AddonDescriptor
	err = json.Unmarshal([]byte(config.AddonDescriptor), &addonDescriptor)
	if err != nil {
		return errors.Wrap(err, "Reading AddonDescriptor failed [Make sure abapAddonAssemblyKit...CheckCVs|CheckPV|ReserveNextPackages steps have been run before]")
	}

	maxRuntimeInMinutes := time.Duration(config.MaxRuntimeInMinutes) * time.Minute
	pollIntervalsInMilliseconds := time.Duration(config.PollIntervalsInMilliseconds) * time.Millisecond
	builds, buildsAlreadyReleased, err := executeBuilds(addonDescriptor.Repositories, *conn, maxRuntimeInMinutes, pollIntervalsInMilliseconds)
	if err != nil {
		return errors.Wrap(err, "Starting Builds failed")
	}

	err = checkIfFailedAndPrintLogs(builds)
	if err != nil {
		return errors.Wrap(err, "Check if failed and Printing Logs failed")
	}
	reposBackToCPE, err := downloadSARXML(builds)
	if err != nil {
		return errors.Wrap(err, "Download SAR XML failed")
	}
	// also write the already released packages back to cpe
	for _, b := range buildsAlreadyReleased {
		reposBackToCPE = append(reposBackToCPE, b.repo)
	}
	addonDescriptor.Repositories = reposBackToCPE
	backToCPE, _ := json.Marshal(addonDescriptor)
	cpe.abap.addonDescriptor = string(backToCPE)

	return nil
}

func starting(repos []abaputils.Repository, conn abapbuild.Connector, delayBetweenPostsInSeconds time.Duration) ([]buildWithRepository, []buildWithRepository, error) {
	var builds []buildWithRepository
	var buildsAlreadyReleased []buildWithRepository
	for _, repo := range repos {
		assemblyBuild := abapbuild.Build{
			Connector: conn,
		}
		buildRepo := buildWithRepository{
			build: assemblyBuild,
			repo:  repo,
		}
		if repo.Status == "P" {
			err := buildRepo.start()
			if err != nil {
				return builds, buildsAlreadyReleased, err
			}
			builds = append(builds, buildRepo)
		} else {
			log.Entry().Infof("Packages %s is in status '%s'. No need to run the assembly", repo.PackageName, repo.Status)
			buildsAlreadyReleased = append(buildsAlreadyReleased, buildRepo)
		}

		//as batch events in the ABAP Backend need a little time
		time.Sleep(delayBetweenPostsInSeconds)
	}
	return builds, buildsAlreadyReleased, nil
}

func executeBuilds(repos []abaputils.Repository, conn abapbuild.Connector, maxRuntimeInMinutes time.Duration, pollIntervalsInMilliseconds time.Duration) ([]buildWithRepository, []buildWithRepository, error) {
	var builds []buildWithRepository
	var buildsAlreadyReleased []buildWithRepository
	for _, repo := range repos {
		assemblyBuild := abapbuild.Build{
			Connector: conn,
		}
		buildRepo := buildWithRepository{
			build: assemblyBuild,
			repo:  repo,
		}
		if repo.Status == "P" {
			err := buildRepo.start()
			if err != nil {
				buildRepo.build.RunState = abapbuild.Failed
				log.Entry().Error(err)
				log.Entry().Info("Continueing with other builds (if any)")
			} else {
				err = buildRepo.waitToBeFinished(maxRuntimeInMinutes, pollIntervalsInMilliseconds)
				if err != nil {
					buildRepo.build.RunState = abapbuild.Failed
					log.Entry().Error(err)
					log.Entry().Error("Continueing with other builds (if any) but keep in Mind that even if this build finishes beyond timeout the result is not trustworthy due to possible side effects!")
				}
			}
			builds = append(builds, buildRepo)

		} else {
			log.Entry().Infof("Packages %s is in status '%s'. No need to run the assembly", repo.PackageName, repo.Status)
			buildsAlreadyReleased = append(buildsAlreadyReleased, buildRepo)
		}
	}
	return builds, buildsAlreadyReleased, nil
}

func (br *buildWithRepository) waitToBeFinished(maxRuntimeInMinutes time.Duration, pollIntervalsInMilliseconds time.Duration) error {
	timeout := time.After(maxRuntimeInMinutes)
	ticker := time.Tick(pollIntervalsInMilliseconds)
	for {
		select {
		case <-timeout:
			return errors.Errorf("Timed out: (max Runtime %v reached)", maxRuntimeInMinutes)
		case <-ticker:
			br.build.Get()
			if br.build.IsFinished() {
				log.Entry().Infof("Assembly of %s is not yet finished, check again in %s", br.repo.PackageName, pollIntervalsInMilliseconds)
			} else {
				return nil
			}
		}
	}
}

func (br *buildWithRepository) start() error {
	if br.repo.Name == "" || br.repo.Version == "" || br.repo.SpLevel == "" || br.repo.Namespace == "" || br.repo.PackageType == "" || br.repo.PackageName == "" {
		return errors.New("Parameters missing. Please provide software component name, version, sp-level, namespace, packagetype and packagename")
	}
	valuesInput := abapbuild.Values{
		Values: []abapbuild.Value{
			{
				ValueID: "SWC",
				Value:   br.repo.Name,
			},
			{
				ValueID: "CVERS",
				Value:   br.repo.Name + "." + br.repo.Version + "." + br.repo.SpLevel,
			},
			{
				ValueID: "NAMESPACE",
				Value:   br.repo.Namespace,
			},
			{
				ValueID: "PACKAGE_NAME_" + br.repo.PackageType,
				Value:   br.repo.PackageName,
			},
		},
	}
	if br.repo.PredecessorCommitID != "" {
		valuesInput.Values = append(valuesInput.Values,
			abapbuild.Value{ValueID: "PREVIOUS_DELIVERY_COMMIT",
				Value: br.repo.PredecessorCommitID})
	}
	phase := "BUILD_" + br.repo.PackageType
	log.Entry().Infof("Starting assembly of package %s", br.repo.PackageName)
	return br.build.Start(phase, valuesInput)
}

func checkIfFailedAndPrintLogs(builds []buildWithRepository) error {
	var buildFailed bool = false
	for i := range builds {
		if builds[i].build.RunState == abapbuild.Failed {
			log.Entry().Errorf("Assembly of %s failed", builds[i].repo.PackageName)
			buildFailed = true
		}
		builds[i].build.PrintLogs()
	}
	if buildFailed {
		return errors.New("At least the assembly of one package failed")
	}
	return nil
}

func downloadSARXML(builds []buildWithRepository) ([]abaputils.Repository, error) {
	var reposBackToCPE []abaputils.Repository
	resultName := "SAR_XML"
	envPath := filepath.Join(GeneralConfig.EnvRootPath, "commonPipelineEnvironment", "abap")
	for i, b := range builds {
		resultSARXML, err := b.build.GetResult(resultName)
		if err != nil {
			return reposBackToCPE, err
		}
		sarPackage := resultSARXML.AdditionalInfo
		downloadPath := filepath.Join(envPath, path.Base(sarPackage))
		log.Entry().Infof("Downloading SAR file %s to %s", path.Base(sarPackage), downloadPath)
		err = resultSARXML.Download(downloadPath)
		if err != nil {
			return reposBackToCPE, err
		}
		builds[i].repo.SarXMLFilePath = downloadPath
		reposBackToCPE = append(reposBackToCPE, builds[i].repo)
	}
	return reposBackToCPE, nil
}
