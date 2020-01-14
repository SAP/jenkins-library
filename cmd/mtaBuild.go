package cmd

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"gopkg.in/yaml.v2"
)

const descriptorFile = "mta.yaml"

type Mta struct {
	Id string `yaml:"ID,omitempty"`
}

func mtaBuild(config mtaBuildOptions) error {
	updateVersion()

	var buildCommand string
	var options []string
	options = append(options, fmt.Sprintf("--mtar %v", getName(config)))
	switch config.MtaBuildTool {
	case "cloudMbt":
		options = prepareOptions(options, config)
		buildCommand = fmt.Sprintf("mbt build %v", options)
	case "classic":
		// If it is not configured, it is expected on the PATH
		var jarPath = "mta.jar"
		if len(config.MtaJarLocation) > 0 {
			jarPath = config.MtaJarLocation
		}
		options = prepareLegacyOptions(options, config)
		buildCommand = fmt.Sprintf("java -jar %v %v build", jarPath, options)
	default:
		log.Entry().
			WithField("mtaBuildTool", config.MtaBuildTool).
			Error("MTA build tool '${configuration.mtaBuildTool}' not supported!")
	}

	log.Entry().Info("Executing mta build call: '${mtaCall}'.")

	//[Q]: Why extending the path?
	//  [A]: To be sure e.g. grunt can be found
	//[Q]: Why escaping \$PATH ?
	//  [A]: We want to extend the PATH variable in e.g. the container and not substituting it with the Jenkins environment when using ${PATH}
	//[Q]: Why escaping \\$PATH a second time?
	//  [A]: To make sure GO can handle it.
	var customPath string = "PATH=\\$PATH:./node_modules/.bin"

	buildCommandTokens := tokenize(buildCommand)
	c := command.Command{}
	if err := c.RunExecutable(fmt.Sprintf("%v %v", customPath, buildCommandTokens[0]), buildCommandTokens[1:]...); err != nil {
		log.Entry().
			WithError(err).
			WithField("command", buildCommand).
			Fatal("failed to execute build command")
	}

	//TODO: write mta name to piper env
	//script?.commonPipelineEnvironment?.setMtarFilePath("${mtarName}")

	return nil
}

func prepareOptions(options []string, configuration mtaBuildOptions) []string {
	options = append(options, fmt.Sprintf("--platform %v", configuration.Platform))
	options = append(options, "--target ./")
	if len(configuration.Extension) > 0 {
		options = append(options, fmt.Sprintf("--extensions %v", configuration.Extension))
	}
	return options
}

func prepareLegacyOptions(options []string, configuration mtaBuildOptions) []string {
	options = append(options, fmt.Sprintf("--build-target %v", configuration.BuildTarget))
	options = append(options, "--target ./")
	if len(configuration.Extension) > 0 {
		options = append(options, fmt.Sprintf("--extension %v", configuration.Extension))
	}
	return options
}

func getName(configuration mtaBuildOptions) (name string) {
	name = strings.TrimSpace(configuration.MtarName)
	if len(name) <= 0 {
		name = fmt.Sprintf("%v.mtar", getMtaIdFromFile())
	}
	return
}

func getMtaIdFromFile() string {
	var mta Mta

	data, err := ioutil.ReadFile(descriptorFile)
	if err != nil {
		log.Entry().WithField("file", descriptorFile).Fatal("Could not read file.")
	}

	yaml.Unmarshal(data, &mta)

	if len(mta.Id) <= 0 {
		log.Entry().
			WithField("file", descriptorFile).
			Fatal("Property 'ID' not found in file.")
	}
	return mta.Id
}

func updateVersion() {
	/*	if exists, err := piperutils.FileExists(descriptorFile); !exists || err != nil {
			log.Entry().
				WithError(err).
				Fatal("File not found.")
		}
	*/
	c := command.Command{}
	updateCommand := fmt.Sprintf("-ie \"s/\\\\${timestamp}/`date +%%Y%%m%%d%%H%%M%%S`/g\" \"%v\"", descriptorFile)
	if err := c.RunExecutable("sed", updateCommand); err != nil {
		log.Entry().
			WithError(err).
			WithField("command", updateCommand).
			Fatal("failed to execute command")
	}
}
