package npm

import (
	"bytes"
	"github.com/SAP/jenkins-library/pkg/log"
	"io"
	"strings"

)

type NpmRegistryOptions struct {
	DefaultNpmRegistry string
	SapNpmRegistry     string
}

type runner interface {
	SetDir(d string)
	SetEnv(e []string)
	Stdout(out io.Writer)
	Stderr(err io.Writer)
}

type execRunner interface {
	runner
	RunExecutable(e string, p ...string) error
}
//fixme import or copy the type?

func SetNpmRegistries(options *NpmRegistryOptions, execRunner execRunner) error {
	environment := []string{}
	const sapRegistry = "@sap:registry"
	const npmRegistry = "registry"
	configurableRegistries := []string{npmRegistry, sapRegistry}
	for _, registry := range configurableRegistries {
		var buffer bytes.Buffer
		execRunner.Stdout(&buffer)
		err := execRunner.RunExecutable("npm", "config", "get", registry)
		execRunner.Stdout(log.Writer())
		if err != nil {
			return err
		}
		preConfiguredRegistry := buffer.String()

		log.Entry().Info("Discovered pre-configured npm registry " + preConfiguredRegistry)

		if registry == npmRegistry && options.DefaultNpmRegistry != "" && (strings.HasPrefix(preConfiguredRegistry, "undefined") || strings.HasPrefix(preConfiguredRegistry, "https://registry.npmjs.org")) {
			log.Entry().Info("npm registry " + registry + " was not configured, setting it to " + options.DefaultNpmRegistry)
			environment = append(environment, "npm_config_"+registry+"="+options.DefaultNpmRegistry)
		}

		if registry == sapRegistry && (strings.HasPrefix(preConfiguredRegistry, "undefined") || strings.HasPrefix(preConfiguredRegistry, "https://npm.sap.com")) {
			log.Entry().Info("npm registry " + registry + " was not configured, setting it to " + options.SapNpmRegistry)
			environment = append(environment, "npm_config_"+registry+"="+options.SapNpmRegistry)
		}
	}

	log.Entry().Info("Setting environment: " + strings.Join(environment, ", "))
	execRunner.SetEnv(environment)
	return nil
}
