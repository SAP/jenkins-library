package kubernetes

// func TestRunUtils(t *testing.T) {
// 	t.Run("Get container info", func(t *testing.T) {
// 		testTable := []struct {
// 			config                HelmExecuteOptions
// 			expectedContainerInfo map[string]string
// 			expectedError         error
// 		}{
// 			{
// 				config: HelmExecuteOptions{
// 					Image:                "dtzar/helm-kubectl:3.4.1",
// 					ContainerImageName:   "",
// 					ContainerImageTag:    "",
// 					ContainerRegistryURL: "https://hub.docker.com/",
// 				},
// 				expectedContainerInfo: map[string]string{
// 					"containerImageName": "dtzar/helm-kubectl",
// 					"containerImageTag":  "3.4.1",
// 				},
// 				expectedError: nil,
// 			},
// 		}

// 		for _, testCase := range testTable {
// 			containerInfo, err := getContainerInfo(testCase.config)
// 			assert.NoError(t, err)
// 			assert.Equal(t, testCase.expectedContainerInfo["containerImageName"], containerInfo["containerImageName"])
// 			assert.Equal(t, testCase.expectedContainerInfo["containerImageTag"], containerInfo["containerImageTag"])

// 		}
// 	})
// }
