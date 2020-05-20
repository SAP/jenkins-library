package npm

import (
	"bytes"
	"github.com/SAP/jenkins-library/pkg/log"
	"io"
	"strings"
)

// RegistryOptions holds the configured urls for npm registries
type RegistryOptions struct {
	DefaultNpmRegistry string
	SapNpmRegistry     string
}

type execRunner interface {
	Stdout(out io.Writer)
	RunExecutable(executable string, params ...string) error
}

// SetNpmRegistries configures the given npm registries.
// CAUTION: This will change the npm configuration in the user's home directory.
func SetNpmRegistries(options *RegistryOptions, execRunner execRunner) error {
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

		if registryIsNonEmpty(preConfiguredRegistry) {
			log.Entry().Info("Discovered pre-configured npm registry " + registry + " with value " + preConfiguredRegistry)
		}

		if registry == npmRegistry && options.DefaultNpmRegistry != "" && registryRequiresConfiguration(preConfiguredRegistry, "https://registry.npmjs.org") {
			log.Entry().Info("npm registry " + registry + " was not configured, setting it to " + options.DefaultNpmRegistry)
			err = execRunner.RunExecutable("npm", "config", "set", registry, options.DefaultNpmRegistry)
			if err != nil {
				return err
			}
		}

		if registry == sapRegistry && options.SapNpmRegistry != "" && registryRequiresConfiguration(preConfiguredRegistry, "https://npm.sap.com") {
			log.Entry().Info("npm registry " + registry + " was not configured, setting it to " + options.SapNpmRegistry)
			err = execRunner.RunExecutable("npm", "config", "set", registry, options.SapNpmRegistry)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func registryIsNonEmpty(preConfiguredRegistry string) bool {
	return !strings.HasPrefix(preConfiguredRegistry, "undefined") && len(preConfiguredRegistry) > 0
}

func registryRequiresConfiguration(preConfiguredRegistry, url string) bool {
	return strings.HasPrefix(preConfiguredRegistry, "undefined") || strings.HasPrefix(preConfiguredRegistry, url)
}
