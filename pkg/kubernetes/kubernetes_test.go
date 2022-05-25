package kubernetes

import "github.com/SAP/jenkins-library/pkg/mock"

type kubernetesDeployMockUtils struct {
	shouldFail     bool
	requestedUrls  []string
	requestedFiles []string
	*mock.FilesMock
	*mock.ExecMockRunner
	*mock.HttpClientMock
}

func newKubernetesDeployMockUtils() kubernetesDeployMockUtils {
	utils := kubernetesDeployMockUtils{
		shouldFail:     false,
		FilesMock:      &mock.FilesMock{},
		ExecMockRunner: &mock.ExecMockRunner{},
	}
	return utils
}
