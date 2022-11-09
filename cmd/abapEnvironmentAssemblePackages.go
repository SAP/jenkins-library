package cmd

import (
	"path"
	"path/filepath"
	"time"

	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
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
	utils := piperutils.Files{}
	err := runAbapEnvironmentAssemblePackages(&config, telemetryData, &autils, &utils, &client, cpe)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapEnvironmentAssemblePackages(config *abapEnvironmentAssemblePackagesOptions, telemetryData *telemetry.CustomData, com abaputils.Communication, utils piperutils.FileUtils, client abapbuild.HTTPSendLoader, cpe *abapEnvironmentAssemblePackagesCommonPipelineEnvironment) error {
	connBuild := new(abapbuild.Connector)
	if errConBuild := initAssemblePackagesConnection(connBuild, config, com, client); errConBuild != nil {
		return errConBuild
	}

	addonDescriptor := new(abaputils.AddonDescriptor)
	if err := addonDescriptor.InitFromJSONstring(config.AddonDescriptor); err != nil {
		return errors.Wrap(err, "Reading AddonDescriptor failed [Make sure abapAddonAssemblyKit...CheckCVs|CheckPV|ReserveNextPackages steps have been run before]")
	}

	builds, err := executeBuilds(addonDescriptor.Repositories, *connBuild, time.Duration(config.MaxRuntimeInMinutes)*time.Minute, time.Duration(config.PollIntervalsInMilliseconds)*time.Millisecond)
	if err != nil {
		return errors.Wrap(err, "Starting Builds for Repositories with reserved AAKaaS packages failed")
	}

	err = checkIfFailedAndPrintLogs(builds)
	if err != nil {
		return errors.Wrap(err, "Checking for failed Builds and Printing Build Logs failed")
	}

	_, err = downloadResultToFile(builds, "SAR_XML", false) //File is present in ABAP build system and uploaded to AAKaaS, no need to fill up jenkins with it
	if err != nil {
		return errors.Wrap(err, "Download of Build Artifact SAR_XML failed")
	}

	var filesToPublish []piperutils.Path
	filesToPublish, err = downloadResultToFile(builds, "DELIVERY_LOGS.ZIP", true)
	if err != nil {
		return errors.Wrap(err, "Download of Build Artifact DELIVERY_LOGS.ZIP failed")
	}

	log.Entry().Infof("Publishing %v files", len(filesToPublish))
	piperutils.PersistReportsAndLinks("abapEnvironmentAssemblePackages", "", utils, filesToPublish, nil)

	var reposBackToCPE []abaputils.Repository
	for _, b := range builds {
		reposBackToCPE = append(reposBackToCPE, b.repo)
	}
	addonDescriptor.SetRepositories(reposBackToCPE)
	cpe.abap.addonDescriptor = addonDescriptor.AsJSONstring()

	return nil
}

func executeBuilds(repos []abaputils.Repository, conn abapbuild.Connector, maxRuntimeInMinutes time.Duration, pollInterval time.Duration) ([]buildWithRepository, error) {
	var builds []buildWithRepository

	for _, repo := range repos {

		buildRepo := buildWithRepository{
			build: abapbuild.Build{
				Connector: conn,
			},
			repo: repo,
		}

		if repo.Status == "P" {
			buildRepo.repo.InBuildScope = true
			err := buildRepo.start()
			if err != nil {
				buildRepo.build.RunState = abapbuild.Failed
				log.Entry().Error(err)
				log.Entry().Info("Continueing with other builds (if any)")
			} else {
				err = buildRepo.waitToBeFinished(maxRuntimeInMinutes, pollInterval)
				if err != nil {
					buildRepo.build.RunState = abapbuild.Failed
					log.Entry().Error(err)
					log.Entry().Error("Continuing with other builds (if any) but keep in Mind that even if this build finishes beyond timeout the result is not trustworthy due to possible side effects!")
				}
			}
		} else {
			log.Entry().Infof("Packages %s is in status '%s'. No need to run the assembly", repo.PackageName, repo.Status)
		}

		builds = append(builds, buildRepo)
	}
	return builds, nil
}

func (br *buildWithRepository) waitToBeFinished(maxRuntimeInMinutes time.Duration, pollInterval time.Duration) error {
	timeout := time.After(maxRuntimeInMinutes)
	ticker := time.Tick(pollInterval)
	for {
		select {
		case <-timeout:
			return errors.Errorf("Timed out: (max Runtime %v reached)", maxRuntimeInMinutes)
		case <-ticker:
			if err := br.build.Get(); err != nil {
				return err
			}
			if !br.build.IsFinished() {
				log.Entry().Infof("Assembly of %s is not yet finished, check again in %s", br.repo.PackageName, pollInterval)
			} else {
				return nil
			}
		}
	}
}

func (br *buildWithRepository) start() error {
	if br.repo.Name == "" || br.repo.Version == "" || br.repo.SpLevel == "" || br.repo.PackageType == "" || br.repo.PackageName == "" {
		return errors.New("Parameters missing. Please provide software component name, version, sp-level, packagetype and packagename")
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
				ValueID: "PACKAGE_TYPE",
				Value:   br.repo.PackageType,
			},
			{
				ValueID: "PACKAGE_NAME_" + br.repo.PackageType,
				Value:   br.repo.PackageName,
			},
		},
	}
	if br.repo.Namespace != "" {
		valuesInput.Values = append(valuesInput.Values,
			abapbuild.Value{ValueID: "NAMESPACE",
				Value: br.repo.Namespace})
	}
	if br.repo.UseClassicCTS {
		valuesInput.Values = append(valuesInput.Values,
			abapbuild.Value{ValueID: "useClassicCTS",
				Value: "true"})
	}
	if br.repo.PredecessorCommitID != "" {
		valuesInput.Values = append(valuesInput.Values,
			abapbuild.Value{ValueID: "PREVIOUS_DELIVERY_COMMIT",
				Value: br.repo.PredecessorCommitID})
	}
	if br.repo.CommitID != "" {
		// old value to be used until 2302 [can be deleted latest with 2308]
		valuesInput.Values = append(valuesInput.Values,
			abapbuild.Value{ValueID: "ACTUAL_DELIVERY_COMMIT",
				Value: br.repo.CommitID})
		// new value used as of 2302
		valuesInput.Values = append(valuesInput.Values,
			abapbuild.Value{ValueID: "CURRENT_DELIVERY_COMMIT",
				Value: br.repo.CommitID})
	}
	if len(br.repo.Languages) > 0 {
		valuesInput.Values = append(valuesInput.Values,
			abapbuild.Value{ValueID: "SSDC_EXPORT_LANGUAGE_VECTOR",
				Value: br.repo.GetAakAasLanguageVector()})
	}
	if br.repo.AdditionalPiecelist != "" {
		valuesInput.Values = append(valuesInput.Values,
			abapbuild.Value{ValueID: "ADDITIONAL_PIECELIST",
				Value: br.repo.AdditionalPiecelist})
	}
	phase := "BUILD_" + br.repo.PackageType
	log.Entry().Infof("Starting assembly of package %s", br.repo.PackageName)
	return br.build.Start(phase, valuesInput)
}

func downloadResultToFile(builds []buildWithRepository, resultName string, publish bool) ([]piperutils.Path, error) {
	envPath := filepath.Join(GeneralConfig.EnvRootPath, "abapBuild")
	var filesToPublish []piperutils.Path

	for i, b := range builds {
		if b.repo.Status != "P" {
			continue
		}
		buildResult, err := b.build.GetResult(resultName)
		if err != nil {
			return filesToPublish, err
		}
		var fileName string
		if len(buildResult.AdditionalInfo) <= 255 {
			fileName = buildResult.AdditionalInfo
		} else {
			fileName = buildResult.Name
		}
		downloadPath := filepath.Join(envPath, path.Base(fileName))
		log.Entry().Infof("Downloading %s file %s to %s", resultName, path.Base(fileName), downloadPath)
		err = buildResult.Download(downloadPath)
		if err != nil {
			return filesToPublish, err
		}
		if resultName == "SAR_XML" {
			builds[i].repo.SarXMLFilePath = downloadPath
		}
		if publish {
			log.Entry().Infof("Add %s to be published", resultName)
			filesToPublish = append(filesToPublish, piperutils.Path{Target: downloadPath, Name: resultName, Mandatory: true})
		}
	}
	return filesToPublish, nil
}

func checkIfFailedAndPrintLogs(builds []buildWithRepository) error {
	var buildFailed bool = false
	for i := range builds {
		if builds[i].build.RunState == abapbuild.Failed {
			log.Entry().Errorf("Assembly of %s failed", builds[i].repo.PackageName)
			buildFailed = true
		}
		if builds[i].build.BuildID != "" {
			if err := builds[i].build.PrintLogs(); err != nil {
				return err
			}
		}
	}
	if buildFailed {
		return errors.New("At least the assembly of one package failed")
	}
	return nil
}

func initAssemblePackagesConnection(conn *abapbuild.Connector, config *abapEnvironmentAssemblePackagesOptions, com abaputils.Communication, client abapbuild.HTTPSendLoader) error {
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
	connConfig.CertificateNames = config.CertificateNames

	err := conn.InitBuildFramework(connConfig, com, client)
	if err != nil {
		return errors.Wrap(err, "Connector initialization for communication with the ABAP system failed")
	}

	return nil
}
