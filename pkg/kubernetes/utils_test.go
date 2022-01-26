package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunUtils(t *testing.T) {
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

// func TestSplitImageName(t *testing.T) {
// 	tt := []struct {
// 		in       string
// 		outImage string
// 		outTag   string
// 		outError error
// 	}{
// 		{in: "", outImage: "", outTag: "", outError: fmt.Errorf("Failed to split image name ''")},
// 		{in: "path/to/image", outImage: "path/to/image", outTag: "", outError: nil},
// 		{in: "path/to/image:tag", outImage: "path/to/image", outTag: "tag", outError: nil},
// 		{in: "https://my.registry.com/path/to/image:tag", outImage: "", outTag: "", outError: fmt.Errorf("Failed to split image name 'https://my.registry.com/path/to/image:tag'")},
// 	}
// 	for _, test := range tt {
// 		i, tag, err := SplitFullImageName(test.in)
// 		assert.Equal(t, test.outImage, i, "Image value unexpected")
// 		assert.Equal(t, test.outTag, tag, "Tag value unexpected")
// 		assert.Equal(t, test.outError, err, "Error value not as expected")
// 	}
// }
