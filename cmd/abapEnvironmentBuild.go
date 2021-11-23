package cmd

import (
	"fmt"
	"path"
	"path/filepath"
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

func abapEnvironmentBuild(config abapEnvironmentBuildOptions, telemetryData *telemetry.CustomData) {
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
	err := runAbapEnvironmentBuild(&config, telemetryData, utils, &autils, &client)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapEnvironmentBuild(config *abapEnvironmentBuildOptions, telemetryData *telemetry.CustomData, utils abapEnvironmentBuildUtils, com abaputils.Communication, client abapbuild.HTTPSendLoader) error {
	conn := new(abapbuild.Connector)

	// TODO wrappe die fehler
	if err := initConnection(conn, config, com, client); err != nil {
		return err
	}

	//erzeuge value liste
	// TODO lieber in bfw?
	values, err := parseValues(config.Values)
	if err != nil {
		return err
	}
	var build abapbuild.Build
	build.Connector = *conn
	if err := build.Start(config.Phase, values); err != nil {
		return err
	}
	if err := build.Poll(time.Duration(config.MaxRuntimeInMinutes), time.Duration(config.PollingIntervallInSeconds)); err != nil {
		return err
	}
	if err := build.PrintLogs(); err != nil {
		return err
	}
	if err := build.EndedWithError(config.TreatWarningsAsError); err != nil {
		return err
	}

	//download == true?
	if config.DownloadResultFiles {
		if len(config.ResultFilenames) > 0 {
			for _, name := range config.ResultFilenames {
				result, err := build.GetResult(name)
				if err != nil {
					return err
				}
				// TODO wohin speichern?
				var fileName string
				if len(result.AdditionalInfo) <= 255 {
					fileName = result.AdditionalInfo
				} else {
					fileName = result.Name
				}
				envPath := filepath.Join(GeneralConfig.EnvRootPath, "abapBuild")
				downloadPath := filepath.Join(envPath, path.Base(fileName))
				if err := result.Download(downloadPath); err != nil {
					return err
				}
			}
		}
		// TODO alle downloaden, hier?
		/*	if err := build.GetResults(); err != nil {
					return err
				}
				for _, task := range b.tasks {
					task.
			// TODO
			//}
		*/
	}
	//spezielle download files nur?
	//download
	//

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
	// TODO macht runtime überhaupt was
	connConfig.MaxRuntimeInMinutes = config.MaxRuntimeInMinutes

	log.Entry().Infof("Host config %s , Host connConfig %s", config.Host, connConfig.Host)

	err := conn.InitBuildFramework(connConfig, com, client)
	if err != nil {
		return errors.Wrap(err, "Connector initialization for communication with the ABAP system failed")
	}

	// TODO an besseren ort schieben, jetzt nur zum testen
	conn.Client.SetOptions(piperhttp.ClientOptions{
		//TrustedCerts: []string{"yi3Cert.cer"},
		TrustedCerts: config.CustomTLSCertificateNames,
	})

	return nil
}

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

func parseValue(inputValue string) (abapbuild.Value, error) {
	var value abapbuild.Value
	split := strings.Split(inputValue, ",")
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
