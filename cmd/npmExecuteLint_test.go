package cmd

import (
	"errors"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/npm"
	"github.com/bmatcuk/doublestar"
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

type npmMockUtilsBundle struct {
	execRunner mock.ExecMockRunner
	files      map[string]string
}

func (u *npmMockUtilsBundle) FileExists(path string) (bool, error) {
	_, exists := u.files[path]
	return exists, nil
}

func (u *npmMockUtilsBundle) FileRead(path string) ([]byte, error) {
	return []byte(u.files[path]), nil
}

// duplicated from nexusUpload_test.go for now, refactor later?
func (u *npmMockUtilsBundle) Glob(pattern string) ([]string, error) {
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

func (u *npmMockUtilsBundle) Getwd() (dir string, err error) {
	return "/project", nil
}

func (u *npmMockUtilsBundle) Chdir(dir string) error {
	return nil
}

func (u *npmMockUtilsBundle) GetExecRunner() npm.ExecRunner {
	return &u.execRunner
}

func TestNpmExecuteLint(t *testing.T) {
	t.Run("Call with ci-lint script and one package.json", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.files["package.json"] = "{\"scripts\": { \"ci-lint\": \"\" } }"
		options := npmExecuteLintOptions{}
		options.DefaultNpmRegistry = "foo.bar"

		err := runNpmExecuteLint(&utils, &options)

		assert.NoError(t, err)
		assert.Equal(t, 3, len(utils.execRunner.Calls))
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "ci-lint", "--silent"}}, utils.execRunner.Calls[2])
	})

	t.Run("Call with ci-lint script and two package.json", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.files["package.json"] = "{\"scripts\": { \"ci-lint\": \"\" } }"
		utils.files["src/package.json"] = "{\"scripts\": { \"ci-lint\": \"\" } }"
		options := npmExecuteLintOptions{}
		options.DefaultNpmRegistry = "foo.bar"

		err := runNpmExecuteLint(&utils, &options)

		assert.NoError(t, err)
		assert.Equal(t, 6, len(utils.execRunner.Calls))
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "ci-lint", "--silent"}}, utils.execRunner.Calls[2])
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "ci-lint", "--silent"}}, utils.execRunner.Calls[5])
	})

	t.Run("Call default with ESLint config from user", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.files["package.json"] = "{\"name\": \"Test\" }"
		utils.files[".eslintrc.json"] = "{\"name\": \"Test\" }"
		options := npmExecuteLintOptions{}
		options.DefaultNpmRegistry = "foo.bar"

		err := runNpmExecuteLint(&utils, &options)

		assert.NoError(t, err)
		assert.Equal(t, mock.ExecCall{Exec: "npx", Params: []string{"eslint", ".", "-f", "checkstyle", "-o", "./0_defaultlint.xml", "--ignore-pattern", "node_modules/", "--ignore-pattern", ".eslintrc.js"}}, utils.execRunner.Calls[2])
		assert.Equal(t, 3, len(utils.execRunner.Calls))
	})

	t.Run("Call default with two ESLint configs from user", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.files["package.json"] = "{\"name\": \"Test\" }"
		utils.files[".eslintrc.json"] = "{\"name\": \"Test\" }"
		utils.files["src/.eslintrc.json"] = "{\"name\": \"Test\" }"
		options := npmExecuteLintOptions{}
		options.DefaultNpmRegistry = "foo.bar"

		err := runNpmExecuteLint(&utils, &options)

		assert.NoError(t, err)
		assert.Equal(t, mock.ExecCall{Exec: "npx", Params: []string{"eslint", ".", "-f", "checkstyle", "-o", "./0_defaultlint.xml", "--ignore-pattern", "node_modules/", "--ignore-pattern", ".eslintrc.js"}}, utils.execRunner.Calls[2])
		assert.Equal(t, mock.ExecCall{Exec: "npx", Params: []string{"eslint", "src/**/*.js", "-f", "checkstyle", "-o", "./1_defaultlint.xml", "--ignore-pattern", "node_modules/", "--ignore-pattern", ".eslintrc.js"}}, utils.execRunner.Calls[3])
		assert.Equal(t, 4, len(utils.execRunner.Calls))
	})

	t.Run("Default without ESLint config", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.files["package.json"] = "{\"name\": \"Test\" }"
		options := npmExecuteLintOptions{}
		options.DefaultNpmRegistry = "foo.bar"

		err := runNpmExecuteLint(&utils, &options)

		assert.NoError(t, err)
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"install", "eslint@^7.0.0", "typescript@^3.7.4", "@typescript-eslint/parser@^3.0.0", "@typescript-eslint/eslint-plugin@^3.0.0"}}, utils.execRunner.Calls[2])
		assert.Equal(t, mock.ExecCall{Exec: "npx", Params: []string{"--no-install", "eslint", ".", "--ext", ".js,.jsx,.ts,.tsx", "-c", ".pipeline/.eslintrc.json", "-f", "checkstyle", "-o", "./defaultlint.xml", "--ignore-pattern", ".eslintrc.js"}}, utils.execRunner.Calls[3])
		assert.Equal(t, 4, len(utils.execRunner.Calls))
	})
	t.Run("Ignore ESLint config in node_modules and .pipeline/", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.files["package.json"] = "{\"name\": \"Test\" }"
		utils.files["node_modules/.eslintrc.json"] = "{\"name\": \"Test\" }"
		utils.files[".pipeline/.eslintrc.json"] = "{\"name\": \"Test\" }"
		options := npmExecuteLintOptions{}
		options.DefaultNpmRegistry = "foo.bar"

		err := runNpmExecuteLint(&utils, &options)

		assert.NoError(t, err)
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"install", "eslint@^7.0.0", "typescript@^3.7.4", "@typescript-eslint/parser@^3.0.0", "@typescript-eslint/eslint-plugin@^3.0.0"}}, utils.execRunner.Calls[2])
		assert.Equal(t, mock.ExecCall{Exec: "npx", Params: []string{"--no-install", "eslint", ".", "--ext", ".js,.jsx,.ts,.tsx", "-c", ".pipeline/.eslintrc.json", "-f", "checkstyle", "-o", "./defaultlint.xml", "--ignore-pattern", ".eslintrc.js"}}, utils.execRunner.Calls[3])
		assert.Equal(t, 4, len(utils.execRunner.Calls))
	})

	t.Run("Call with ci-lint script and failOnError", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.files["package.json"] = "{\"scripts\": { \"ci-lint\": \"\" } }"
		utils.execRunner = mock.ExecMockRunner{ShouldFailOnCommand: map[string]error{"npm run ci-lint --silent": errors.New("exit 1")}}
		options := npmExecuteLintOptions{}
		options.FailOnError = true
		options.DefaultNpmRegistry = "foo.bar"

		err := runNpmExecuteLint(&utils, &options)

		assert.EqualError(t, err, "failed to run npm script ci-lint: exit 1")
		assert.Equal(t, 3, len(utils.execRunner.Calls))
		assert.Equal(t, mock.ExecCall{Exec: "npm", Params: []string{"run", "ci-lint", "--silent"}}, utils.execRunner.Calls[2])
	})

	t.Run("Call default with ESLint config from user and failOnError", func(t *testing.T) {
		utils := newNpmMockUtilsBundle()
		utils.files["package.json"] = "{\"name\": \"Test\" }"
		utils.files[".eslintrc.json"] = "{\"name\": \"Test\" }"
		utils.execRunner = mock.ExecMockRunner{ShouldFailOnCommand: map[string]error{"eslint . -f checkstyle -o ./0_defaultlint.xml --ignore-pattern node_modules/ --ignore-pattern .eslintrc.js": errors.New("exit 1")}}
		options := npmExecuteLintOptions{}
		options.FailOnError = true
		options.DefaultNpmRegistry = "foo.bar"

		err := runNpmExecuteLint(&utils, &options)

		assert.EqualError(t, err, "failed to run ESLint with config .eslintrc.json: exit 1")
		assert.Equal(t, mock.ExecCall{Exec: "npx", Params: []string{"eslint", ".", "-f", "checkstyle", "-o", "./0_defaultlint.xml", "--ignore-pattern", "node_modules/", "--ignore-pattern", ".eslintrc.js"}}, utils.execRunner.Calls[2])
		assert.Equal(t, 3, len(utils.execRunner.Calls))
	})

}

func newNpmMockUtilsBundle() npmMockUtilsBundle {
	utils := npmMockUtilsBundle{}
	utils.files = map[string]string{}
	return utils
}
