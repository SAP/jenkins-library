package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"

	dockermock "github.com/SAP/jenkins-library/pkg/docker/mock"
	"github.com/SAP/jenkins-library/pkg/mock"
)

const (
	customDockerConfig = `{"auths":{"source.registry":{"auth":"c291cmNldXNlcjpzb3VyY2VwYXNzd29yZA=="},"target.registry":{"auth":"dGFyZ2V0dXNlcjp0YXJnZXRwYXNzd29yZA=="}}}`
	dockerConfig       = `{
	"auths": {
		"source.registry": {
			"auth": "c291cmNldXNlcjpzb3VyY2VwYXNzd29yZA=="
		},
		"target.registry": {
			"auth": "dGFyZ2V0dXNlcjp0YXJnZXRwYXNzd29yZA=="
		},
		"test.registry": {
			"auth": "dGVzdHVzZXI6dGVzdHBhc3N3b3Jk"
		}
	}
}`
)

type imagePushToRegistryMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
	*dockermock.CraneMockUtils
}

func newImagePushToRegistryMockUtils(craneUtils *dockermock.CraneMockUtils) *imagePushToRegistryMockUtils {
	utils := &imagePushToRegistryMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
		CraneMockUtils: craneUtils,
	}

	return utils
}

func TestRunImagePushToRegistry(t *testing.T) {
	t.Parallel()

	t.Run("good case", func(t *testing.T) {
		t.Parallel()

		config := imagePushToRegistryOptions{
			SourceRegistryURL:      "https://source.registry",
			SourceImages:           []string{"source-image:latest"},
			SourceRegistryUser:     "sourceuser",
			SourceRegistryPassword: "sourcepassword",
			TargetRegistryURL:      "https://target.registry",
			TargetImages:           []string{"target-image:latest"},
			TargetRegistryUser:     "targetuser",
			TargetRegistryPassword: "targetpassword",
		}
		craneMockUtils := &dockermock.CraneMockUtils{}
		utils := newImagePushToRegistryMockUtils(craneMockUtils)
		err := runImagePushToRegistry(&config, nil, utils)
		assert.NoError(t, err)
		createdConfig, err := utils.FileRead(targetDockerConfigPath)
		assert.NoError(t, err)
		assert.Equal(t, customDockerConfig, string(createdConfig))
	})

	t.Run("failed to copy image", func(t *testing.T) {
		t.Parallel()

		config := imagePushToRegistryOptions{
			SourceRegistryURL:      "https://source.registry",
			SourceRegistryUser:     "sourceuser",
			SourceRegistryPassword: "sourcepassword",
			SourceImages:           []string{"source-image:latest"},
			TargetRegistryURL:      "https://target.registry",
			TargetRegistryUser:     "targetuser",
			TargetRegistryPassword: "targetpassword",
		}
		craneMockUtils := &dockermock.CraneMockUtils{
			ErrCopyImage: dockermock.ErrCopyImage,
		}
		utils := newImagePushToRegistryMockUtils(craneMockUtils)
		err := runImagePushToRegistry(&config, nil, utils)
		assert.EqualError(t, err, "failed to copy images: copy image err")
	})

	t.Run("failed to push local image", func(t *testing.T) {
		t.Parallel()

		config := imagePushToRegistryOptions{
			SourceRegistryURL:      "https://source.registry",
			SourceRegistryUser:     "sourceuser",
			SourceRegistryPassword: "sourcepassword",
			SourceImages:           []string{"source-image:latest"},
			TargetRegistryURL:      "https://target.registry",
			TargetRegistryUser:     "targetuser",
			TargetRegistryPassword: "targetpassword",
			LocalDockerImagePath:   "/local/path",
		}
		craneMockUtils := &dockermock.CraneMockUtils{
			ErrLoadImage: dockermock.ErrLoadImage,
		}
		utils := newImagePushToRegistryMockUtils(craneMockUtils)
		err := runImagePushToRegistry(&config, nil, utils)
		assert.EqualError(t, err, "failed to push local image to \"target.registry\": load image err")
	})
}

func TestHandleCredentialsForPrivateRegistry(t *testing.T) {
	t.Parallel()

	craneMockUtils := &dockermock.CraneMockUtils{}
	t.Run("no custom docker config provided", func(t *testing.T) {
		t.Parallel()

		utils := newImagePushToRegistryMockUtils(craneMockUtils)
		utils.AddFile("targetDockerConfigPath", []byte("abc"))
		err := handleCredentialsForPrivateRegistry("", "target.registry", "targetuser", "targetpassword", utils)
		assert.NoError(t, err)
		createdConfigFile, err := utils.FileRead(targetDockerConfigPath)
		assert.NoError(t, err)
		assert.Equal(t, `{"auths":{"target.registry":{"auth":"dGFyZ2V0dXNlcjp0YXJnZXRwYXNzd29yZA=="}}}`, string(createdConfigFile))
	})

	t.Run("custom docker config provided", func(t *testing.T) {
		t.Parallel()

		utils := newImagePushToRegistryMockUtils(craneMockUtils)
		utils.AddFile(targetDockerConfigPath, []byte(customDockerConfig))
		err := handleCredentialsForPrivateRegistry(targetDockerConfigPath, "test.registry", "testuser", "testpassword", utils)
		assert.NoError(t, err)
		createdConfigFile, err := utils.FileRead(targetDockerConfigPath)
		assert.NoError(t, err)
		assert.Equal(t, dockerConfig, string(createdConfigFile))
	})

	t.Run("wrong format of docker config", func(t *testing.T) {
		t.Parallel()

		utils := newImagePushToRegistryMockUtils(craneMockUtils)
		utils.AddFile(targetDockerConfigPath, []byte(`{auths:}`))
		err := handleCredentialsForPrivateRegistry("", "test.registry", "testuser", "testpassword", utils)
		assert.EqualError(t, err, "failed to create new docker config: failed to unmarshal json file '/root/.docker/config.json': invalid character 'a' looking for beginning of object key string")
	})
}

func TestPushLocalImageToTargetRegistry(t *testing.T) {
	t.Parallel()
	t.Run("good case", func(t *testing.T) {
		t.Parallel()

		craneMockUtils := &dockermock.CraneMockUtils{}
		config := &imagePushToRegistryOptions{
			LocalDockerImagePath: "/image/path",
			TargetRegistryURL:    "https://target.registry",
			TagLatest:            false,
		}
		utils := newImagePushToRegistryMockUtils(craneMockUtils)
		err := pushLocalImageToTargetRegistry(config, utils)
		assert.NoError(t, err)
	})

	t.Run("bad case - failed to load image", func(t *testing.T) {
		t.Parallel()

		craneMockUtils := &dockermock.CraneMockUtils{
			ErrLoadImage: dockermock.ErrLoadImage,
		}
		config := &imagePushToRegistryOptions{
			LocalDockerImagePath: "/image/path",
			TargetRegistryURL:    "https://target.registry",
			TagLatest:            false,
		}
		utils := newImagePushToRegistryMockUtils(craneMockUtils)
		err := pushLocalImageToTargetRegistry(config, utils)
		assert.EqualError(t, err, "load image err")
	})

	t.Run("bad case - failed to push image", func(t *testing.T) {
		t.Parallel()

		craneMockUtils := &dockermock.CraneMockUtils{
			ErrPushImage: dockermock.ErrPushImage,
		}
		config := &imagePushToRegistryOptions{
			LocalDockerImagePath: "/image/path",
			TargetRegistryURL:    "https://target.registry",
			TargetImages:         []string{"my-image:1.0.0"},
			TagLatest:            false,
		}
		utils := newImagePushToRegistryMockUtils(craneMockUtils)
		err := pushLocalImageToTargetRegistry(config, utils)
		assert.EqualError(t, err, "push image err")
	})
}

func TestParseDockerImageName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, image, expected string
	}{
		{
			name:     "registry + imagename + tag",
			image:    "test.io/repo/test-image:1.0.0-12345",
			expected: "test.io/repo/test-image",
		},
		{
			name:     "registry + imagename + tag (registry with port)",
			image:    "test.io:50000/repo/test-image:1.0.0-12345",
			expected: "test.io:50000/repo/test-image",
		},
		{
			name:     "registry + imagename",
			image:    "test-test.io/repo/testimage",
			expected: "test-test.io/repo/testimage",
		},
		{
			name:     "imagename + tag",
			image:    "testImage:1.0.0",
			expected: "testImage",
		},
		{
			name:     "imagename",
			image:    "test-image",
			expected: "test-image",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			actual := parseDockerImageName(test.image)
			assert.Equal(t, test.expected, actual)
		})
	}
}
