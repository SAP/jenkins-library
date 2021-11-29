package cmd

import (
	"encoding/json"
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

	FileExists(filename string) (bool, error)

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The abapEnvironmentBuildUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type abapEnvironmentBuildUtilsBundle struct {
	*command.Command
	*piperutils.Files

	// Embed more structs as necessary to implement methods or interfaces you add to abapEnvironmentBuildUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// abapEnvironmentBuildUtilsBundle and forward to the implementation of the dependency.
}

func newAbapEnvironmentBuildUtils() abapEnvironmentBuildUtils {
	utils := abapEnvironmentBuildUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func abapEnvironmentBuild(config abapEnvironmentBuildOptions, telemetryData *telemetry.CustomData, cpe *abapEnvironmentBuildCommonPipelineEnvironment) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newAbapEnvironmentBuildUtils()

	// TODO das irgendwie in die utils rein?
	c := command.Command{}
	var autils = abaputils.AbapUtils{
		Exec: &c,
	}
	client := piperhttp.Client{}

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runAbapEnvironmentBuild(&config, telemetryData, utils, &autils, &client, time.Duration(config.MaxRuntimeInMinutes)*time.Minute, time.Duration(config.PollingIntervallInSeconds)*time.Second, cpe)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapEnvironmentBuild(config *abapEnvironmentBuildOptions, telemetryData *telemetry.CustomData, utils abapEnvironmentBuildUtils, com abaputils.Communication, client abapbuild.HTTPSendLoader,
	maxRuntime time.Duration, PollingIntervall time.Duration, cpe *abapEnvironmentBuildCommonPipelineEnvironment) error {
	conn := new(abapbuild.Connector)

	// TODO wrappe die fehler
	if err := initConnection(conn, config, com, client); err != nil {
		return err
	}

	//stringValues := "[{\"value_id\":\"ID1\",\"value\":\"Value1\"}]"
	var values abapbuild.Values
	if err := json.Unmarshal([]byte(config.Values), &values.Values); err != nil {
		return err
	}
	var cpevalues abapbuild.Values
	if err := json.Unmarshal([]byte(config.CpeValues), &cpevalues.Values); err != nil {
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

	//wenn in der cpe ein wert steht der auch in der config steht, gewinnt config
	for i := len(cpevalues.Values) - 1; i >= 0; i-- {
		_, present := m[cpevalues.Values[i].ValueID]
		if present {
			cpevalues.Values = append(cpevalues.Values[:i], cpevalues.Values[i+1:]...)
		}
	}

	values.Values = append(values.Values, cpevalues.Values...)

	//erzeuge value liste
	// TODO lieber in bfw?
	//addonDescriptorCPE, _ := abaputils.ConstructAddonDescriptorFromJSON([]byte(config.AddonDescriptor))

	/*values, err := parseValues(config.Values)
	if err != nil {
		return err
	} */

	build := abapbuild.Build{
		Connector: *conn,
	}

	if err := build.Start(config.Phase, values); err != nil {
		return err
	}
	if err := build.Poll(maxRuntime, PollingIntervall); err != nil {
		return err
	}
	if err := build.PrintLogs(); err != nil {
		return err
	}
	if err := build.EndedWithError(config.TreatWarningsAsError); err != nil {
		return err
	}

	if config.PublishAllDownloadedResultFiles {
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
			if err := result.DownloadWithFilenamePrefix(config.SubDirectoryForDownload, config.FilenamePrefixForDownload); err != nil {
				return err
			}
		}
	}

	if config.PublishAllDownloadedResultFiles {
		build.PublishAllDownloadedResults("abapEnvironmentBuild")
	} else {
		if err := build.PublishDownloadedResults("abapEnvironmentBuild", config.PublishResultFilenames); err != nil {
			return err
		}
	}

	//TODO values von build nehmen und in die struktur pressen und das wegschreiben
	//TODO test start
	type cpeValue struct {
		ValueID string `json:"value_id"`
		Value   string `json:"value"`
	}

	build.GetValues()
	var testValues []cpeValue
	byt, _ := json.Marshal(&build.Values)
	json.Unmarshal(byt, &testValues)

	//Das brauch ich um am ende die Values wegzuschreiben -> muss also eigentlich nach utnen
	jsonBytes, _ := json.Marshal(testValues)
	cpe.build.values = string(jsonBytes)
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

func initConnection(conn *abapbuild.Connector, config *abapEnvironmentBuildOptions, com abaputils.Communication, client abapbuild.HTTPSendLoader) error {
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

	err := conn.InitBuildFramework(connConfig, com, client)
	if err != nil {
		return errors.Wrap(err, "Connector initialization for communication with the ABAP system failed")
	}
	return nil
}

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

//TODO delete
/*
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
