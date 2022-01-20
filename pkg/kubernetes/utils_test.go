package kubernetes

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunUtils(t *testing.T) {
	t.Run("Split full image name", func(t *testing.T) {
		var err error
		testTable := []struct {
			testInput             map[string]string
			expectedContainerInfo map[string]string
			expectedError         error
		}{
			{
				testInput: map[string]string{
					"image":     "dtzar/helm-kubectl:3.4.1",
					"imageName": "",
					"tag":       "",
				},
				expectedContainerInfo: map[string]string{
					"containerImageName": "dtzar/helm-kubectl",
					"containerImageTag":  "3.4.1",
				},
				expectedError: nil,
			},
			{
				testInput: map[string]string{
					"image":     "dtzar",
					"imageName": "",
					"tag":       "",
				},
				expectedContainerInfo: map[string]string{
					"containerImageName": "dtzar",
					"containerImageTag":  "",
				},
				expectedError: nil,
			},
			{
				testInput: map[string]string{
					"image":     "",
					"imageName": "",
					"tag":       "",
				},
				expectedContainerInfo: map[string]string{
					"containerImageName": "dtzar",
					"containerImageTag":  "",
				},
				expectedError: errors.New("failed to split image name ''"),
			},
		}

		for _, testCase := range testTable {
			testCase.testInput["imageName"], testCase.testInput["tag"], err = splitFullImageName(testCase.testInput["image"])
			if testCase.expectedError == nil {
				assert.NoError(t, err)
				assert.Equal(t, testCase.expectedContainerInfo["containerImageName"], testCase.testInput["imageName"])
				assert.Equal(t, testCase.expectedContainerInfo["containerImageTag"], testCase.testInput["tag"])
			} else {
				assert.Error(t, err)
				assert.Equal(t, testCase.expectedError, err)
			}
		}
	})

	t.Run("Get container info", func(t *testing.T) {
		testTable := []struct {
			config                HelmExecuteOptions
			expectedContainerInfo map[string]string
			expectedError         error
		}{
			{
				config: HelmExecuteOptions{
					Image:                "dtzar/helm-kubectl:3.4.1",
					ContainerImageName:   "",
					ContainerImageTag:    "",
					ContainerRegistryURL: "https://hub.docker.com/",
				},
				expectedContainerInfo: map[string]string{
					"containerImageName": "dtzar/helm-kubectl",
					"containerImageTag":  "3.4.1",
				},
				expectedError: nil,
			},
		}

		for _, testCase := range testTable {
			containerInfo, err := getContainerInfo(testCase.config)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedContainerInfo["containerImageName"], containerInfo["containerImageName"])
			assert.Equal(t, testCase.expectedContainerInfo["containerImageTag"], containerInfo["containerImageTag"])

		}
	})
}
