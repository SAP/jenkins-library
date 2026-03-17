package cmd

import (
	"fmt"
	"path"
	"path/filepath"
	"time"

	"errors"

	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
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

	telemetryData.BuildTool = "ABAP Build Framework"

	if err := runAbapEnvironmentAssemblePackages(&config, &autils, &utils, &client, cpe); err != nil {
		telemetryData.ErrorCode = err.Error()
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapEnvironmentAssemblePackages(config *abapEnvironmentAssemblePackagesOptions, com abaputils.Communication, utils piperutils.FileUtils, client abapbuild.HTTPSendLoader, cpe *abapEnvironmentAssemblePackagesCommonPipelineEnvironment) error {
	log.Entry().Info("╔═════════════════════════════════╗")
	log.Entry().Info("║ abapEnvironmentAssemblePackages ║")
	log.Entry().Info("╚═════════════════════════════════╝")

	addonDescriptor := new(abaputils.AddonDescriptor)
	if err := addonDescriptor.InitFromJSONstring(config.AddonDescriptor); err != nil {
		return fmt.Errorf("Reading AddonDescriptor failed [Make sure abapAddonAssemblyKit...CheckCVs|CheckPV|ReserveNextPackages steps have been run before]: %w", err)
	}

	builds, assembleError := runAssemblePackages(config, com, utils, client, addonDescriptor)
	if assembleError != nil && builds != nil {
		addonDescriptor.ErrorText = assembleError.Error()
		log.Entry().Info("---------------------------------")
		log.Entry().Error("During the Assembly errors occured on following levels:")
		for _, build := range builds {
			var errorText string
			if build.repo.ErrorText == "" {
				errorText = "<No Error>"
			} else {
				errorText = build.repo.ErrorText
			}
			log.Entry().Errorf("Software Component %s: %s", build.repo.Name, errorText)
		}
		log.Entry().Errorf("Product %s: %s", addonDescriptor.AddonProduct, addonDescriptor.ErrorText)
	}

	var reposBackToCPE []abaputils.Repository
	for _, b := range builds {
		reposBackToCPE = append(reposBackToCPE, b.repo)
	}
	addonDescriptor.SetRepositories(reposBackToCPE)
	cpe.abap.addonDescriptor = addonDescriptor.AsJSONstring()

	return assembleError
}

func runAssemblePackages(config *abapEnvironmentAssemblePackagesOptions, com abaputils.Communication, utils piperutils.FileUtils, client abapbuild.HTTPSendLoader, addonDescriptor *abaputils.AddonDescriptor) ([]buildWithRepository, error) {
	connBuild := new(abapbuild.Connector)
	if errConBuild := initAssemblePackagesConnection(connBuild, config, com, client); errConBuild != nil {
		return nil, errConBuild
	}

	builds, err := executeBuilds(addonDescriptor, *connBuild, time.Duration(config.MaxRuntimeInMinutes)*time.Minute, time.Duration(config.PollIntervalsInMilliseconds)*time.Millisecond, config.AlternativePhaseName)
	if err != nil {
		return builds, fmt.Errorf("Starting Builds for Repositories with reserved AAKaaS packages failed: %w", err)
	}

	if err := checkIfFailedAndPrintLogs(builds); err != nil {
		return builds, fmt.Errorf("Checking for failed Builds and Printing Build Logs failed: %w", err)
	}

	if _, err := downloadResultToFile(builds, "SAR_XML", false); err != nil {
		return builds, fmt.Errorf("Download of Build Artifact SAR_XML failed: %w", err)
	}

	var filesToPublish []piperutils.Path
	filesToPublish, err = downloadResultToFile(builds, "DELIVERY_LOGS.ZIP", true)
	if err != nil {
		return builds, fmt.Errorf("Download of Build Artifact DELIVERY_LOGS.ZIP failed: %w", err)
	}

	log.Entry().Infof("Publishing %v files", len(filesToPublish))
	piperutils.PersistReportsAndLinks("abapEnvironmentAssemblePackages", "", utils, filesToPublish, nil)

	return builds, nil
}

func executeBuilds(addonDescriptor *abaputils.AddonDescriptor, conn abapbuild.Connector, maxRuntimeInMinutes time.Duration, pollInterval time.Duration, altenativePhaseName string) ([]buildWithRepository, error) {
	var builds []buildWithRepository

	for _, repo := range addonDescriptor.Repositories {

		buildRepo := buildWithRepository{
			build: abapbuild.Build{
				Connector: conn,
			},
			repo: repo,
		}

		if repo.Status == "P" {
			buildRepo.repo.InBuildScope = true
			err := buildRepo.start(addonDescriptor, altenativePhaseName)
			if err != nil {
				buildRepo.build.RunState = abapbuild.Failed
				buildRepo.repo.ErrorText = fmt.Sprint(err)
				log.Entry().Error(err)
				log.Entry().Info("Continueing with other builds (if any)")
			} else {
				err = buildRepo.waitToBeFinished(maxRuntimeInMinutes, pollInterval)
				if err != nil {
					buildRepo.build.RunState = abapbuild.Failed
					log.Entry().Error(err)
					buildRepo.repo.ErrorText = fmt.Sprint(err)
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
			return fmt.Errorf("Timed out: (max Runtime %v reached)", maxRuntimeInMinutes)
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

func (br *buildWithRepository) start(addonDescriptor *abaputils.AddonDescriptor, altenativePhaseName string) error {
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
				ValueID: "SEMANTIC_VERSION",
				Value:   br.repo.VersionYAML,
			},
			{
				ValueID: "PACKAGE_TYPE",
				Value:   br.repo.PackageType,
			},
			{
				ValueID: "PACKAGE_NAME_" + br.repo.PackageType,
				Value:   br.repo.PackageName,
			},
			{
				ValueID: "addonDescriptor",
				Value:   addonDescriptor.AsReducedJson(),
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
		valuesInput.Values = append(valuesInput.Values,
			abapbuild.Value{ValueID: "CURRENT_DELIVERY_COMMIT",
				Value: br.repo.CommitID})
	}
	if br.repo.Tag != "" {
		valuesInput.Values = append(valuesInput.Values, abapbuild.Value{ValueID: "CURRENT_DELIVERY_TAG", Value: br.repo.Tag})
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

	var phase string
	if altenativePhaseName != "" {
		phase = altenativePhaseName
	} else {
		phase = "BUILD_" + br.repo.PackageType
	}

	log.Entry().Infof("Starting assembly of package %s as %s", br.repo.PackageName, phase)
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
		if builds[i].build.ResultState == abapbuild.Erroneous {
			log.Entry().Errorf("Assembly of %s revealed errors", builds[i].repo.PackageName)
			buildFailed = true
		}
		if builds[i].build.BuildID != "" {
			if err := builds[i].build.PrintLogs(); err != nil {
				return err
			}
			cause, err := builds[i].build.DetermineFailureCause()
			if err != nil {
				return err
			} else if cause != "" {
				builds[i].repo.ErrorText = cause
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
		return fmt.Errorf("Connector initialization for communication with the ABAP system failed: %w", err)
	}

	return nil
}
