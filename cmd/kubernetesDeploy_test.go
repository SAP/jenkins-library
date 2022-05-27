package cmd

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/kubernetes"
	"github.com/SAP/jenkins-library/pkg/kubernetes/mocks"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/stretchr/testify/assert"
)

func TestRunKubernetesDeploy(t *testing.T) {
	config := kubernetes.KubernetesOptions{}
	tt := []struct {
		deployTool  string
		methodName  string
		expectedErr error
	}{
		{"helm", "RunHelmDeploy", fmt.Errorf("RunHelmDeploy method successfully finished")},
		{"helm3", "RunHelmDeploy", fmt.Errorf("RunHelmDeploy method successfully finished")},
		{"kubectl", "RunKubectlDeploy", fmt.Errorf("RunKubectlDeploy method successfully finished")},
		{"", "", fmt.Errorf("Failed to execute deployments")},
	}

	for i, test := range tt {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			config.DeployTool = test.deployTool
			kubernetesDeploy := &mocks.KubernetesDeploy{}
			kubernetesDeploy.On(test.methodName).Return(test.expectedErr)
			err := runKubernetesDeploy(config, &telemetry.CustomData{}, kubernetesDeploy)
			assert.Equal(t, test.expectedErr, err)
		})
	}
}
