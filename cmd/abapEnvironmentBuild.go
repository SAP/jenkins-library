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

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The abapEnvironmentBuildUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type abapEnvironmentBuildUtilsBundle struct {
	*command.Command
	*piperutils.Files
	*piperhttp.Client
	*abaputils.AbapUtils
	maxRuntime       time.Duration
	pollingIntervall time.Duration

	// Embed more structs as necessary to implement methods or interfaces you add to abapEnvironmentBuildUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// abapEnvironmentBuildUtilsBundle and forward to the implementation of the dependency.
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
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newAbapEnvironmentBuildUtils(time.Duration(config.MaxRuntimeInMinutes), time.Duration(config.PollingIntervallInSeconds))

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.

	err := runAbapEnvironmentBuild(&config, telemetryData, utils, cpe)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapEnvironmentBuild(config *abapEnvironmentBuildOptions, telemetryData *telemetry.CustomData, utils abapEnvironmentBuildUtils, cpe *abapEnvironmentBuildCommonPipelineEnvironment) error {

	//TODO checke mal warum da immer der falsche fehler steht..

	log.Entry().Info("blubblub")
	log.Entry().Info("and more stuff")
	//log.SetErrorCategory(log.ErrorConfiguration)
	return errors.New("Das ist ein völlig neuer einzigartiger FEHLER!")

	conn := new(abapbuild.Connector)

	// TODO wrappe die fehler
	if err := initConnection(conn, config, utils); err != nil {
		return err
	}
	values, err := generateValues(config)
	if err != nil {
		return err
	}

	//TODO delete
	log.Entry().Infof("Values used %s", values.Values)

	finalValues, err := runBuild(conn, config, utils, values)
	if err != nil {
		return err
	}
	cpe.build.values = finalValues

	//
	/*
		//TODO beginn generiertes beispielcoding
		// ***********************************************************************************************
		// Example of calling methods from external dependencies directly on utils:
		exists, err := utils.FileExists("file.txt")
		if err != nil {
			// It is good practice to set an error category.
			// Most likely you want to do this at the place where enough context is known.
			log.SetErrorCategory(log.ErrorConfiguration)
			// Always wrap non-descriptive errors to enrich them with context for when they appear in the log:
			return fmt.Errorf("failed to check for important file: %w", err)
		}
		if !exists {
			log.SetErrorCategory(log.ErrorConfiguration)
			return fmt.Errorf("cannot run without important file")
		}
		// ***********************************************************************************************
	*/

	return nil
}

//func initConnection(conn *abapbuild.Connector, config *abapEnvironmentBuildOptions, utils abapEnvironmentBuildUtils, client abapbuild.HTTPSendLoader, maxRuntime time.Duration, pollingIntervall time.Duration) error {
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
	conn.MaxRuntime = utils.getMaxRuntime()            //maxRuntime
	conn.PollingInterval = utils.getPollingIntervall() //pollingIntervall
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
		config: config,
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
	config *abapEnvironmentBuildOptions
}

func (b *myBuild) Start(values abapbuild.Values) error {
	if err := b.Build.Start(b.config.Phase, values); err != nil {
		return err
	}
	return nil
}

func (b *myBuild) EndedWithError() error {
	if err := b.Build.EndedWithError(b.config.TreatWarningsAsError); err != nil {
		return err
	}
	return nil
}

func (b *myBuild) Download() error {
	if b.config.DownloadAllResultFiles {
		if err := b.DownloadResults(b.config.SubDirectoryForDownload, b.config.FilenamePrefixForDownload); err != nil {
			return err
		}
	} else {
		//download nur spezifizierte
		for _, name := range b.config.DownloadResultFilenames {
			result, err := b.GetResult(name)
			if err != nil {
				return err
			}
			if err := result.DownloadWithFilenamePrefixAndTargetDirectory(b.config.SubDirectoryForDownload, b.config.FilenamePrefixForDownload); err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *myBuild) Publish(utils abapEnvironmentBuildUtils) error {
	if b.config.PublishAllDownloadedResultFiles {
		b.PublishAllDownloadedResults("abapEnvironmentBuild", utils)
	} else {
		if err := b.PublishDownloadedResults("abapEnvironmentBuild", b.config.PublishResultFilenames, utils); err != nil {
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
	byt, err := json.Marshal(&b.Values)
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
	log.Entry().Infof("config values %s", stringValues)
	if err := json.Unmarshal([]byte(stringValues), &vE.values); err != nil {
		return err
	}
	vE.m = make(map[string]string)

	// falls es einen wert doppelt in der config gibt -> fehler
	for _, value := range vE.values {
		_, present := vE.m[value.ValueID]
		if present {
			//TODO bessere error message
			return errors.New("Value duplicate")
		}
		vE.m[value.ValueID] = value.Value
	}
	return nil
}

func (vE *valuesEvaluator) appendValues(stringValues string) error {
	var values []abapbuild.Value
	//TODO delete
	log.Entry().Infof("cpe values %s", stringValues)
	if len(stringValues) > 0 {
		if err := json.Unmarshal([]byte(stringValues), &values); err != nil {
			return err
		}
		//wenn in der cpe ein wert steht der auch in der config steht, gewinnt config
		for i := len(values) - 1; i >= 0; i-- {
			_, present := vE.m[values[i].ValueID]
			if present || (values[i].ValueID == "PHASE") {
				//TODO delete
				log.Entry().Infof("remove value %s", values[i])
				values = append(values[:i], values[i+1:]...)
			}
		}
		vE.values = append(vE.values, values...)
	}
	return nil
}

//TODO delete
/*
//TODO delete
func parseValues(inputValues []string) (abapbuild.Values, error) {
	var values abapbuild.Values
	for _, inputValue := range inputValues {
		value, err := parseValue(inputValue)
		if err != nil {
			return values, err
		}
		values.Values = append(values.Values, value)
	}
	return values, nil
}

//TODO delete
func parseValue(inputValue string) (abapbuild.Value, error) {
	var value abapbuild.Value
	//TODO ; wieder durch , ersetzen -> und dann config.yml anpassen
	split := strings.Split(inputValue, ";")
	if len(split) != 2 {
		//TODO sinnvolle errormessage
		return value, errors.New("")
	}
	for _, v := range split {
		valueSplit := strings.Split(v, ":")
		if len(valueSplit) != 2 {
			//TODO sinnvolle errormessage
			return value, errors.New("")
		}
		if strings.TrimSpace(valueSplit[0]) == "value_id" {
			value.ValueID = strings.TrimSpace(valueSplit[1])
		}
		if strings.TrimSpace(valueSplit[0]) == "value" {
			value.Value = strings.TrimSpace(valueSplit[1])
		}
	}
	return value, nil
}

	config.Values = []string{"{value_id:'ID1','value':'Value1'}"}
	//config.Values = []string{"value_id: PACKAGES, value: /BUILD/AUNIT_DUMMY_TESTS", "value_id: MyId1, value: AunitValue1", "value_id: MyId2, value: AunitValue2"}
	config.CpeValues = []string{"value_id: PACKAGES, value: /BUILD/AUNIT_DUMMY_TESTS", "value_id: MyId1, value: CPEValue1", "value_id: MyId3, value: CPEValue3"}
	//TODO am ende erwartet: package ist eh gleich, soll MyId1, MyId2 von config.Values, MyId1 von CpeValues verwerfen, dafür MyId3 dazu packen
	var configValue abapbuild.Value
	var configValues abapbuild.Values
	for _, inputConfigValue := range config.Values {
		if err := json.Unmarshal([]byte(inputConfigValue), configValue); err != nil {
			return err
		}
		configValues.Values = append(configValues.Values, configValue)
	}
	var cpeValue abapbuild.Value
	var cpeValues abapbuild.Values
	for _, inputCpeValue := range config.CpeValues {
		if err := json.Unmarshal([]byte(inputCpeValue), cpeValue); err != nil {
			return err
		}
		cpeValues.Values = append(cpeValues.Values, cpeValue)
	}
*/

/*
	//TODO delete
	log.Entry().Infof("config values %s", config.Values)
	if err := json.Unmarshal([]byte(config.Values), &values.Values); err != nil {
		return err
	}
	m := make(map[string]string)

	// falls es einen wert doppelt in der config gibt -> fehler
	for _, value := range values.Values {
		_, present := m[value.ValueID]
		if present {
			//TODO bessere error message
			return errors.New("Value duplicate")
		}
		m[value.ValueID] = value.Value
	}

	var cpevalues abapbuild.Values
	//TODO delete
	log.Entry().Infof("cpe values %s", config.CpeValues)
	if len(config.CpeValues) > 0 {
		if err := json.Unmarshal([]byte(config.CpeValues), &cpevalues.Values); err != nil {
			return err
		}
		//wenn in der cpe ein wert steht der auch in der config steht, gewinnt config
		for i := len(cpevalues.Values) - 1; i >= 0; i-- {
			_, present := m[cpevalues.Values[i].ValueID]
			if present || (cpevalues.Values[i].ValueID == "PHASE") {
				//TODO delete
				log.Entry().Infof("remove value %s", cpevalues.Values[i])
				cpevalues.Values = append(cpevalues.Values[:i], cpevalues.Values[i+1:]...)
			}
		}

		values.Values = append(values.Values, cpevalues.Values...)
	}
*/

//TODO delete
//************************build
/*

	build := abapbuild.Build{
		Connector: *conn,
	}
	if err := build.Start(config.Phase, values); err != nil {
		return err
	}
	if err := build.Poll(); err != nil {
		return err
	}
	if err := build.PrintLogs(); err != nil {
		return err
	}
	if err := build.EndedWithError(config.TreatWarningsAsError); err != nil {
		return err
	}

		if config.DownloadAllResultFiles {
		if err := build.DownloadResults(config.SubDirectoryForDownload, config.FilenamePrefixForDownload); err != nil {
			return err
		}
	} else {
		//download nur spezifizierte
		for _, name := range config.DownloadResultFilenames {
			result, err := build.GetResult(name)
			if err != nil {
				return err
			}
			if err := result.DownloadWithFilenamePrefixAndTargetDirectory(config.SubDirectoryForDownload, config.FilenamePrefixForDownload); err != nil {
				return err
			}
		}
	}

		if config.PublishAllDownloadedResultFiles {
		build.PublishAllDownloadedResults("abapEnvironmentBuild", utils)
	} else {
		if err := build.PublishDownloadedResults("abapEnvironmentBuild", config.PublishResultFilenames, utils); err != nil {
			return err
		}
	}

		type cpeValue struct {
		ValueID string `json:"value_id"`
		Value   string `json:"value"`
	}

	build.GetValues()
	var cpeValues []cpeValue
	byt, _ := json.Marshal(&build.Values)
	json.Unmarshal(byt, &cpeValues)
	jsonBytes, _ := json.Marshal(cpeValues)
	cpe.build.values = string(jsonBytes)
*/

//TODO checke mal warum da immer der falsche fehler steht..
/*
	log.Entry().Info("blubblub")
	log.Entry().Info("and more stuff")
	log.SetErrorCategory(log.ErrorConfiguration)
	return errors.New("Das ist ein FEHLER!")
*/
