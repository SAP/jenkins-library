//go:build unit
// +build unit

package cmd

import (
	"errors"
	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/npm"
	"github.com/stretchr/testify/assert"
	"path/filepath"
	"testing"
)

type mockLintUtilsBundle struct {
	*mock.FilesMock
	execRunner *mock.ExecMockRunner
}

func (u *mockLintUtilsBundle) getExecRunner() command.ExecRunner {
	return u.execRunner
}

func (u *mockLintUtilsBundle) getGeneralPurposeConfig(configURL string) {
	u.AddFile(filepath.Join(".pipeline", ".eslintrc.json"), []byte(`abc`))
}

func newLintMockUtilsBundle() mockLintUtilsBundle {
	utils := mockLintUtilsBundle{FilesMock: &mock.FilesMock{}, execRunner: &mock.ExecMockRunner{}}
	return utils
}

func TestNpmExecuteLint(t *testing.T) {
	defaultConfig := npmExecuteLintOptions{RunScript: "ci-lint", OutputFormat: "checkstyle", OutputFileName: "defaultlint.xml"}

	t.Run("Call with ci-lint script and one package.json", func(t *testing.T) {
		lintUtils := newLintMockUtilsBundle()
		lintUtils.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"\" } }"))

		npmUtils := npm.NewNpmMockUtilsBundle()
		npmUtils.ExecRunner = lintUtils.execRunner
		npmUtils.FilesMock = lintUtils.FilesMock

		config := defaultConfig
		config.FailOnError = true

		npmExecutor := npm.NpmExecutorMock{Utils: npmUtils, Config: npm.NpmConfig{RunScripts: []string{"ci-lint"}, RunOptions: []string{"--silent"}}}
		err := runNpmExecuteLint(&npmExecutor, &lintUtils, &config)

		assert.NoError(t, err)
	})

	t.Run("Call default with ESLint config from user", func(t *testing.T) {
		lintUtils := newLintMockUtilsBundle()
		lintUtils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		lintUtils.AddFile(".eslintrc.json", []byte("{\"name\": \"Test\" }"))

		config := defaultConfig
		config.DefaultNpmRegistry = "foo.bar"

		npmUtils := newNpmMockUtilsBundle()
		npmUtils.execRunner = lintUtils.execRunner
		npmExecutor := npm.Execute{Utils: &npmUtils, Options: npm.ExecutorOptions{}}

		err := runNpmExecuteLint(&npmExecutor, &lintUtils, &config)

		if assert.NoError(t, err) {
			if assert.Equal(t, 2, len(lintUtils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npx", Params: []string{
					"eslint",
					".",
					"-f", "checkstyle",
					"--ignore-pattern", "node_modules/",
					"--ignore-pattern", ".eslintrc.js",
					"-o", "./0_defaultlint.xml"}}, lintUtils.execRunner.Calls[1])
			}
		}
	})

	t.Run("Call default with ESLint config from user - no redirect to file, stylish format", func(t *testing.T) {
		lintUtils := newLintMockUtilsBundle()
		lintUtils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		lintUtils.AddFile(".eslintrc.json", []byte("{\"name\": \"Test\" }"))

		config := npmExecuteLintOptions{RunScript: "ci-lint", OutputFormat: "stylish", OutputFileName: ""}
		config.DefaultNpmRegistry = "foo.bar"

		npmUtils := newNpmMockUtilsBundle()
		npmUtils.execRunner = lintUtils.execRunner
		npmExecutor := npm.Execute{Utils: &npmUtils, Options: npm.ExecutorOptions{}}

		err := runNpmExecuteLint(&npmExecutor, &lintUtils, &config)

		if assert.NoError(t, err) {
			if assert.Equal(t, 2, len(lintUtils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npx", Params: []string{
					"eslint",
					".",
					"-f", "stylish",
					"--ignore-pattern",
					"node_modules/",
					"--ignore-pattern", ".eslintrc.js",
					// no -o, --output-file in this case.
				}}, lintUtils.execRunner.Calls[1])
			}
		}
	})

	t.Run("Call default with two ESLint configs from user", func(t *testing.T) {
		lintUtils := newLintMockUtilsBundle()
		lintUtils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		lintUtils.AddFile(".eslintrc.json", []byte("{\"name\": \"Test\" }"))
		lintUtils.AddFile(filepath.Join("src", ".eslintrc.json"), []byte("{\"name\": \"Test\" }"))

		config := defaultConfig
		config.DefaultNpmRegistry = "foo.bar"

		npmUtils := newNpmMockUtilsBundle()
		npmUtils.execRunner = lintUtils.execRunner
		npmExecutor := npm.Execute{Utils: &npmUtils, Options: npm.ExecutorOptions{}}

		err := runNpmExecuteLint(&npmExecutor, &lintUtils, &config)

		if assert.NoError(t, err) {
			if assert.Equal(t, 3, len(lintUtils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npx", Params: []string{
					"eslint",
					".",
					"-f", "checkstyle",
					"--ignore-pattern", "node_modules/",
					"--ignore-pattern", ".eslintrc.js",
					"-o", "./0_defaultlint.xml",
				}}, lintUtils.execRunner.Calls[1])
				assert.Equal(t, mock.ExecCall{Exec: "npx", Params: []string{
					"eslint",
					"src/**/*.js",
					"-f", "checkstyle",
					"--ignore-pattern", "node_modules/",
					"--ignore-pattern", ".eslintrc.js",
					"-o", "./1_defaultlint.xml",
				}}, lintUtils.execRunner.Calls[2])
			}
		}
	})

	t.Run("Call default with two ESLint configs from user - no redirect to file, stylish format", func(t *testing.T) {
		lintUtils := newLintMockUtilsBundle()
		lintUtils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		lintUtils.AddFile(".eslintrc.json", []byte("{\"name\": \"Test\" }"))
		lintUtils.AddFile(filepath.Join("src", ".eslintrc.json"), []byte("{\"name\": \"Test\" }"))

		config := defaultConfig
		config.DefaultNpmRegistry = "foo.bar"
		config.OutputFormat = "stylish"
		config.OutputFileName = ""

		npmUtils := newNpmMockUtilsBundle()
		npmUtils.execRunner = lintUtils.execRunner
		npmExecutor := npm.Execute{Utils: &npmUtils, Options: npm.ExecutorOptions{}}

		err := runNpmExecuteLint(&npmExecutor, &lintUtils, &config)

		if assert.NoError(t, err) {
			if assert.Equal(t, 3, len(lintUtils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npx", Params: []string{
					"eslint",
					".",
					"-f", "stylish",
					"--ignore-pattern", "node_modules/",
					"--ignore-pattern", ".eslintrc.js",
					// no  -o --output-file in this case.
				}}, lintUtils.execRunner.Calls[1])
				assert.Equal(t, mock.ExecCall{Exec: "npx", Params: []string{
					"eslint",
					"src/**/*.js",
					"-f", "stylish",
					"--ignore-pattern", "node_modules/",
					"--ignore-pattern", ".eslintrc.js",
					// no  -o --output-file in this case.
				}}, lintUtils.execRunner.Calls[2])
			}
		}
	})

	t.Run("Default without ESLint config", func(t *testing.T) {
		lintUtils := newLintMockUtilsBundle()
		lintUtils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))

		config := defaultConfig
		config.DefaultNpmRegistry = "foo.bar"

		npmUtils := newNpmMockUtilsBundle()
		npmUtils.execRunner = lintUtils.execRunner
		npmExecutor := npm.Execute{Utils: &npmUtils, Options: npm.ExecutorOptions{}}

		err := runNpmExecuteLint(&npmExecutor, &lintUtils, &config)

		if assert.NoError(t, err) {
			if assert.Equal(t, 3, len(lintUtils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"install", "eslint@^7.0.0", "typescript@^3.7.4", "@typescript-eslint/parser@^3.0.0", "@typescript-eslint/eslint-plugin@^3.0.0"}}, lintUtils.execRunner.Calls[1])
				assert.Equal(t, mock.ExecCall{Exec: "npx", Params: []string{
					"--no-install",
					"eslint",
					".",
					"--ext",
					".js,.jsx,.ts,.tsx",
					"-c", ".pipeline/.eslintrc.json",
					"-f", "checkstyle",
					"--ignore-pattern", ".eslintrc.js",
					"-o", "./defaultlint.xml",
				}}, lintUtils.execRunner.Calls[2])
			}
		}
	})

	t.Run("Default without ESLint config - no redirect to file, stylish format", func(t *testing.T) {
		lintUtils := newLintMockUtilsBundle()
		lintUtils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))

		config := defaultConfig
		config.DefaultNpmRegistry = "foo.bar"
		config.OutputFormat = "stylish"
		config.OutputFileName = ""

		npmUtils := newNpmMockUtilsBundle()
		npmUtils.execRunner = lintUtils.execRunner
		npmExecutor := npm.Execute{Utils: &npmUtils, Options: npm.ExecutorOptions{}}

		err := runNpmExecuteLint(&npmExecutor, &lintUtils, &config)

		if assert.NoError(t, err) {
			if assert.Equal(t, 3, len(lintUtils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"install", "eslint@^7.0.0", "typescript@^3.7.4", "@typescript-eslint/parser@^3.0.0", "@typescript-eslint/eslint-plugin@^3.0.0"}}, lintUtils.execRunner.Calls[1])
				assert.Equal(t, mock.ExecCall{Exec: "npx", Params: []string{
					"--no-install",
					"eslint",
					".",
					"--ext",
					".js,.jsx,.ts,.tsx",
					"-c", ".pipeline/.eslintrc.json",
					"-f", "stylish",
					"--ignore-pattern", ".eslintrc.js",
					// no -o --output-file in this case.
				}}, lintUtils.execRunner.Calls[2])
			}
		}
	})

	t.Run("Call with ci-lint script and failOnError", func(t *testing.T) {
		lintUtils := newLintMockUtilsBundle()
		lintUtils.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"\" } }"))
		lintUtils.execRunner = &mock.ExecMockRunner{ShouldFailOnCommand: map[string]error{"npm run ci-lint --silent": errors.New("exit 1")}}

		config := defaultConfig
		config.FailOnError = true
		config.DefaultNpmRegistry = "foo.bar"

		npmUtils := newNpmMockUtilsBundle()
		npmUtils.execRunner = lintUtils.execRunner
		npmUtils.FilesMock = lintUtils.FilesMock
		npmExecutor := npm.Execute{Utils: &npmUtils, Options: npm.ExecutorOptions{}}

		err := runNpmExecuteLint(&npmExecutor, &lintUtils, &config)

		if assert.EqualError(t, err, "ci-lint script execution failed with error: failed to run npm script ci-lint: exit 1. This might be the result of severe linting findings, or some other issue while executing the script. Please examine the linting results in the UI, the cilint.xml file, if available, or the log above. ") {
			if assert.Equal(t, 2, len(lintUtils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "ci-lint", "--silent"}}, lintUtils.execRunner.Calls[1])
			}
		}
	})

	t.Run("Call default with ESLint config from user and failOnError", func(t *testing.T) {
		lintUtils := newLintMockUtilsBundle()
		lintUtils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		lintUtils.AddFile(".eslintrc.json", []byte("{\"name\": \"Test\" }"))
		lintUtils.execRunner = &mock.ExecMockRunner{ShouldFailOnCommand: map[string]error{
			"eslint . -f checkstyle --ignore-pattern node_modules/ --ignore-pattern .eslintrc.js -o ./0_defaultlint.xml": errors.New("exit 1")}}

		config := defaultConfig
		config.FailOnError = true
		config.DefaultNpmRegistry = "foo.bar"

		npmUtils := newNpmMockUtilsBundle()
		npmUtils.execRunner = lintUtils.execRunner
		npmExecutor := npm.Execute{Utils: &npmUtils, Options: npm.ExecutorOptions{}}

		err := runNpmExecuteLint(&npmExecutor, &lintUtils, &config)

		if assert.EqualError(t, err, "Lint execution failed. This might be the result of severe linting findings, problems with the provided ESLint configuration (.eslintrc.json), or another issue. Please examine the linting results in the UI or in 0_defaultlint.xml, if available, or the log above. ") {
			if assert.Equal(t, 2, len(lintUtils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npx", Params: []string{
					"eslint",
					".",
					"-f", "checkstyle",
					"--ignore-pattern", "node_modules/",
					"--ignore-pattern", ".eslintrc.js",
					"-o", "./0_defaultlint.xml",
				}}, lintUtils.execRunner.Calls[1])
			}
		}
	})

	t.Run("Call default with ESLint config from user and failOnError - no redirect to file, stylish format", func(t *testing.T) {
		lintUtils := newLintMockUtilsBundle()
		lintUtils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		lintUtils.AddFile(".eslintrc.json", []byte("{\"name\": \"Test\" }"))
		lintUtils.execRunner = &mock.ExecMockRunner{ShouldFailOnCommand: map[string]error{
			"eslint . -f stylish --ignore-pattern node_modules/ --ignore-pattern .eslintrc.js": errors.New("exit 1")}}

		config := defaultConfig
		config.FailOnError = true
		config.OutputFormat = "stylish"
		config.OutputFileName = ""
		config.DefaultNpmRegistry = "foo.bar"

		npmUtils := newNpmMockUtilsBundle()
		npmUtils.execRunner = lintUtils.execRunner
		npmExecutor := npm.Execute{Utils: &npmUtils, Options: npm.ExecutorOptions{}}

		err := runNpmExecuteLint(&npmExecutor, &lintUtils, &config)

		if assert.EqualError(t, err, "Lint execution failed. This might be the result of severe linting findings, problems with the provided ESLint configuration (.eslintrc.json), or another issue. Please examine the linting results in the UI or in 0_defaultlint.xml, if available, or the log above. ") {
			if assert.Equal(t, 2, len(lintUtils.execRunner.Calls)) {
				assert.Equal(t, mock.ExecCall{Exec: "npx", Params: []string{
					"eslint",
					".",
					"-f", "stylish",
					"--ignore-pattern", "node_modules/",
					"--ignore-pattern", ".eslintrc.js",
					// no -o, --output-file in this case.
				}}, lintUtils.execRunner.Calls[1])
			}
		}
	})

	t.Run("Find ESLint configs", func(t *testing.T) {
		lintUtils := newLintMockUtilsBundle()
		lintUtils.AddFile("package.json", []byte("{\"name\": \"Test\" }"))
		lintUtils.AddFile(".eslintrc.json", []byte("{\"name\": \"Test\" }"))
		lintUtils.AddFile("src/.eslintrc.json", []byte("{\"name\": \"Test\" }"))
		lintUtils.AddFile("node_modules/.eslintrc.json", []byte("{\"name\": \"Test\" }")) // should be filtered out
		lintUtils.AddFile(".pipeline/.eslintrc.json", []byte("{\"name\": \"Test\" }"))    // should be filtered out

		eslintConfigs := findEslintConfigs(&lintUtils)
		if assert.Equal(t, 2, len(eslintConfigs)) {
			assert.Contains(t, eslintConfigs, ".eslintrc.json")
			assert.Contains(t, eslintConfigs, filepath.Join("src", ".eslintrc.json"))
		}
	})

	t.Run("Call with ci-lint script and install", func(t *testing.T) {
		lintUtils := newLintMockUtilsBundle()
		lintUtils.AddFile("package.json", []byte("{\"name\": \"test\", \"scripts\": { \"ci-lint\": \"\" } }"))

		npmUtils := npm.NewNpmMockUtilsBundle()
		npmUtils.ExecRunner = lintUtils.execRunner
		npmUtils.FilesMock = lintUtils.FilesMock

		config := defaultConfig
		config.Install = true

		npmExecutor := npm.NpmExecutorMock{Utils: npmUtils, Config: npm.NpmConfig{RunScripts: []string{"ci-lint"}, RunOptions: []string{"--silent"}, Install: true}}
		err := runNpmExecuteLint(&npmExecutor, &lintUtils, &config)

		assert.NoError(t, err)
	})

	t.Run("Call with default and install", func(t *testing.T) {
		lintUtils := newLintMockUtilsBundle()
		lintUtils.AddFile("package.json", []byte("{\"name\": \"test\"}"))

		npmUtils := npm.NewNpmMockUtilsBundle()
		npmUtils.ExecRunner = lintUtils.execRunner
		npmUtils.FilesMock = lintUtils.FilesMock

		config := defaultConfig
		config.Install = true

		npmExecutor := npm.NpmExecutorMock{Utils: npmUtils, Config: npm.NpmConfig{RunScripts: []string{"ci-lint"}, RunOptions: []string{"--silent"}, Install: true}}
		err := runNpmExecuteLint(&npmExecutor, &lintUtils, &config)

		assert.NoError(t, err)
	})

	t.Run("Call with custom runScript", func(t *testing.T) {
		lintUtils := newLintMockUtilsBundle()
		lintUtils.AddFile("package.json", []byte("{\"name\": \"test\", \"scripts\": { \"lint:ci\": \"\" } }"))

		npmUtils := npm.NewNpmMockUtilsBundle()
		npmUtils.ExecRunner = lintUtils.execRunner
		npmUtils.FilesMock = lintUtils.FilesMock

		config := defaultConfig
		config.RunScript = "lint:ci"

		npmExecutor := npm.NpmExecutorMock{Utils: npmUtils, Config: npm.NpmConfig{RunScripts: []string{"lint:ci"}, RunOptions: []string{"--silent"}}}
		err := runNpmExecuteLint(&npmExecutor, &lintUtils, &config)

		assert.NoError(t, err)
	})

	t.Run("Call with empty runScript and failOnError", func(t *testing.T) {
		lintUtils := newLintMockUtilsBundle()
		lintUtils.AddFile("package.json", []byte("{\"scripts\": { \"ci-lint\": \"\" } }"))
		lintUtils.execRunner = &mock.ExecMockRunner{ShouldFailOnCommand: map[string]error{"npm run ci-lint --silent": errors.New("exit 1")}}

		config := defaultConfig
		config.FailOnError = true
		config.RunScript = ""

		npmUtils := newNpmMockUtilsBundle()
		npmUtils.execRunner = lintUtils.execRunner
		npmUtils.FilesMock = lintUtils.FilesMock
		npmExecutor := npm.Execute{Utils: &npmUtils, Options: npm.ExecutorOptions{}}

		err := runNpmExecuteLint(&npmExecutor, &lintUtils, &config)

		assert.EqualError(t, err, "runScript is not allowed to be empty!")
	})

	t.Run("Test linter installation failed", func(t *testing.T) {
		lintUtils := newLintMockUtilsBundle()
		lintUtils.execRunner = &mock.ExecMockRunner{ShouldFailOnCommand: map[string]error{"npm install eslint@^7.0.0 typescript@^3.7.4 @typescript-eslint/parser@^3.0.0 @typescript-eslint/eslint-plugin@^3.0.0": errors.New("exit 1")}}

		npmUtils := newNpmMockUtilsBundle()
		npmUtils.execRunner = lintUtils.execRunner
		npmUtils.FilesMock = lintUtils.FilesMock

		config := defaultConfig
		config.FailOnError = true

		npmExecutor := npm.Execute{Utils: &npmUtils, Options: npm.ExecutorOptions{}}
		err := runNpmExecuteLint(&npmExecutor, &lintUtils, &config)

		assert.EqualError(t, err, "linter installation failed: exit 1")
	})

	t.Run("Test npx eslint fail", func(t *testing.T) {
		lintUtils := newLintMockUtilsBundle()
		lintUtils.execRunner = &mock.ExecMockRunner{ShouldFailOnCommand: map[string]error{"npx --no-install eslint . --ext .js,.jsx,.ts,.tsx -c .pipeline/.eslintrc.json -f checkstyle --ignore-pattern .eslintrc.js -o ./defaultlint.xml": errors.New("exit 1")}}

		npmUtils := newNpmMockUtilsBundle()
		npmUtils.execRunner = lintUtils.execRunner
		npmUtils.FilesMock = lintUtils.FilesMock

		config := defaultConfig
		config.FailOnError = true

		npmExecutor := npm.Execute{Utils: &npmUtils, Options: npm.ExecutorOptions{}}
		err := runNpmExecuteLint(&npmExecutor, &lintUtils, &config)

		assert.EqualError(t, err, "lint execution failed. This might be the result of severe linting findings. The lint configuration used can be found here: https://raw.githubusercontent.com/SAP/jenkins-library/master/resources/.eslintrc.json")
	})

}
