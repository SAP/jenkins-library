package cmd

import (
	"encoding/json"
	"fmt"
	"time"

	"errors"

	abapbuild "github.com/SAP/jenkins-library/pkg/abap/build"
	"github.com/SAP/jenkins-library/pkg/abaputils"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
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

	telemetryData.BuildTool = "ABAP Build Framework"

	client := piperhttp.Client{}
	err := runAbapEnvironmentAssembleConfirm(&config, &autils, &client, cpe)
	if err != nil {
		telemetryData.ErrorCode = err.Error()
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapEnvironmentAssembleConfirm(config *abapEnvironmentAssembleConfirmOptions, com abaputils.Communication, client abapbuild.HTTPSendLoader, cpe *abapEnvironmentAssembleConfirmCommonPipelineEnvironment) error {
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
	connConfig.CertificateNames = config.CertificateNames

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
	var releasePackagesFailed error = nil
	for _, repo := range repos {
		assemblyBuild := abapbuild.Build{
			Connector: conn,
		}
		buildRepo := buildWithRepository{
			build: assemblyBuild,
			repo:  repo,
		}
		if repo.InBuildScope && repo.Status == "R" {
			err := buildRepo.startConfirm()
			if err != nil {
				return confirmedBuilds, err
			}
			confirmedBuilds = append(confirmedBuilds, buildRepo)
		} else if repo.InBuildScope && repo.Status != "R" {
			errormessage := "Release of package '" + repo.PackageName + "' must have failed as still in status: '" + repo.Status + "'"
			if releasePackagesFailed == nil {
				releasePackagesFailed = errors.New(errormessage)
			} else {
				releasePackagesFailed = fmt.Errorf("%s: %w", errormessage, releasePackagesFailed)
			}
		} else {
			log.Entry().Infof("Packages %s was not assembled in this pipeline run, thus no need to confirm", repo.PackageName)
		}

		//as batch events in the ABAP Backend need a little time
		time.Sleep(delayBetweenPosts)
	}
	return confirmedBuilds, releasePackagesFailed
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
				if err := builds[i].build.Get(); err != nil {
					return err
				}
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
	if b.repo.Name == "" || b.repo.PackageName == "" {
		return errors.New("Parameters missing. Please provide software component name, namespace and packagename")
	}
	valuesInput := abapbuild.Values{
		Values: []abapbuild.Value{
			{
				ValueID: "SWC",
				Value:   b.repo.Name,
			},
		},
	}
	if b.repo.Namespace != "" {
		// Steampunk Use Case, Namespace provided by AAKaaS
		valuesInput.Values = append(valuesInput.Values,
			abapbuild.Value{ValueID: "SSDC-delta",
				Value: b.repo.Namespace + b.repo.PackageName})
	} else {
		// Traditional SWCs, Namespace to be provided in assembly system via build script
		valuesInput.Values = append(valuesInput.Values,
			abapbuild.Value{ValueID: "PACKAGE_TYPE",
				Value: b.repo.PackageType})
		valuesInput.Values = append(valuesInput.Values,
			abapbuild.Value{ValueID: "PACKAGE_NAME_" + b.repo.PackageType,
				Value: b.repo.PackageName})
	}

	phase := "BUILD_CONFIRM"
	log.Entry().Infof("Starting confirmation of package %s", b.repo.PackageName)
	return b.build.Start(phase, valuesInput)
}
