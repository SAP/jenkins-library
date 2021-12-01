package cmd

import (
	"encoding/json"
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

type abapEnvironmentBuildUtils interface {
	command.ExecRunner

	FileExists(filename string) (bool, error)
	abaputils.Communication
	abapbuild.Publish
	abapbuild.HTTPSendLoader
	getMaxRuntime() time.Duration
	getPollingIntervall() time.Duration
}

type abapEnvironmentBuildUtilsBundle struct {
	*command.Command
	*piperutils.Files
	*piperhttp.Client
	*abaputils.AbapUtils
	maxRuntime       time.Duration
	pollingIntervall time.Duration
}

func (aEBUB *abapEnvironmentBuildUtilsBundle) getMaxRuntime() time.Duration {
	return aEBUB.maxRuntime
}

func (aEBUB *abapEnvironmentBuildUtilsBundle) getPollingIntervall() time.Duration {
	return aEBUB.pollingIntervall
}

func (aEBUB *abapEnvironmentBuildUtilsBundle) PersistReportsAndLinks(stepName, workspace string, reports, links []piperutils.Path) {
	abapbuild.PersistReportsAndLinks(stepName, workspace, reports, links)
}

func newAbapEnvironmentBuildUtils(maxRuntime time.Duration, pollingIntervall time.Duration) abapEnvironmentBuildUtils {
	utils := abapEnvironmentBuildUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
		Client:  &piperhttp.Client{},
		AbapUtils: &abaputils.AbapUtils{
			Exec: &command.Command{},
		},
		maxRuntime:       maxRuntime * time.Minute,
		pollingIntervall: pollingIntervall * time.Second,
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func abapEnvironmentBuild(config abapEnvironmentBuildOptions, telemetryData *telemetry.CustomData, cpe *abapEnvironmentBuildCommonPipelineEnvironment) {
	utils := newAbapEnvironmentBuildUtils(time.Duration(config.MaxRuntimeInMinutes), time.Duration(config.PollingIntervallInSeconds))
	err := runAbapEnvironmentBuild(&config, telemetryData, utils, cpe)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapEnvironmentBuild(config *abapEnvironmentBuildOptions, telemetryData *telemetry.CustomData, utils abapEnvironmentBuildUtils, cpe *abapEnvironmentBuildCommonPipelineEnvironment) error {
	conn := new(abapbuild.Connector)

	// TODO wrappe die fehler
	if err := initConnection(conn, config, utils); err != nil {
		return err
	}
	values, err := generateValues(config)
	if err != nil {
		return errors.Wrap(err, "Generating the values from config failed")
	}

	//TODO delete
	log.Entry().Infof("Values used %s", values.Values)

	finalValues, err := runBuild(conn, config, utils, values)
	if err != nil {
		return err
	}
	cpe.build.values = finalValues
	return nil
}

func initConnection(conn *abapbuild.Connector, config *abapEnvironmentBuildOptions, utils abapEnvironmentBuildUtils) error {
	var connConfig abapbuild.ConnectorConfiguration
	connConfig.CfAPIEndpoint = config.CfAPIEndpoint
	connConfig.CfOrg = config.CfOrg
	connConfig.CfSpace = config.CfSpace
	connConfig.CfServiceInstance = config.CfServiceInstance
	connConfig.CfServiceKeyName = config.CfServiceKeyName
	connConfig.Host = config.Host
	connConfig.Username = config.Username
	connConfig.Password = config.Password
	connConfig.MaxRuntimeInMinutes = config.MaxRuntimeInMinutes
	connConfig.CertificateNames = config.CertificateNames

	err := conn.InitBuildFramework(connConfig, utils, utils)
	if err != nil {
		return errors.Wrap(err, "Connector initialization for communication with the ABAP system failed")
	}
	conn.MaxRuntime = utils.getMaxRuntime()
	conn.PollingInterval = utils.getPollingIntervall()
	log.Entry().Infof("MaxRuntime %s", conn.MaxRuntime)
	log.Entry().Infof("polling intervall %s", conn.PollingInterval)
	return nil
}

// ***********************************Run Build***************************************************************
func runBuild(conn *abapbuild.Connector, config *abapEnvironmentBuildOptions, utils abapEnvironmentBuildUtils, values abapbuild.Values) (string, error) {
	build := myBuild{
		Build: abapbuild.Build{
			Connector: *conn,
		},
		abapEnvironmentBuildOptions: config,
	}
	if err := build.Start(values); err != nil {
		return "", err
	}

	if err := build.Poll(); err != nil {
		return "", err
	}
	if err := build.PrintLogs(); err != nil {
		return "", err
	}
	if err := build.EndedWithError(); err != nil {
		return "", err
	}
	if err := build.Download(); err != nil {
		return "", err
	}
	if err := build.Publish(utils); err != nil {
		return "", err
	}

	finalValues, err := build.GetFinalValues()
	if err != nil {
		return "", err
	}
	return finalValues, nil
}

type myBuild struct {
	abapbuild.Build
	*abapEnvironmentBuildOptions
}

func (b *myBuild) Start(values abapbuild.Values) error {
	if err := b.Build.Start(b.abapEnvironmentBuildOptions.Phase, values); err != nil {
		return errors.Wrap(err, "Error during the execution of the build")
	}
	return nil
}

func (b *myBuild) EndedWithError() error {
	if err := b.Build.EndedWithError(b.TreatWarningsAsError); err != nil {
		return err
	}
	return nil
}

func (b *myBuild) Download() error {
	if b.DownloadAllResultFiles {
		log.Entry().Infof("Downloading all available result files")
		if err := b.DownloadAllResults(b.SubDirectoryForDownload, b.FilenamePrefixForDownload); err != nil {
			return errors.Wrap(err, "Error during the download of the result files")
		}
	} else {
		if err := b.DownloadResults(b.DownloadResultFilenames, b.SubDirectoryForDownload, b.FilenamePrefixForDownload); err != nil {
			return err
			//TODO error
		}
	}
	return nil
}

func (b *myBuild) Publish(utils abapEnvironmentBuildUtils) error {
	if b.PublishAllDownloadedResultFiles {
		b.PublishAllDownloadedResults("abapEnvironmentBuild", utils)
	} else {
		if err := b.PublishDownloadedResults("abapEnvironmentBuild", b.PublishResultFilenames, utils); err != nil {
			return err
		}
	}
	return nil
}

func (b *myBuild) GetFinalValues() (string, error) {
	type cpeValue struct {
		ValueID string `json:"value_id"`
		Value   string `json:"value"`
	}

	b.GetValues()
	var cpeValues []cpeValue
	byt, err := json.Marshal(&b.Build.Values)
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(byt, &cpeValues); err != nil {
		return "", err
	}
	jsonBytes, err := json.Marshal(cpeValues)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}

// **********************************Generate Values**************************************************************
func generateValues(config *abapEnvironmentBuildOptions) (abapbuild.Values, error) {
	var values abapbuild.Values
	vE := valuesEvaluator{}
	if err := vE.initialize(config.Values); err != nil {
		return values, err
	}
	if err := vE.appendValues(config.CpeValues); err != nil {
		return values, err
	}
	values.Values = vE.values
	return values, nil
}

type valuesEvaluator struct {
	values []abapbuild.Value
	m      map[string]string
}

func (vE *valuesEvaluator) initialize(stringValues string) error {
	log.Entry().Infof("Input values from config %s", stringValues)
	if err := json.Unmarshal([]byte(stringValues), &vE.values); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.Wrapf(err, "Could not convert the values %s from the config", stringValues)
	}

	vE.m = make(map[string]string)
	for _, value := range vE.values {
		_, present := vE.m[value.ValueID]
		if present {
			log.SetErrorCategory(log.ErrorConfiguration)
			return errors.Errorf("Value_id %s is not unique in the config", value.ValueID)
		}
		vE.m[value.ValueID] = value.Value
	}
	return nil
}

func (vE *valuesEvaluator) appendValues(stringValues string) error {
	var values []abapbuild.Value
	log.Entry().Infof("Input values from the commonPipelineEnvironment %s", stringValues)
	if len(stringValues) > 0 {
		if err := json.Unmarshal([]byte(stringValues), &values); err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return errors.Wrapf(err, "Could not convert the values %s from the commonPipelineEnvironment", stringValues)
		}
		for i := len(values) - 1; i >= 0; i-- {
			_, present := vE.m[values[i].ValueID]
			if present || (values[i].ValueID == "PHASE") {
				log.Entry().Infof("Value %s already exists in config -> discard this value", values[i])
				values = append(values[:i], values[i+1:]...)
			}
		}
		vE.values = append(vE.values, values...)
	}
	return nil
}
