package kubernetes

import (
	"testing"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type helmMockUtilsBundle struct {
	*mock.FilesMock
	*mock.ExecMockRunner
}

// func (u *helmMockUtilsBundle) GetExecRunner() HelmDeployUtils {
// 	return u.execRunner
// }

func newHelmMockUtilsBundle() helmMockUtilsBundle {
	utils := helmMockUtilsBundle{ExecMockRunner: &mock.ExecMockRunner{}}
	return utils
}

func TestRunHelmLint(t *testing.T) {

	t.Run("Helm package command", func(t *testing.T) {
		t.Parallel()
		utils := newHelmMockUtilsBundle()

		config := HelmExecuteOptions{
			ChartPath:      ".",
			DeploymentName: "testPackage",
		}
		// runScripts := []string{"package", "."}

		err := RunHelmPackage(config, utils, log.Writer())

		if assert.NoError(t, err) {
			assert.Equal(t, mock.ExecCall{Exec: "helm", Params: []string{"package", "."}}, utils.Calls[0])
		}
	})

}
