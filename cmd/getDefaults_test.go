package cmd

import (
	"io"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func defaultsOpenFileMock(name string, tokens map[string]string) (io.ReadCloser, error) {
	var r string
	switch name {
	case "TestAddCustomDefaults_default1":
		r = "default1"
	case "TestAddCustomDefaults_default2":
		r = "default3"
	default:
		r = ""
	}
	return ioutil.NopCloser(strings.NewReader(r)), nil
}

func TestDefaultsCommand(t *testing.T) {
	cmd := DefaultsCommand()

	gotReq := []string{}
	gotOpt := []string{}

	cmd.Flags().VisitAll(func(pflag *flag.Flag) {
		annotations, found := pflag.Annotations[cobra.BashCompOneRequiredFlag]
		if found && annotations[0] == "true" {
			gotReq = append(gotReq, pflag.Name)
		} else {
			gotOpt = append(gotOpt, pflag.Name)
		}
	})

	t.Run("Required flags", func(t *testing.T) {
		exp := []string{"defaultsFile"}
		assert.Equal(t, exp, gotReq, "required flags incorrect")
	})

	t.Run("Optional flags", func(t *testing.T) {
		exp := []string{"output", "outputFile"}
		assert.Equal(t, exp, gotOpt, "optional flags incorrect")
	})

	t.Run("Run", func(t *testing.T) {
		t.Run("Success case", func(t *testing.T) {
			defaultsOptions.openFile = defaultsOpenFileMock
			defaultsOptions.defaultsFiles = []string{"test", "test"}
			cmd.Run(cmd, []string{})
		})
	})
}

func TestGenerateDefaults(t *testing.T) {
	testParams := []struct {
		name          string
		defaultsFiles []string
		expected      string
	}{
		{
			name:          "Single defaults file",
			defaultsFiles: []string{"test"},
			expected:      `{"test":"general: null\nstages: null\nsteps: null\n"}`,
		},
		{
			name:          "Multiple defaults files",
			defaultsFiles: []string{"test1", "test2"},
			expected: `[{"test1":"general: null\nstages: null\nsteps: null\n"},` +
				`{"test2":"general: null\nstages: null\nsteps: null\n"}]`,
		},
	}

	utils := newGetDefaultsUtilsUtils()
	defaultsOptions.openFile = defaultsOpenFileMock

	for _, test := range testParams {
		t.Run(test.name, func(t *testing.T) {
			defaultsOptions.defaultsFiles = test.defaultsFiles
			result, _ := generateDefaults(utils)
			assert.Equal(t, test.expected, string(result))
		})
	}
}
