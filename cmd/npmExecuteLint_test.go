package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type mockLintUtilsBundle struct {
	*mock.FilesMock
	execRunner *mock.ExecMockRunner
}

func (u *mockLintUtilsBundle) getExecRunner() execRunner {
	return u.execRunner
}

func (u *mockLintUtilsBundle) getGeneralPurposeConfig(configURL string) error {
	u.AddFile(filepath.Join(".pipeline", ".eslintrc.json"), []byte(`abc`))
	return nil
}

func newNpmMockUtilsBundle() mockLintUtilsBundle {
	utils := mockLintUtilsBundle{FilesMock: &mock.FilesMock{}, execRunner: &mock.ExecMockRunner{}}
	return utils
}

type npmExecuteOptions struct {
	install            bool
	runScripts         []string
	runOptions         []string
	defaultNpmRegistry string
	sapNpmRegistry     string
}

type npmExecutorMock struct {
	utils   mockLintUtilsBundle
	options npmExecuteOptions
}

// FindPackageJSONFilesWithScript mocking implementation
func (n *npmExecutorMock) FindPackageJSONFilesWithScript(packageJSONFiles []string, script string) ([]string, error) {
	var packagesWithScript []string
	for _, file := range packageJSONFiles {
		fileContent, err := n.utils.FileRead(file)
		if err != nil {
			return nil, err
		}
		if strings.Contains(string(fileContent), script) {
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
			if !(n.options.install) {
				// set in each directory to respect existing config in rc fileUtils
				err = n.SetNpmRegistries()
				if err != nil {
					return err
				}
			}

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
		dir := filepath.Dir(packageJSON)
		if dir == "." {
			dir = ""
		} else {
			dir = dir + string(os.PathSeparator)
		}
		if n.utils.HasFile(dir + "package-lock.json") {
			err := n.utils.execRunner.RunExecutable("npm", "ci")
			if err != nil {
				return err
			}
		} else if n.utils.HasFile(dir + "yarn.lock") {
			err := n.utils.execRunner.RunExecutable("yarn", "install", "--frozen-lockfile")
			if err != nil {
				return err
			}
		} else {
			err := n.utils.execRunner.RunExecutable("npm", "install")
			if err != nil {
				return err
			}
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
		err := n.utils.execRunner.RunExecutable("npm", "config", "get", registry)
		preConfiguredRegistry := buffer.String()
		if strings.HasPrefix(preConfiguredRegistry, "undefined") {
			if registry == npmRegistry && n.options.defaultNpmRegistry != "" {
				err = n.utils.execRunner.RunExecutable("npm", "config", "set", registry, n.options.defaultNpmRegistry)
				if err != nil {
					return err
				}
			}
			if registry == sapRegistry && n.options.sapNpmRegistry != "" {
				err = n.utils.execRunner.RunExecutable("npm", "config", "set", registry, n.options.sapNpmRegistry)
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
	unfilteredPackageJSONFiles, _ := n.utils.Glob("**/package.json")

	for _, file := range unfilteredPackageJSONFiles {
		if !(strings.Contains(file, "node_modules") || strings.HasPrefix(file, "gen"+string(os.PathSeparator)) || strings.Contains(file, string(os.PathSeparator)+"gen"+string(os.PathSeparator))) {
			packageJSONFiles = append(packageJSONFiles, file)
		}
	}
	return packageJSONFiles
}

func TestNpmExecuteLint(t *testing.T) {
	t.Run("Call with ci-lint script and one package.json", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"\" } }"))

		config := npmExecuteLintOptions{}

		npmExecutor := npmExecutorMock{utils: utils, options: npmExecuteOptions{
			install:            false,
			runScripts:         []string{"ci-lint"},
			runOptions:         []string{"--silent"},
			defaultNpmRegistry: config.DefaultNpmRegistry,
			sapNpmRegistry:     config.SapNpmRegistry,
		}}
		err := runNpmExecuteLint(&npmExecutor, &utils, &config)

		if assert.NoError(t, err) {
			if assert.Equal(t, 3, len(utils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "ci-lint", "--silent"}}, utils.execRunner.Calls[2])
			}
		}
	})

	t.Run("Call with ci-lint script and two package.json", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"\" } }"))
		utils.AddFile(filepath.Join("src", "package.json"), []byte("{\"scripts\": { \"ci-lint\": \"\" } }"))
		config := npmExecuteLintOptions{}
		config.DefaultNpmRegistry = "foo.bar"

		npmExecutor := npmExecutorMock{utils: utils, options: npmExecuteOptions{
			install:            false,
			runScripts:         []string{"ci-lint"},
			runOptions:         []string{"--silent"},
			defaultNpmRegistry: config.DefaultNpmRegistry,
			sapNpmRegistry:     config.SapNpmRegistry,
		}}
		err := runNpmExecuteLint(&npmExecutor, &utils, &config)

		if assert.NoError(t, err) {
			if assert.Equal(t, 6, len(utils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "ci-lint", "--silent"}}, utils.execRunner.Calls[2])
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "ci-lint", "--silent"}}, utils.execRunner.Calls[5])
			}
		}
	})

	t.Run("Call default with ESLint config from user", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile(".eslintrc.json", []byte("{\"name\": \"Test\" }"))
		config := npmExecuteLintOptions{}
		config.DefaultNpmRegistry = "foo.bar"

		npmExecutor := npmExecutorMock{utils: utils, options: npmExecuteOptions{
			install:            false,
			runScripts:         []string{"ci-lint"},
			runOptions:         []string{"--silent"},
			defaultNpmRegistry: config.DefaultNpmRegistry,
			sapNpmRegistry:     config.SapNpmRegistry,
		}}
		err := runNpmExecuteLint(&npmExecutor, &utils, &config)

		if assert.NoError(t, err) {
			if assert.Equal(t, 3, len(utils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npx", Params: []string{"eslint", ".", "-f", "checkstyle", "-o", "./0_defaultlint.xml", "--ignore-pattern", "node_modules/", "--ignore-pattern", ".eslintrc.js"}}, utils.execRunner.Calls[2])
			}
		}
	})

	t.Run("Call default with two ESLint configs from user", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile(".eslintrc.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile(filepath.Join("src", ".eslintrc.json"), []byte("{\"name\": \"Test\" }"))
		config := npmExecuteLintOptions{}
		config.DefaultNpmRegistry = "foo.bar"

		npmExecutor := npmExecutorMock{utils: utils, options: npmExecuteOptions{
			install:            false,
			runScripts:         []string{"ci-lint"},
			runOptions:         []string{"--silent"},
			defaultNpmRegistry: config.DefaultNpmRegistry,
			sapNpmRegistry:     config.SapNpmRegistry,
		}}
		err := runNpmExecuteLint(&npmExecutor, &utils, &config)

		if assert.NoError(t, err) {
			if assert.Equal(t, 4, len(utils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npx", Params: []string{"eslint", ".", "-f", "checkstyle", "-o", "./0_defaultlint.xml", "--ignore-pattern", "node_modules/", "--ignore-pattern", ".eslintrc.js"}}, utils.execRunner.Calls[2])
				assert.Equal(t, mock.ExecCall{Exec: "npx", Params: []string{"eslint", "src/**/*.js", "-f", "checkstyle", "-o", "./1_defaultlint.xml", "--ignore-pattern", "node_modules/", "--ignore-pattern", ".eslintrc.js"}}, utils.execRunner.Calls[3])
			}
		}
	})

	t.Run("Default without ESLint config", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		config := npmExecuteLintOptions{}
		config.DefaultNpmRegistry = "foo.bar"

		npmExecutor := npmExecutorMock{utils: utils, options: npmExecuteOptions{
			install:            false,
			runScripts:         []string{"ci-lint"},
			runOptions:         []string{"--silent"},
			defaultNpmRegistry: config.DefaultNpmRegistry,
			sapNpmRegistry:     config.SapNpmRegistry,
		}}
		err := runNpmExecuteLint(&npmExecutor, &utils, &config)

		if assert.NoError(t, err) {
			if assert.Equal(t, 4, len(utils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"install", "eslint@^7.0.0", "typescript@^3.7.4", "@typescript-eslint/parser@^3.0.0", "@typescript-eslint/eslint-plugin@^3.0.0"}}, utils.execRunner.Calls[2])
				assert.Equal(t, mock.ExecCall{Exec: "npx", Params: []string{"--no-install", "eslint", ".", "--ext", ".js,.jsx,.ts,.tsx", "-c", ".pipeline/.eslintrc.json", "-f", "checkstyle", "-o", "./defaultlint.xml", "--ignore-pattern", ".eslintrc.js"}}, utils.execRunner.Calls[3])
			}
		}
	})

	t.Run("Ignore ESLint config in node_modules and .pipeline/", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile(filepath.Join("node_modules", ".eslintrc.json"), []byte("{\"name\": \"Test\" }"))
		utils.AddFile(filepath.Join(".pipeline", ".eslintrc.json"), []byte("{\"name\": \"Test\" }"))
		config := npmExecuteLintOptions{}
		config.DefaultNpmRegistry = "foo.bar"

		npmExecutor := npmExecutorMock{utils: utils, options: npmExecuteOptions{
			install:            false,
			runScripts:         []string{"ci-lint"},
			runOptions:         []string{"--silent"},
			defaultNpmRegistry: config.DefaultNpmRegistry,
			sapNpmRegistry:     config.SapNpmRegistry,
		}}
		err := runNpmExecuteLint(&npmExecutor, &utils, &config)

		if assert.NoError(t, err) {
			if assert.Equal(t, 4, len(utils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"install", "eslint@^7.0.0", "typescript@^3.7.4", "@typescript-eslint/parser@^3.0.0", "@typescript-eslint/eslint-plugin@^3.0.0"}}, utils.execRunner.Calls[2])
				assert.Equal(t, mock.ExecCall{Exec: "npx", Params: []string{"--no-install", "eslint", ".", "--ext", ".js,.jsx,.ts,.tsx", "-c", ".pipeline/.eslintrc.json", "-f", "checkstyle", "-o", "./defaultlint.xml", "--ignore-pattern", ".eslintrc.js"}}, utils.execRunner.Calls[3])
			}
		}

	})

	t.Run("Call with ci-lint script and failOnError", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"\" } }"))
		utils.execRunner = &mock.ExecMockRunner{ShouldFailOnCommand: map[string]error{"npm run ci-lint --silent": errors.New("exit 1")}}
		config := npmExecuteLintOptions{}
		config.FailOnError = true
		config.DefaultNpmRegistry = "foo.bar"

		npmExecutor := npmExecutorMock{utils: utils, options: npmExecuteOptions{
			install:            false,
			runScripts:         []string{"ci-lint"},
			runOptions:         []string{"--silent"},
			defaultNpmRegistry: config.DefaultNpmRegistry,
			sapNpmRegistry:     config.SapNpmRegistry,
		}}
		err := runNpmExecuteLint(&npmExecutor, &utils, &config)

		if assert.EqualError(t, err, "ci-lint script execution failed with error: failed to run npm script ci-lint: exit 1. This might be the result of severe linting findings, or some other issue while executing the script. Please examine the linting results in the UI, the ci-lint.xml file, if available, or the log above. ") {
			if assert.Equal(t, 3, len(utils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "ci-lint", "--silent"}}, utils.execRunner.Calls[2])
			}
		}
	})

	t.Run("Call default with ESLint config from user and failOnError", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		utils.AddFile(".eslintrc.json", []byte("{\"name\": \"Test\" }"))
		utils.execRunner = &mock.ExecMockRunner{ShouldFailOnCommand: map[string]error{"eslint . -f checkstyle -o ./0_defaultlint.xml --ignore-pattern node_modules/ --ignore-pattern .eslintrc.js": errors.New("exit 1")}}
		config := npmExecuteLintOptions{}
		config.FailOnError = true
		config.DefaultNpmRegistry = "foo.bar"

		npmExecutor := npmExecutorMock{utils: utils, options: npmExecuteOptions{
			install:            false,
			runScripts:         []string{"ci-lint"},
			runOptions:         []string{"--silent"},
			defaultNpmRegistry: config.DefaultNpmRegistry,
			sapNpmRegistry:     config.SapNpmRegistry,
		}}
		err := runNpmExecuteLint(&npmExecutor, &utils, &config)

		if assert.EqualError(t, err, "Lint execution failed. This might be the result of severe linting findings, problems with the provided ESLint configuration (.eslintrc.json), or another issue. Please examine the linting results in the UI or in 0_defaultlint.xml, if available, or the log above. ") {
			if assert.Equal(t, 3, len(utils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npx", Params: []string{"eslint", ".", "-f", "checkstyle", "-o", "./0_defaultlint.xml", "--ignore-pattern", "node_modules/", "--ignore-pattern", ".eslintrc.js"}}, utils.execRunner.Calls[2])
			}
		}
	})
}
