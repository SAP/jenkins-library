package cmd

import (
	"encoding/json"
	"reflect"
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
	abaputils.Communication
	abapbuild.Publish
	abapbuild.HTTPSendLoader
	getMaxRuntime() time.Duration
	getPollingIntervall() time.Duration
}

type abapEnvironmentBuildUtilsBundle struct {
	*command.Command
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
	utils := newAbapEnvironmentBuildUtils(time.Duration(config.MaxRuntimeInMinutes), time.Duration(config.PollingIntervalInSeconds))
	if err := runAbapEnvironmentBuild(&config, telemetryData, &utils, cpe); err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapEnvironmentBuild(config *abapEnvironmentBuildOptions, telemetryData *telemetry.CustomData, utils *abapEnvironmentBuildUtils, cpe *abapEnvironmentBuildCommonPipelineEnvironment) error {

	conn := new(abapbuild.Connector)
	if err := initConnection(conn, config, utils); err != nil {
		return errors.Wrap(err, "Connector initialization for communication with the ABAP system failed")
	}

	valuesList, err := evaluateAddonDescriptor(config)
	if err != nil {
		//TODO fehlerbehandlung
		return errors.Wrap(err, "??")
	}
	//TODO zwei methoden für die zwei zweige
	if len(valuesList) > 0 {
		for _, values := range valuesList {
			//TODO die values mit den configvalues usw vermischen
			//TODO damit die builds laufen lassen

			finalValues, err := runBuild(conn, config, utils, values)
			if err != nil {
				return errors.Wrap(err, "Error during execution of build framework")
			}
			//TODO mit den finalValues muss ich das irgendiwe anders machen, aufakkumulieren oder so
			cpe.abap.buildValues = finalValues
		}
	} else {
		values, err := generateValues(config)
		if err != nil {
			return errors.Wrap(err, "Generating the values from config failed")
		}
		finalValues, err := runBuild(conn, config, utils, values)
		if err != nil {
			return errors.Wrap(err, "Error during execution of build framework")
		}
		cpe.abap.buildValues = finalValues
	}
	return nil
}

func initConnection(conn *abapbuild.Connector, config *abapEnvironmentBuildOptions, utils *abapEnvironmentBuildUtils) error {
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

	if err := conn.InitBuildFramework(connConfig, *utils, *utils); err != nil {
		return err
	}

	conn.MaxRuntime = (*utils).getMaxRuntime()
	conn.PollingInterval = (*utils).getPollingIntervall()
	return nil
}

// ***********************************Run Build***************************************************************
func runBuild(conn *abapbuild.Connector, config *abapEnvironmentBuildOptions, utils *abapEnvironmentBuildUtils, values abapbuild.Values) (string, error) {
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
		return "", errors.Wrap(err, "Error during the polling for the final state of the build run")
	}

	if err := build.PrintLogs(); err != nil {
		return "", errors.Wrap(err, "Error printing the logs")
	}
	if err := build.EvaluteIfBuildSuccessful(); err != nil {
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
		return errors.Wrap(err, "Error starting the build framework")
	}
	return nil
}

func (b *myBuild) EvaluteIfBuildSuccessful() error {
	if err := b.Build.EvaluteIfBuildSuccessful(b.TreatWarningsAsError); err != nil {
		return errors.Wrap(err, "Build ended without success")
	}
	return nil
}

func (b *myBuild) Download() error {
	if b.DownloadAllResultFiles {
		if err := b.DownloadAllResults(b.SubDirectoryForDownload, b.FilenamePrefixForDownload); err != nil {
			return errors.Wrap(err, "Error during the download of the result files")
		}
	} else {
		if err := b.DownloadResults(b.DownloadResultFilenames, b.SubDirectoryForDownload, b.FilenamePrefixForDownload); err != nil {
			return errors.Wrapf(err, "Error during the download of the result files %s", b.DownloadResultFilenames)
		}
	}
	return nil
}

func (b *myBuild) Publish(utils *abapEnvironmentBuildUtils) error {
	if b.PublishAllDownloadedResultFiles {
		b.PublishAllDownloadedResults("abapEnvironmentBuild", *utils)
	} else {
		if err := b.PublishDownloadedResults("abapEnvironmentBuild", b.PublishResultFilenames, *utils); err != nil {
			return errors.Wrapf(err, "Error during the publish of the result files %s", b.PublishResultFilenames)
		}
	}
	return nil
}

func (b *myBuild) GetFinalValues() (string, error) {
	type cpeValue struct {
		ValueID string `json:"value_id"`
		Value   string `json:"value"`
	}

	if err := b.GetValues(); err != nil {
		return "", errors.Wrapf(err, "Error getting the values from build framework")
	}
	var cpeValues []cpeValue
	byt, err := json.Marshal(&b.Build.Values)
	if err != nil {
		return "", errors.Wrapf(err, "Error converting the values from the build framework")
	}
	if err := json.Unmarshal(byt, &cpeValues); err != nil {
		return "", errors.Wrapf(err, "Error converting the values from the build framework into the structure for the commonPipelineEnvironment")
	}
	jsonBytes, err := json.Marshal(cpeValues)
	if err != nil {
		return "", errors.Wrapf(err, "Error converting the converted values")
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
	if err := vE.appendStringValues(config.CpeValues); err != nil {
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
	if err := json.Unmarshal([]byte(stringValues), &vE.values); err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.Wrapf(err, "Could not convert the values %s from the config", stringValues)
	}

	vE.m = make(map[string]string)
	for _, value := range vE.values {
		if (len(value.ValueID) == 0) || (len(value.Value) == 0) {
			log.SetErrorCategory(log.ErrorConfiguration)
			return errors.Errorf("Values %s from config have not the right format", stringValues)
		}
		_, present := vE.m[value.ValueID]
		if present {
			log.SetErrorCategory(log.ErrorConfiguration)
			return errors.Errorf("Value_id %s is not unique in the config", value.ValueID)
		}
		vE.m[value.ValueID] = value.Value
	}
	return nil
}

func (vE *valuesEvaluator) appendStringValues(stringValues string) error {
	var values []abapbuild.Value
	if len(stringValues) > 0 {
		if err := json.Unmarshal([]byte(stringValues), &values); err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return errors.Wrapf(err, "Could not convert the values %s from the commonPipelineEnvironment", stringValues)
		}
		vE.appendValues(values)
	}
	return nil
}

func (vE *valuesEvaluator) appendValues(values []abapbuild.Value) {
	for i := len(values) - 1; i >= 0; i-- {
		_, present := vE.m[values[i].ValueID]
		if present || (values[i].ValueID == "PHASE") {
			log.Entry().Infof("Value %s already exists in config -> discard this value", values[i])
			values = append(values[:i], values[i+1:]...)
		}
	}
	vE.values = append(vE.values, values...)
}

//**********************************Evaluate AddonDescriptor**************************************************************
type myRepo struct {
	abaputils.Repository
}

type condition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

func (mR *myRepo) checkCondition(config *abapEnvironmentBuildOptions) bool {
	var conditions []condition
	if err := json.Unmarshal([]byte(config.ConditionOnAddonDescriptor), &conditions); err != nil {
		//TODO fehlerbehandlung
	}
	for _, cond := range conditions {
		use, err := mR.amI(cond.Field, cond.Operator, cond.Value)
		if err != nil {
			//TODO fehlerbehandlung
		}
		if !use {
			return use //false
		}
	}
	return true
}

func (mR *myRepo) generateValues() abapbuild.Values {
	var values abapbuild.Values

	fields := reflect.ValueOf(mR.Repository)
	typeOfS := fields.Type()
	for i := 0; i < fields.NumField(); i++ {
		var value abapbuild.Value
		value.ValueID = typeOfS.Field(i).Name
		value.Value = fields.Field(i).String()
		values.Values = append(values.Values, value)
	}
	return values
}

func (mR *myRepo) getField(field string) string {
	//TODO das müsste doch eigentlich auch ein fehler geben? wenn man nach einem feld sucht das es nicht gibt?
	r := reflect.ValueOf(mR)
	f := reflect.Indirect(r).FieldByName(field)
	return string(f.String())
}

func (mR *myRepo) amI(field string, operator string, comp string) (bool, error) {
	operators := OperatorCallback{
		"==": Equal,
		"!=": Unequal,
	}
	name := mR.getField(field)
	if fn, ok := operators[operator]; ok {
		return fn(name, comp), nil
	}
	return false, errors.Errorf("Invalid operator %s", operator)
}

func evaluateAddonDescriptor(config *abapEnvironmentBuildOptions) ([]abapbuild.Values, error) {
	var valuesList []abapbuild.Values
	if len(config.AddonDescriptor) > 0 && config.UseAddonDescriptor {
		addonDescriptor := new(abaputils.AddonDescriptor)
		if err := addonDescriptor.InitFromJSONstring(config.AddonDescriptor); err != nil {
			//TODO fehlerbehandlung
			return valuesList, errors.Wrap(err, "??")
		}
		for _, repo := range addonDescriptor.Repositories {
			myRepo := myRepo{
				Repository: repo,
			}
			use := myRepo.checkCondition(config)
			if use {
				values := myRepo.generateValues()
				valuesList = append(valuesList, values)
			}
		}
	}
	return valuesList, nil
}

type OperatorCallback map[string]func(string, string) bool

func Equal(a, b string) bool {
	return a == b
}

func Unequal(a, b string) bool {
	return a != b
}
