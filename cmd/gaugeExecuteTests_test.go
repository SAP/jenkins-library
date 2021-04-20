package cmd

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type gaugeExecuteTestsMockUtils struct {
	*mock.ShellMockRunner
	*mock.FilesMock
}

func newGaugeExecuteTestsTestsUtils() gaugeExecuteTestsMockUtils {
	utils := gaugeExecuteTestsMockUtils{
		ShellMockRunner: &mock.ShellMockRunner{},
		FilesMock:       &mock.FilesMock{},
	}
	return utils
}

func TestRunGaugeExecuteTests(t *testing.T) {
	t.Parallel()

	t.Run("success case", func(t *testing.T) {
		t.Parallel()
		config := &gaugeExecuteTestsOptions{
			InstallCommand: "curl -SsL https://downloads.gauge.org/stable | sh -s -- --location=$HOME/bin/gauge",
			LanguageRunner: "java",
			RunCommand:     "gauge run",
		}
		expectedParams := `export HOME=${HOME:-$(pwd)}
if [ "$HOME" = "/" ]; then export HOME=$(pwd); fi
export PATH=$HOME/bin/gauge:$PATH
mkdir -p $HOME/bin/gauge
curl -SsL https://downloads.gauge.org/stable | sh -s -- --location=$HOME/bin/gauge
gauge install html-report
gauge install xml-report
gauge install java
gauge run`

		mockUtils := newGaugeExecuteTestsTestsUtils()

		err := runGaugeExecuteTests(config, nil, &mockUtils)
		assert.NoError(t, err)
		assert.Equal(t, expectedParams, mockUtils.Calls[0])
	})
}
