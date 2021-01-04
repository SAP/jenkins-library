package cmd

import (
	"bytes"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
	"net/http"
	"path/filepath"
	"strings"
	"text/template"
)

type executeNewmanUtils interface {
	Glob(pattern string) (matches []string, err error)

	RunShell(shell, script string) error
	RunExecutable(executable string, params ...string) error
}

type executeNewmanUtilsBundle struct {
	*command.Command
	*piperutils.Files

	// Embed more structs as necessary to implement methods or interfaces you add to executeNewmanUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// executeNewmanUtilsBundle and forward to the implementation of the dependency.
}

func newExecuteNewmanUtils() executeNewmanUtils {
	utils := executeNewmanUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

func executeNewman(config executeNewmanOptions, _ *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newExecuteNewmanUtils()

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err := runExecuteNewman(&config, utils)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runExecuteNewman(config *executeNewmanOptions, utils executeNewmanUtils) error {

	collectionList, err := utils.Glob(config.NewmanCollection)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.Wrapf(err, "Could not execute global search for '%v'", config.NewmanCollection)
	}

	if collectionList == nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return fmt.Errorf("no collection found with pattern '%v'", config.NewmanCollection)
	} else {
		log.Entry().Infof("Found files '%v'", collectionList)
	}

	err = logVersions(utils)
	// TODO: should error in version logging cause failure?
	if err != nil {
		return err
	}

	err = installNewman(config.NewmanInstallCommand, utils)
	if err != nil {
		return err
	}

	for _, collection := range collectionList {
		cmd, err := resolveTemplate(config, collection)
		if err != nil {
			return err
		}

		commandSecrets := ""
		hasSecrets := len(config.CfAppsWithSecrets) > 0
		if hasSecrets {
			//	CloudFoundry cfUtils = new CloudFoundry(script); // TODO: ???
			for _, appName := range config.CfAppsWithSecrets {
				var clientId, clientSecret string
				// def xsuaaCredentials = cfUtils.getXsuaaCredentials(config.cloudFoundry.apiEndpoint, // TODO: ???
				// config.cloudFoundry.org,
				// config.cloudFoundry.space,
				// config.cloudFoundry.credentialsId,
				// appName,
				// config.verbose ? true : false ) //to avoid config.verbose as "null" if undefined in yaml and since function parameter boolean

				commandSecrets += " --env-var " + appName + "_clientid=" + clientId + " --env-var " + appName + "_clientsecret=" + clientSecret
				// TODO: How to do echo in golang?
				// echo "Exposing client id and secret for ${appName}: as ${appName}_clientid and ${appName}_clientsecret to newman"
			}
		}

		if !config.FailOnError {
			cmd += " --suppress-exit-code"
		}

		if hasSecrets {
			//	echo "PATH=\$PATH:~/.npm-global/bin newman ${command} **env/secrets**" // TODO: How to do this?

			//utils.SetDir(".") // TODO: Need this?
			err := utils.RunShell("/bin/sh", "set +x")
			if err != nil {
				log.SetErrorCategory(log.ErrorService)
				return errors.Wrap(err, "The execution of the newman tests failed, see the log for details.")
			}
		}

		args := []string{"PATH=\\$PATH:~/.npm-global/bin newman", cmd}
		if hasSecrets {
			args = append(args, commandSecrets)
		}
		script := strings.Join(args, " ")
		//utils.SetDir(".") // TODO: Need this?
		err = utils.RunShell("/bin/sh", script)
		if err != nil {
			log.SetErrorCategory(log.ErrorService)
			return errors.Wrap(err, "The execution of the newman tests failed, see the log for details.")
		}
	}

	return nil
}

func logVersions(utils executeNewmanUtils) error {
	//utils.SetDir(".") // TODO: Need this?
	//returnStatus: true // TODO: How to do this? If necessary at all.
	err := utils.RunExecutable("node", "--version")
	if err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		return errors.Wrap(err, "error logging node version")
	}

	//utils.SetDir(".") // TODO: Need this?
	//returnStatus: true // TODO: How to do this? If necessary at all.
	err = utils.RunExecutable("npm", "--version")
	if err != nil {
		log.SetErrorCategory(log.ErrorInfrastructure)
		return errors.Wrap(err, "error logging npm version")
	}

	return nil
}

func installNewman(newmanInstallCommand string, utils executeNewmanUtils) error {
	args := []string{"NPM_CONFIG_PREFIX=~/.npm-global", newmanInstallCommand}
	script := strings.Join(args, " ")
	//utils.SetDir(".") // TODO: Need this?
	err := utils.RunShell("/bin/sh", script)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return errors.Wrap(err, "error installing newman")
	}
	return nil
}

func resolveTemplate(config *executeNewmanOptions, collection string) (string, error) {
	collectionDisplayName := defineCollectionDisplayName(collection)

	type TemplateConfig struct {
		Config                interface{}
		CollectionDisplayName string
		NewmanCollection      string
		// TODO: New field as structs cannot be extended in Go
	}

	templ, err := template.New("template").Parse(config.NewmanRunCommand)
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return "", errors.Wrap(err, "could not parse newman command template")
	}
	buf := new(bytes.Buffer)
	// TODO: Config and CollectionDisplayName must be capitalized <-> was small letter in groovy --> Templates must be adapted
	err = templ.Execute(buf, TemplateConfig{
		Config:                config,
		CollectionDisplayName: collectionDisplayName,
		NewmanCollection:      collection,
	})
	if err != nil {
		log.SetErrorCategory(log.ErrorConfiguration)
		return "", errors.Wrap(err, "error on executing template")
	}
	cmd := buf.String()
	return cmd, nil
}

func defineCollectionDisplayName(collection string) string {
	replacedSeparators := strings.Replace(collection, string(filepath.Separator), "_", -1)
	return strings.Split(replacedSeparators, ".")[0]
}

func getXsuaaCredentials(apiEndpoint, org, space, credentialsId, appName string, verbose bool) (string, string, error) {
	return getAppEnvironment(apiEndpoint, org, space, credentialsId, appName, verbose)
}

func getAppEnvironment(apiEndpoint, org, space, credentialsId, appName string, verbose bool) (string, string, error) {

	authEndpoint, err := getAuthEndPoint(apiEndpoint, verbose)
	if err != nil {
		return "", "", err
	}

	_, err = getBearerToken(authEndpoint, credentialsId, verbose)
	if err != nil {
		return "", "", err
	}

	//	appUrl := getAppRefUrl(apiEndpoint, org, space, bearerToken, appName, verbose)
	//
	//	response := script.httpRequest
	//url:
	//	"${appUrl}/env", quiet: !verbose,
	//		customHeaders:[[name: 'Authorization', value: "${bearerToken}"]]
	//def envJson = script.readJSON text:"${response.content}"
	//return envJson
	return "", "", nil
}

func getAuthEndPoint(apiEndpoint string, verbose bool) (string, error) {
	// TODO: need full struct here?
	type responseJson struct {
		authorization_endpoint string
	}

	response, err := http.Get(apiEndpoint + "/v2/info") // TODO: Verbose
	if err != nil {
		return "", err
	}
	resJson := responseJson{}
	err = piperhttp.ParseHTTPResponseBodyJSON(response, resJson)
	if err != nil {
		return "", err
	}
	return resJson.authorization_endpoint, nil
}

func getBearerToken(authorizationEndpoint, credentialsId string, verbose bool) (string, error) {
	return "", nil

	//	client := &http.Client{}
	//	req, err := http.NewRequest("GET", "mydomain.com", nil)
	//	if err != nil {
	//		return "", err
	//	}
	//	req.SetBasicAuth(username, passwd)
	//	resp, err := client.Do(req)
	// TODO: How to handle multiple credentials in GO?
	//	script.withCredentials([script.usernamePassword(credentialsId: credentialsId, usernameVariable: 'usercf', passwordVariable: 'passwordcf')]) {
	//def token = script.httpRequest url:"${authorizationEndpoint}/oauth/token", quiet: !verbose,
	//httpMode:'POST',
	//requestBody: "username=${script.usercf}&password=${script.passwordcf}&client_id=cf&grant_type=password&response_type=token",
	//customHeaders: [[name: 'Content-Type', value: 'application/x-www-form-urlencoded'], [name: 'Authorization', value: 'Basic Y2Y6']]
	//def responseJson = script.readJSON text:"${token.content}"
	//return "Bearer ${responseJson.access_token.trim()}"
	//}
}
