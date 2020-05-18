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
	SetEnv(e []string)
	Stdout(out io.Writer)
}

type execRunner interface {
	runner
	RunExecutable(e string, p ...string) error
}

func SetNpmRegistries(options *NpmRegistryOptions, execRunner execRunner) error {
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

		if registry == npmRegistry && options.DefaultNpmRegistry != "" && registryRequiresConfiguration(preConfiguredRegistry, "https://registry.npmjs.org") {
			log.Entry().Info("npm registry " + registry + " was not configured, setting it to " + options.DefaultNpmRegistry)
			err = execRunner.RunExecutable("npm", "config", "set", registry, options.DefaultNpmRegistry)
			if err != nil {
				return err
			}
		}

		if registry == sapRegistry && registryRequiresConfiguration(preConfiguredRegistry, "https://npm.sap.com") {
			log.Entry().Info("npm registry " + registry + " was not configured, setting it to " + options.SapNpmRegistry)
			err = execRunner.RunExecutable("npm", "config", "set", registry, options.SapNpmRegistry)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func registryRequiresConfiguration(preConfiguredRegistry, url string) bool {
	return strings.HasPrefix(preConfiguredRegistry, "undefined") || strings.HasPrefix(preConfiguredRegistry, url)
}
