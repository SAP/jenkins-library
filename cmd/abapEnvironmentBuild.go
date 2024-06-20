package cmd

import (
	"encoding/json"
	"net/url"
	"reflect"
	"strings"
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
	abapbuild.HTTPSendLoader
	piperutils.FileUtils
	getMaxRuntime() time.Duration
	getPollingInterval() time.Duration
	publish()
}

type abapEnvironmentBuildUtilsBundle struct {
	*command.Command
	*piperhttp.Client
	*abaputils.AbapUtils
	*piperutils.Files
	maxRuntime      time.Duration
	pollingInterval time.Duration
	storePublish    publish
}

type publish struct {
	stepName  string
	workspace string
	reports   []piperutils.Path
	links     []piperutils.Path
}

func (p *publish) publish(utils piperutils.FileUtils) {
	if p.stepName != "" {
		piperutils.PersistReportsAndLinks(p.stepName, p.workspace, utils, p.reports, p.links)
	}
}

func (aEBUB *abapEnvironmentBuildUtilsBundle) publish() {
	aEBUB.storePublish.publish(aEBUB)
}

func (aEBUB *abapEnvironmentBuildUtilsBundle) getMaxRuntime() time.Duration {
	return aEBUB.maxRuntime
}

func (aEBUB *abapEnvironmentBuildUtilsBundle) getPollingInterval() time.Duration {
	return aEBUB.pollingInterval
}

func (aEBUB *abapEnvironmentBuildUtilsBundle) PersistReportsAndLinks(stepName, workspace string, reports, links []piperutils.Path) {
	// abapbuild.PersistReportsAndLinks(stepName, workspace, reports, links)
	if aEBUB.storePublish.stepName == "" {
		aEBUB.storePublish.stepName = stepName
		aEBUB.storePublish.workspace = workspace
		aEBUB.storePublish.reports = reports
		aEBUB.storePublish.links = links
	} else {
		aEBUB.storePublish.reports = append(aEBUB.storePublish.reports, reports...)
		aEBUB.storePublish.links = append(aEBUB.storePublish.reports, links...)
	}
}

func newAbapEnvironmentBuildUtils(maxRuntime time.Duration, pollingInterval time.Duration) abapEnvironmentBuildUtils {
	utils := abapEnvironmentBuildUtilsBundle{
		Command: &command.Command{},
		Client:  &piperhttp.Client{},
		AbapUtils: &abaputils.AbapUtils{
			Exec: &command.Command{},
		},
		maxRuntime:      maxRuntime * time.Minute,
		pollingInterval: pollingInterval * time.Second,
		storePublish:    publish{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func abapEnvironmentBuild(config abapEnvironmentBuildOptions, telemetryData *telemetry.CustomData, cpe *abapEnvironmentBuildCommonPipelineEnvironment) {
	utils := newAbapEnvironmentBuildUtils(time.Duration(config.MaxRuntimeInMinutes), time.Duration(config.PollingIntervalInSeconds))
	telemetryData.BuildTool = "ABAP Build Framework"

	if err := runAbapEnvironmentBuild(&config, utils, cpe); err != nil {
		telemetryData.ErrorCode = err.Error()
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapEnvironmentBuild(config *abapEnvironmentBuildOptions, utils abapEnvironmentBuildUtils, cpe *abapEnvironmentBuildCommonPipelineEnvironment) error {

	log.Entry().Info("╔════════════════════════════════╗")
	log.Entry().Info("║ abapEnvironmentBuild           ║")
	log.Entry().Info("╠════════════════════════════════╣")
	log.Entry().Infof("║ %-30v ║", config.Phase)
	log.Entry().Info("╙────────────────────────────────╜")

	conn := new(abapbuild.Connector)
	if err := initConnection(conn, config, utils); err != nil {
		return errors.Wrap(err, "Connector initialization for communication with the ABAP system failed")
	}

	valuesList, err := evaluateAddonDescriptor(config)
	if err != nil {
		return errors.Wrap(err, "Error during the evaluation of the AddonDescriptor")
	}

	finalValues, err := runBuilds(conn, config, utils, valuesList)
	// files should be published, even if an error occured
	utils.publish()
	if err != nil {
		return err
	}

	cpe.abap.buildValues, err = convertValuesForCPE(finalValues)
	if err != nil {
		return errors.Wrap(err, "Error during the conversion of the values for the commonPipelineenvironment")
	}
	return nil
}

func runBuilds(conn *abapbuild.Connector, config *abapEnvironmentBuildOptions, utils abapEnvironmentBuildUtils, valuesList [][]abapbuild.Value) ([]abapbuild.Value, error) {
	var finalValues []abapbuild.Value
	// No addonDescriptor involved
	if len(valuesList) == 0 {
		values, err := generateValuesOnlyFromConfig(config)
		if err != nil {
			return finalValues, errors.Wrap(err, "Generating the values from config failed")
		}
		finalValues, err = runBuild(conn, config, utils, values)
		if err != nil {
			return finalValues, errors.Wrap(err, "Error during execution of build framework")
		}
	} else {
		// Run several times for each repository in the addonDescriptor
		var errstrings []string
		vE := valuesEvaluator{}
		vE.m = make(map[string]string)
		for _, values := range valuesList {
			cummulatedValues, err := generateValuesWithAddonDescriptor(config, values)
			if err != nil {
				return finalValues, errors.Wrap(err, "Error generating input values")
			}
			finalValuesForOneBuild, err := runBuild(conn, config, utils, cummulatedValues)
			if err != nil {
				err = errors.Wrapf(err, "Build with input values %s failed", values2string(values))
				if config.StopOnFirstError {
					return finalValues, err
				}
				errstrings = append(errstrings, err.Error())
			}
			finalValuesForOneBuild = removeAddonDescriptorValues(finalValuesForOneBuild, values)
			// This means: probably values are duplicated, but the first one wins -> perhaps change this in the future if needed
			if err := vE.appendValuesIfNotPresent(finalValuesForOneBuild, false); err != nil {
				errstrings = append(errstrings, err.Error())
			}
		}
		finalValues = vE.generateValueSlice()
		if len(errstrings) > 0 {
			finalError := errors.Errorf("%d out %d build runs failed:\n%s", len(errstrings), len(valuesList), (strings.Join(errstrings, "\n")))
			return finalValues, finalError
		}
	}
	return finalValues, nil
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
	connConfig.Parameters = url.Values{}
	if len(config.AbapSourceClient) != 0 {
		connConfig.Parameters.Add("sap-client", config.AbapSourceClient)
	}

	if err := conn.InitBuildFramework(connConfig, utils, utils); err != nil {
		return err
	}

	conn.MaxRuntime = utils.getMaxRuntime()
	conn.PollingInterval = utils.getPollingInterval()
	return nil
}

// ***********************************Run Build***************************************************************
func runBuild(conn *abapbuild.Connector, config *abapEnvironmentBuildOptions, utils abapEnvironmentBuildUtils, values []abapbuild.Value) ([]abapbuild.Value, error) {
	var finalValues []abapbuild.Value
	var inputValues abapbuild.Values
	inputValues.Values = values

	build := myBuild{
		Build: abapbuild.Build{
			Connector: *conn,
		},
		abapEnvironmentBuildOptions: config,
	}
	if err := build.Start(inputValues); err != nil {
		return finalValues, err
	}

	if err := build.Poll(); err != nil {
		return finalValues, errors.Wrap(err, "Error during the polling for the final state of the build run")
	}

	if err := build.PrintLogs(); err != nil {
		return finalValues, errors.Wrap(err, "Error printing the logs")
	}

	errBuildRun := build.EvaluteIfBuildSuccessful()

	if err := build.Download(); err != nil {
		if errBuildRun != nil {
			errWraped := errors.Errorf("Download failed after execution of build failed: %v. Build error: %v", err, errBuildRun)
			return finalValues, errWraped
		}
		return finalValues, err
	}
	if err := build.Publish(utils); err != nil {
		return finalValues, err
	}

	finalValues, err := build.GetFinalValues()
	if err != nil {
		return finalValues, err
	}
	return finalValues, errBuildRun
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

func (b *myBuild) Publish(utils abapEnvironmentBuildUtils) error {
	if b.PublishAllDownloadedResultFiles {
		b.PublishAllDownloadedResults("abapEnvironmentBuild", utils)
	} else {
		if err := b.PublishDownloadedResults("abapEnvironmentBuild", b.PublishResultFilenames, utils); err != nil {
			return errors.Wrapf(err, "Error during the publish of the result files %s", b.PublishResultFilenames)
		}
	}
	return nil
}

func (b *myBuild) GetFinalValues() ([]abapbuild.Value, error) {
	var values []abapbuild.Value
	if err := b.GetValues(); err != nil {
		return values, errors.Wrapf(err, "Error getting the values from build framework")
	}
	return b.Build.Values, nil
}

// **********************************Values Handling**************************************************************
func convertValuesForCPE(values []abapbuild.Value) (string, error) {
	type cpeValue struct {
		ValueID string `json:"value_id"`
		Value   string `json:"value"`
	}
	var cpeValues []cpeValue
	byt, err := json.Marshal(&values)
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

func removeAddonDescriptorValues(finalValuesFromBuild []abapbuild.Value, valuesFromAddonDescriptor []abapbuild.Value) []abapbuild.Value {
	var finalValues []abapbuild.Value
	mapForAddonDescriptorValues := make(map[string]string)
	for _, value := range valuesFromAddonDescriptor {
		mapForAddonDescriptorValues[value.ValueID] = value.Value
	}
	for _, value := range finalValuesFromBuild {
		_, present := mapForAddonDescriptorValues[value.ValueID]
		if !present {
			finalValues = append(finalValues, value)
		}
	}
	return finalValues
}

func generateValuesWithAddonDescriptor(config *abapEnvironmentBuildOptions, repoValues []abapbuild.Value) ([]abapbuild.Value, error) {
	var values []abapbuild.Value
	vE := valuesEvaluator{}
	// values from config
	if err := vE.initialize(config.Values); err != nil {
		return values, err
	}
	// values from addondescriptor
	if err := vE.appendValuesIfNotPresent(repoValues, true); err != nil {
		return values, err
	}
	// values from commonepipelineEnvironment
	if err := vE.appendStringValuesIfNotPresent(config.CpeValues, false); err != nil {
		return values, err
	}
	values = vE.generateValueSlice()
	return values, nil
}

func generateValuesOnlyFromConfig(config *abapEnvironmentBuildOptions) ([]abapbuild.Value, error) {
	return generateValuesWithAddonDescriptor(config, []abapbuild.Value{})
}

func generateValuesFromString(stringValues string) ([]abapbuild.Value, error) {
	var values []abapbuild.Value
	if len(stringValues) > 0 {
		if err := json.Unmarshal([]byte(stringValues), &values); err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return values, errors.Wrapf(err, "Could not convert the values %s", stringValues)
		}
	}
	return values, nil
}

type valuesEvaluator struct {
	m map[string]string
}

func (vE *valuesEvaluator) initialize(stringValues string) error {
	values, err := generateValuesFromString(stringValues)
	if err != nil {
		return errors.Wrapf(err, "Error converting the vales from the config")
	}
	vE.m = make(map[string]string)
	for _, value := range values {
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

func (vE *valuesEvaluator) appendStringValuesIfNotPresent(stringValues string, throwErrorIfPresent bool) error {
	var values []abapbuild.Value
	values, err := generateValuesFromString(stringValues)
	if err != nil {
		return errors.Wrapf(err, "Error converting the vales from the commonPipelineEnvironment")
	}
	if err := vE.appendValuesIfNotPresent(values, throwErrorIfPresent); err != nil {
		return err
	}
	return nil
}

func (vE *valuesEvaluator) appendValuesIfNotPresent(values []abapbuild.Value, throwErrorIfPresent bool) error {
	for _, value := range values {
		if value.ValueID == "PHASE" || value.ValueID == "BUILD_FRAMEWORK_MODE" {
			continue
		}
		_, present := vE.m[value.ValueID]
		if present {
			if throwErrorIfPresent {
				return errors.Errorf("Value_id %s already existed in the config", value.ValueID)
			}
			log.Entry().Infof("Value '%s':'%s' already existed -> discard this value", value.ValueID, value.Value)
		} else {
			vE.m[value.ValueID] = value.Value
		}
	}
	return nil
}

func (vE *valuesEvaluator) generateValueSlice() []abapbuild.Value {
	var values []abapbuild.Value
	var value abapbuild.Value
	for k, v := range vE.m {
		value.ValueID = k
		value.Value = v
		values = append(values, value)
	}
	return values
}

// **********************************Evaluate AddonDescriptor**************************************************************
type myRepo struct {
	abaputils.Repository
}

type condition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

type useField struct {
	Use    string `json:"use"`
	Rename string `json:"renameTo"`
}

func evaluateAddonDescriptor(config *abapEnvironmentBuildOptions) ([][]abapbuild.Value, error) {
	var listOfValuesList [][]abapbuild.Value
	if len(config.AddonDescriptor) == 0 && len(config.UseFieldsOfAddonDescriptor) > 0 {
		return listOfValuesList, errors.New("Config contains UseFieldsOfAddonDescriptor but no addonDescriptor is provided in the commonPipelineEnvironment")
	}
	if len(config.AddonDescriptor) > 0 {
		addonDescriptor := new(abaputils.AddonDescriptor)
		if err := addonDescriptor.InitFromJSONstring(config.AddonDescriptor); err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return listOfValuesList, errors.Wrap(err, "Error during the conversion of the AddonDescriptor")
		}
		for _, repo := range addonDescriptor.Repositories {
			myRepo := myRepo{
				Repository: repo,
			}
			use, err := myRepo.checkCondition(config)
			if err != nil {
				return listOfValuesList, errors.Wrapf(err, "Checking of ConditionOnAddonDescriptor failed")
			}
			if use {
				values, err := myRepo.generateValues(config)
				if err != nil {
					return listOfValuesList, errors.Wrap(err, "Error generating values from AddonDescriptor")
				}
				if len(values) > 0 {
					listOfValuesList = append(listOfValuesList, values)
				}
			}
		}
	}
	return listOfValuesList, nil
}

func (mR *myRepo) checkCondition(config *abapEnvironmentBuildOptions) (bool, error) {
	var conditions []condition
	if len(config.ConditionOnAddonDescriptor) > 0 {
		if err := json.Unmarshal([]byte(config.ConditionOnAddonDescriptor), &conditions); err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return false, errors.Wrapf(err, "Conversion of ConditionOnAddonDescriptor in the config failed")
		}
		for _, cond := range conditions {
			if cond.Field == "" || cond.Operator == "" || cond.Value == "" {
				log.SetErrorCategory(log.ErrorConfiguration)
				return false, errors.Errorf("Invalid condition for field %s with operator %s and value %s", cond.Field, cond.Operator, cond.Value)
			}
			use, err := mR.amI(cond.Field, cond.Operator, cond.Value)
			if err != nil {
				return false, errors.Wrapf(err, "Checking the field %s failed", cond.Field)
			}
			if !use {
				log.Entry().Infof("addonDescriptor with the name %s does not fulfil the requierement %s%s%s from the ConditionOnAddonDescriptor, therefore it is not used", mR.Name, cond.Field, cond.Operator, cond.Value)
				return false, nil
			}
			log.Entry().Infof("addonDescriptor with the name %s does fulfil the requierement %s%s%s in the ConditionOnAddonDescriptor", mR.Name, cond.Field, cond.Operator, cond.Value)
		}
	}
	return true, nil
}

func (mR *myRepo) generateValues(config *abapEnvironmentBuildOptions) ([]abapbuild.Value, error) {
	var values []abapbuild.Value
	var useFields []useField
	if len(config.UseFieldsOfAddonDescriptor) == 0 {
		log.Entry().Infof("UseFieldsOfAddonDescriptor is empty, nothing is used from the addonDescriptor")
	} else {
		if err := json.Unmarshal([]byte(config.UseFieldsOfAddonDescriptor), &useFields); err != nil {
			log.SetErrorCategory(log.ErrorConfiguration)
			return values, errors.Wrapf(err, "Conversion of UseFieldsOfAddonDescriptor in the config failed")
		}
		m := make(map[string]string)
		for _, uF := range useFields {
			if uF.Use == "" || uF.Rename == "" {
				log.SetErrorCategory(log.ErrorConfiguration)
				return values, errors.Errorf("Invalid UseFieldsOfAddonDescriptor for use %s and renameTo %s", uF.Use, uF.Rename)
			}
			m[uF.Use] = uF.Rename
		}

		fields := reflect.ValueOf(mR.Repository)
		typeOfS := fields.Type()
		for i := 0; i < fields.NumField(); i++ {
			var value abapbuild.Value
			ValueID := typeOfS.Field(i).Name
			rename, present := m[ValueID]
			if present {
				log.Entry().Infof("Use field %s from addonDescriptor and rename it to %s, the value is %s", ValueID, rename, fields.Field(i).String())
				value.ValueID = rename
				value.Value = fields.Field(i).String()
				values = append(values, value)
			}
		}
		if len(values) != len(useFields) {
			log.SetErrorCategory(log.ErrorConfiguration)
			return values, errors.Errorf("Not all fields in UseFieldsOfAddonDescriptor have been found. Probably a 'use' was used which does not exist")
		}
	}
	return values, nil
}

func (mR *myRepo) getField(field string) string {
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
	log.SetErrorCategory(log.ErrorConfiguration)
	return false, errors.Errorf("Invalid operator %s", operator)
}

type OperatorCallback map[string]func(string, string) bool

func Equal(a, b string) bool {
	return a == b
}

func Unequal(a, b string) bool {
	return a != b
}

func values2string(values []abapbuild.Value) string {
	var result string
	for index, value := range values {
		if index > 0 {
			result = result + "; "
		}
		result = result + value.ValueID + " = " + value.Value
	}
	return result
}
