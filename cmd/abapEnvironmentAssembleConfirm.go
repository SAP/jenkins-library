package cmd

import (
	"encoding/json"
	"time"

	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func abapEnvironmentAssembleConfirm(config abapEnvironmentAssembleConfirmOptions, telemetryData *telemetry.CustomData, cpe *abapEnvironmentAssembleConfirmCommonPipelineEnvironment) {
	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	var autils = abaputils.AbapUtils{
		Exec: &c,
	}

	client := piperhttp.Client{}
	err := runAbapEnvironmentAssembleConfirm(&config, telemetryData, &autils, &client, cpe)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapEnvironmentAssembleConfirm(config *abapEnvironmentAssembleConfirmOptions, telemetryData *telemetry.CustomData, com abaputils.Communication, client abapbuild.HTTPSendLoader, cpe *abapEnvironmentAssembleConfirmCommonPipelineEnvironment) error {
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
		return err
	}
	var addonDescriptor abaputils.AddonDescriptor
	err = json.Unmarshal([]byte(config.AddonDescriptor), &addonDescriptor)
	if err != nil {
		return err
	}
	delayBetweenPosts := time.Duration(3 * time.Second)
	builds, err := startingConfirm(addonDescriptor.Repositories, *conn, delayBetweenPosts)
	if err != nil {
		return err
	}
	maxRuntimeInMinutes := time.Duration(config.MaxRuntimeInMinutes) * time.Minute
	pollInterval := time.Duration(60 * time.Second)
	err = polling(builds, maxRuntimeInMinutes, pollInterval)
	if err != nil {
		return err
	}
	err = checkIfFailedAndPrintLogs(builds)
	if err != nil {
		return err
	}
	return nil
}

func startingConfirm(repos []abaputils.Repository, conn abapbuild.Connector, delayBetweenPosts time.Duration) ([]buildWithRepository, error) {
	var confirmedBuilds []buildWithRepository
	for _, repo := range repos {
		assemblyBuild := abapbuild.Build{
			Connector: conn,
		}
		buildRepo := buildWithRepository{
			build: assemblyBuild,
			repo:  repo,
		}
		if repo.InBuildScope {
			err := buildRepo.startConfirm()
			if err != nil {
				return confirmedBuilds, err
			}
			confirmedBuilds = append(confirmedBuilds, buildRepo)
		} else {
			log.Entry().Infof("Packages %s was not assembled in this pipeline run, thus no need to confirm", repo.PackageName)
		}

		//as batch events in the ABAP Backend need a little time
		time.Sleep(delayBetweenPosts)
	}
	return confirmedBuilds, nil
}

func polling(builds []buildWithRepository, maxRuntimeInMinutes time.Duration, pollInterval time.Duration) error {
	timeout := time.After(maxRuntimeInMinutes)
	ticker := time.Tick(pollInterval)
	for {
		select {
		case <-timeout:
			return errors.New("Timed out")
		case <-ticker:
			var allFinished bool = true
			for i := range builds {
				builds[i].build.Get()
				if !builds[i].build.IsFinished() {
					log.Entry().Infof("Assembly of %s is not yet finished, check again in %s", builds[i].repo.PackageName, pollInterval)
					allFinished = false
				}
			}
			if allFinished {
				return nil
			}
		}
	}
}

func (b *buildWithRepository) startConfirm() error {
	if b.repo.Name == "" || b.repo.Namespace == "" || b.repo.PackageName == "" {
		return errors.New("Parameters missing. Please provide software component name, namespace and packagename")
	}
	valuesInput := abapbuild.Values{
		Values: []abapbuild.Value{
			{
				ValueID: "SWC",
				Value:   b.repo.Name,
			},
			{
				ValueID: "SSDC-delta",
				Value:   b.repo.Namespace + b.repo.PackageName,
			},
		},
	}
	phase := "BUILD_CONFIRM"
	log.Entry().Infof("Starting confirmation of package %s", b.repo.PackageName)
	return b.build.Start(phase, valuesInput)
}
