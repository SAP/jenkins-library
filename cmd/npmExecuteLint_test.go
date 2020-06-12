package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/bmatcuk/doublestar"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"sort"
	"strings"
	"testing"
)

type npmLintUtilsBundle struct {
	execRunner mock.ExecMockRunner
	files      map[string]string
}

func (u *npmLintUtilsBundle) fileWrite(path string, content []byte, perm os.FileMode) error {
	u.files[path] = string(content)
	return nil
}

func (u *npmLintUtilsBundle) getExecRunner() execRunner {
	return &u.execRunner
}

func (u *npmLintUtilsBundle) getGeneralPurposeConfig(configURL string) error {
	_ = u.fileWrite(".pipeline/.eslintrc.json", []byte(`abc`), 666)
	return nil
}

// duplicated from nexusUpload_test.go for now, refactor later?
func (u *npmLintUtilsBundle) glob(pattern string) ([]string, error) {
	var matches []string
	for path := range u.files {
		matched, _ := doublestar.Match(pattern, path)
		if matched {
			matches = append(matches, path)
		}
	}
	// The order in m.files is not deterministic, this would result in flaky tests.
	sort.Sort(byLen(matches))
	return matches, nil
}

type npmExecuteOptions struct {
	install            bool
	runScripts         []string
	runOptions         []string
	defaultNpmRegistry string
	defaultSapRegistry string
}

type npmExecutorMock struct {
	utils   npmLintUtilsBundle
	options npmExecuteOptions
}

// FindPackageJSONFilesWithScript mocking implementation
func (n *npmExecutorMock) FindPackageJSONFilesWithScript(packageJSONFiles []string, script string) ([]string, error) {
	var packagesWithScript []string
	for _, file := range packageJSONFiles {
		if strings.Contains(n.utils.files[file], script) {
			packagesWithScript = append(packagesWithScript, file)
		}
	}
	return packagesWithScript, nil
}

// ExecuteAllScripts mocking implementation
func (n *npmExecutorMock) ExecuteAllScripts() error {
	packageJSONFiles := n.FindPackageJSONFiles()

	for _, script := range n.options.runScripts {
		packagesWithScript, err := n.FindPackageJSONFilesWithScript(packageJSONFiles, script)
		if err != nil {
			return err
		}

		if len(packagesWithScript) == 0 {
			log.Entry().Warnf("could not find any package.json file with script " + script)
			continue
		}
		npmRunArgs := []string{"run", script}
		if len(n.options.runOptions) > 0 {
			npmRunArgs = append(npmRunArgs, n.options.runOptions...)
		}

		for range packagesWithScript {
			err = n.utils.execRunner.RunExecutable("npm", npmRunArgs...)
			if err != nil {
				return fmt.Errorf("failed to run npm script %s: %w", script, err)
			}
		}
	}
	return nil
}

// InstallAllDependencies mocking implementation
func (n *npmExecutorMock) InstallAllDependencies(packageJSONFiles []string) error {
	for _, packageJSON := range packageJSONFiles {
		dir := path.Dir(packageJSON)
		if dir == "." {
			dir = ""
		} else {
			dir = dir + "/"
		}
		_, ok := n.utils.files[dir+"package-lock.json"]
		if ok {
			err := n.utils.execRunner.RunExecutable("npm", "ci")
			if err != nil {
				return err
			}
			continue
		}
		_, ok = n.utils.files[dir+"yarn.lock"]
		if ok {
			err := n.utils.execRunner.RunExecutable("yarn", "install", "--frozen-lockfile")
			if err != nil {
				return err
			}
			continue
		}
		err := n.utils.execRunner.RunExecutable("npm", "install")
		if err != nil {
			return err
		}
	}
	return nil
}

// SetNpmRegistries mocking implementation
func (n *npmExecutorMock) SetNpmRegistries() error {
	const sapRegistry = "@sap:registry"
	const npmRegistry = "registry"
	configurableRegistries := []string{npmRegistry, sapRegistry}
	for _, registry := range configurableRegistries {
		var buffer bytes.Buffer
		n.utils.execRunner.Stdout(&buffer)
		err := n.utils.execRunner.RunExecutable("npm", "config", "get", "registry")
		preConfiguredRegistry := buffer.String()
		if strings.HasPrefix(preConfiguredRegistry, "undefined") {
			if registry == npmRegistry {
				err = n.utils.execRunner.RunExecutable("npm", "config", "set", registry, n.options.defaultNpmRegistry)
				if err != nil {
					return err
				}
			}
			if registry == sapRegistry {
				err = n.utils.execRunner.RunExecutable("npm", "config", "set", registry, n.options.defaultSapRegistry)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// FindPackageJSONFiles mocking implementation
func (n *npmExecutorMock) FindPackageJSONFiles() []string {
	var packageJSONFiles []string
	for key, _ := range n.utils.files {
		if strings.Contains(key, "package.json") {
			if !(strings.Contains(key, "node_modules") || strings.HasPrefix(key, "gen/") || strings.Contains(key, "/gen/")) {
				packageJSONFiles = append(packageJSONFiles, key)
			}
		}
	}
	return packageJSONFiles
}

func TestNpmExecuteLint(t *testing.T) {
	t.Run("Call with ci-lint script and one package.json", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.files["package.json"] = "{\"scripts\": { \"ci-lint\": \"\" } }"

		config := npmExecuteLintOptions{}

		npmExecutor := npmExecutorMock{utils: utils, options: npmExecuteOptions{
			install:            false,
			runScripts:         []string{"ci-lint"},
			runOptions:         []string{"--silent"},
			defaultNpmRegistry: config.DefaultNpmRegistry,
			defaultSapRegistry: config.SapNpmRegistry,
		}}
		err := runNpmExecuteLint(&npmExecutor, &utils, &config)

		assert.NoError(t, err)
		assert.Equal(t, 1, len(npmExecutor.utils.execRunner.Calls))
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "ci-lint", "--silent"}}, npmExecutor.utils.execRunner.Calls[0])
	})

	t.Run("Call with ci-lint script and two package.json", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.files["package.json"] = "{\"scripts\": { \"ci-lint\": \"\" } }"
		utils.files["src/package.json"] = "{\"scripts\": { \"ci-lint\": \"\" } }"
		config := npmExecuteLintOptions{}
		config.DefaultNpmRegistry = "foo.bar"

		npmExecutor := npmExecutorMock{utils: utils, options: npmExecuteOptions{
			install:            false,
			runScripts:         []string{"ci-lint"},
			runOptions:         []string{"--silent"},
			defaultNpmRegistry: config.DefaultNpmRegistry,
			defaultSapRegistry: config.SapNpmRegistry,
		}}
		err := runNpmExecuteLint(&npmExecutor, &utils, &config)

		assert.NoError(t, err)
		assert.Equal(t, 2, len(npmExecutor.utils.execRunner.Calls))
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "ci-lint", "--silent"}}, npmExecutor.utils.execRunner.Calls[0])
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "ci-lint", "--silent"}}, npmExecutor.utils.execRunner.Calls[1])
	})

	t.Run("Call default with ESLint config from user", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.files["package.json"] = "{\"name\": \"Test\" }"
		utils.files[".eslintrc.json"] = "{\"name\": \"Test\" }"
		config := npmExecuteLintOptions{}
		config.DefaultNpmRegistry = "foo.bar"

		npmExecutor := npmExecutorMock{utils: utils, options: npmExecuteOptions{
			install:            false,
			runScripts:         []string{"ci-lint"},
			runOptions:         []string{"--silent"},
			defaultNpmRegistry: config.DefaultNpmRegistry,
			defaultSapRegistry: config.SapNpmRegistry,
		}}
		err := runNpmExecuteLint(&npmExecutor, &utils, &config)

		assert.NoError(t, err)
		assert.Equal(t, 1, len(utils.execRunner.Calls))
		assert.Equal(t, mock.ExecCall{Exec: "npx", Params: []string{"eslint", ".", "-f", "checkstyle", "-o", "./0_defaultlint.xml", "--ignore-pattern", "node_modules/", "--ignore-pattern", ".eslintrc.js"}}, utils.execRunner.Calls[0])

	})

	t.Run("Call default with two ESLint configs from user", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.files["package.json"] = "{\"name\": \"Test\" }"
		utils.files[".eslintrc.json"] = "{\"name\": \"Test\" }"
		utils.files["src/.eslintrc.json"] = "{\"name\": \"Test\" }"
		config := npmExecuteLintOptions{}
		config.DefaultNpmRegistry = "foo.bar"

		npmExecutor := npmExecutorMock{utils: utils, options: npmExecuteOptions{
			install:            false,
			runScripts:         []string{"ci-lint"},
			runOptions:         []string{"--silent"},
			defaultNpmRegistry: config.DefaultNpmRegistry,
			defaultSapRegistry: config.SapNpmRegistry,
		}}
		err := runNpmExecuteLint(&npmExecutor, &utils, &config)

		assert.NoError(t, err)
		assert.Equal(t, 2, len(utils.execRunner.Calls))
		assert.Equal(t, mock.ExecCall{Exec: "npx", Params: []string{"eslint", ".", "-f", "checkstyle", "-o", "./0_defaultlint.xml", "--ignore-pattern", "node_modules/", "--ignore-pattern", ".eslintrc.js"}}, utils.execRunner.Calls[0])
		assert.Equal(t, mock.ExecCall{Exec: "npx", Params: []string{"eslint", "src/**/*.js", "-f", "checkstyle", "-o", "./1_defaultlint.xml", "--ignore-pattern", "node_modules/", "--ignore-pattern", ".eslintrc.js"}}, utils.execRunner.Calls[1])
	})

	t.Run("Default without ESLint config", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.files["package.json"] = "{\"name\": \"Test\" }"
		config := npmExecuteLintOptions{}
		config.DefaultNpmRegistry = "foo.bar"

		npmExecutor := npmExecutorMock{utils: utils, options: npmExecuteOptions{
			install:            false,
			runScripts:         []string{"ci-lint"},
			runOptions:         []string{"--silent"},
			defaultNpmRegistry: config.DefaultNpmRegistry,
			defaultSapRegistry: config.SapNpmRegistry,
		}}
		err := runNpmExecuteLint(&npmExecutor, &utils, &config)

		assert.NoError(t, err)
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"install", "eslint@^7.0.0", "typescript@^3.7.4", "@typescript-eslint/parser@^3.0.0", "@typescript-eslint/eslint-plugin@^3.0.0"}}, utils.execRunner.Calls[0])
		assert.Equal(t, mock.ExecCall{Exec: "npx", Params: []string{"--no-install", "eslint", ".", "--ext", ".js,.jsx,.ts,.tsx", "-c", ".pipeline/.eslintrc.json", "-f", "checkstyle", "-o", "./defaultlint.xml", "--ignore-pattern", ".eslintrc.js"}}, utils.execRunner.Calls[1])
		assert.Equal(t, 2, len(utils.execRunner.Calls))
	})

	t.Run("Ignore ESLint config in node_modules and .pipeline/", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.files["package.json"] = "{\"name\": \"Test\" }"
		utils.files["node_modules/.eslintrc.json"] = "{\"name\": \"Test\" }"
		utils.files[".pipeline/.eslintrc.json"] = "{\"name\": \"Test\" }"
		config := npmExecuteLintOptions{}
		config.DefaultNpmRegistry = "foo.bar"

		npmExecutor := npmExecutorMock{utils: utils, options: npmExecuteOptions{
			install:            false,
			runScripts:         []string{"ci-lint"},
			runOptions:         []string{"--silent"},
			defaultNpmRegistry: config.DefaultNpmRegistry,
			defaultSapRegistry: config.SapNpmRegistry,
		}}
		err := runNpmExecuteLint(&npmExecutor, &utils, &config)

		assert.NoError(t, err)
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"install", "eslint@^7.0.0", "typescript@^3.7.4", "@typescript-eslint/parser@^3.0.0", "@typescript-eslint/eslint-plugin@^3.0.0"}}, utils.execRunner.Calls[0])
		assert.Equal(t, mock.ExecCall{Exec: "npx", Params: []string{"--no-install", "eslint", ".", "--ext", ".js,.jsx,.ts,.tsx", "-c", ".pipeline/.eslintrc.json", "-f", "checkstyle", "-o", "./defaultlint.xml", "--ignore-pattern", ".eslintrc.js"}}, utils.execRunner.Calls[1])
		assert.Equal(t, 2, len(utils.execRunner.Calls))
	})

	t.Run("Call with ci-lint script and failOnError", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.files["package.json"] = "{\"scripts\": { \"ci-lint\": \"\" } }"
		utils.execRunner = mock.ExecMockRunner{ShouldFailOnCommand: map[string]error{"npm run ci-lint --silent": errors.New("exit 1")}}
		config := npmExecuteLintOptions{}
		config.FailOnError = true
		config.DefaultNpmRegistry = "foo.bar"

		npmExecutor := npmExecutorMock{utils: utils, options: npmExecuteOptions{
			install:            false,
			runScripts:         []string{"ci-lint"},
			runOptions:         []string{"--silent"},
			defaultNpmRegistry: config.DefaultNpmRegistry,
			defaultSapRegistry: config.SapNpmRegistry,
		}}
		err := runNpmExecuteLint(&npmExecutor, &utils, &config)

		assert.EqualError(t, err, "failed to run npm script ci-lint: exit 1")
		assert.Equal(t, 1, len(npmExecutor.utils.execRunner.Calls))
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "ci-lint", "--silent"}}, npmExecutor.utils.execRunner.Calls[0])
	})

	t.Run("Call default with ESLint config from user and failOnError", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.files["package.json"] = "{\"name\": \"Test\" }"
		utils.files[".eslintrc.json"] = "{\"name\": \"Test\" }"
		utils.execRunner = mock.ExecMockRunner{ShouldFailOnCommand: map[string]error{"eslint . -f checkstyle -o ./0_defaultlint.xml --ignore-pattern node_modules/ --ignore-pattern .eslintrc.js": errors.New("exit 1")}}
		config := npmExecuteLintOptions{}
		config.FailOnError = true
		config.DefaultNpmRegistry = "foo.bar"

		npmExecutor := npmExecutorMock{utils: utils, options: npmExecuteOptions{
			install:            false,
			runScripts:         []string{"ci-lint"},
			runOptions:         []string{"--silent"},
			defaultNpmRegistry: config.DefaultNpmRegistry,
			defaultSapRegistry: config.SapNpmRegistry,
		}}
		err := runNpmExecuteLint(&npmExecutor, &utils, &config)

		assert.EqualError(t, err, "failed to run ESLint with config .eslintrc.json: exit 1")
		assert.Equal(t, mock.ExecCall{Exec: "npx", Params: []string{"eslint", ".", "-f", "checkstyle", "-o", "./0_defaultlint.xml", "--ignore-pattern", "node_modules/", "--ignore-pattern", ".eslintrc.js"}}, utils.execRunner.Calls[0])
		assert.Equal(t, 1, len(utils.execRunner.Calls))
	})

}

func newNpmMockUtilsBundle() npmLintUtilsBundle {
	utils := npmLintUtilsBundle{}
	utils.files = map[string]string{}
	return utils
}
