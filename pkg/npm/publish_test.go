package npm

import (
	"io"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/magiconair/properties"
	"github.com/stretchr/testify/assert"
)

func TestNpmPublish(t *testing.T) {
	tt := []struct {
		name string

		files map[string]string

		packageDescriptors []string
		registryURL        string
		registryUser       string
		registryPassword   string
		packBeforePublish  bool

		expectedPublishConfigPath string
		expectedPublishConfig     string
		expectedError             string
	}{
		{
			name: "success - single project, publish normal, unpacked package - target registry in npmrc",

			files: map[string]string{
				"package.json": `{"name": "piper-project", "version": "0.0.1"}`,
				".piperNpmrc":  "_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nregistry=https://my.private.npm.registry/",
			},

			packageDescriptors: []string{"package.json"},

			expectedPublishConfigPath: `\.piperNpmrc`,
			expectedPublishConfig:     "_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nregistry=https://my.private.npm.registry/",
		},
		{
			name: "success - single project, publish normal, unpacked package - target registry from pipeline",

			files: map[string]string{
				"package.json": `{"name": "piper-project", "version": "0.0.1"}`,
			},

			packageDescriptors: []string{"package.json"},

			expectedPublishConfigPath: `\.piperNpmrc`,
			expectedPublishConfig:     "registry = https://my.private.npm.registry/\n_auth = VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nalways-auth = true\n",

			registryURL:      "https://my.private.npm.registry/",
			registryUser:     "ThisIsTheUser",
			registryPassword: "AndHereIsThePassword",
		},
		{
			name: "success - single project, publish normal, unpacked package - target registry from pipeline (precedence over npmrc)",

			files: map[string]string{
				"package.json": `{"name": "piper-project", "version": "0.0.1"}`,
				".piperNpmrc":  "_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nregistry=https://my.private.npm.registry/",
			},

			packageDescriptors: []string{"package.json"},

			expectedPublishConfigPath: `\.piperNpmrc`,
			expectedPublishConfig:     "_auth = VGhpc0lzVGhlT3RoZXJVc2VyOkFuZEhlcmVJc1RoZU90aGVyUGFzc3dvcmQ=\nregistry = https://my.other.private.npm.registry/\nalways-auth = true\n",

			registryURL:      "https://my.other.private.npm.registry/",
			registryUser:     "ThisIsTheOtherUser",
			registryPassword: "AndHereIsTheOtherPassword",
		},
		{
			name: "success - single project, publish normal, packed - target registry in npmrc",

			files: map[string]string{
				"package.json": `{"name": "piper-project", "version": "0.0.1"}`,
				".piperNpmrc":  "_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nregistry=https://my.private.npm.registry/",
			},

			packageDescriptors: []string{"package.json"},

			packBeforePublish: true,

			expectedPublishConfigPath: `temp-(?:test|[0-9]+)/\.piperNpmrc`,
			expectedPublishConfig:     "_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nregistry=https://my.private.npm.registry/",
		},
		{
			name: "success - single project, publish normal, packed - target registry from pipeline",

			files: map[string]string{
				"package.json": `{"name": "piper-project", "version": "0.0.1"}`,
			},

			packageDescriptors: []string{"package.json"},

			packBeforePublish: true,

			expectedPublishConfigPath: `temp-(?:test|[0-9]+)/\.piperNpmrc`,
			expectedPublishConfig:     "registry = https://my.private.npm.registry/\n_auth = VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nalways-auth = true\n",

			registryURL:      "https://my.private.npm.registry/",
			registryUser:     "ThisIsTheUser",
			registryPassword: "AndHereIsThePassword",
		},
		{
			name: "success - single project, publish normal, packed - target registry from pipeline",

			files: map[string]string{
				"package.json": `{"name": "piper-project", "version": "0.0.1"}`,
				".piperNpmrc":  "_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nregistry=https://my.private.npm.registry/",
			},

			packageDescriptors: []string{"package.json"},

			packBeforePublish: true,

			expectedPublishConfigPath: `temp-(?:test|[0-9]+)/\.piperNpmrc`,
			expectedPublishConfig:     "_auth = VGhpc0lzVGhlT3RoZXJVc2VyOkFuZEhlcmVJc1RoZU90aGVyUGFzc3dvcmQ=\nregistry = https://my.other.private.npm.registry/\nalways-auth = true\n",

			registryURL:      "https://my.other.private.npm.registry/",
			registryUser:     "ThisIsTheOtherUser",
			registryPassword: "AndHereIsTheOtherPassword",
		},
		/*{
			name: "success - publish scoped, packed - target registry in npmrc",

			files: map[string]string{
				"package.json": `{"name": "@piper/project", "version": "0.0.1"}`,
				".piperNpmrc":  testNpmrc.String(),
			},

			packageDescriptors: []string{"package.json"},

			packBeforePublish: true,

			expectedPublishConfigPath: `temp-(?:test|[0-9]+)/\.piperNpmrc`,
			expectedPublishConfig:     testNpmrc,
		},
		{
			name: "success - publish scoped, packed - target registry from pipeline",

			files: map[string]string{
				"package.json": `{"name": "@piper/project", "version": "0.0.1"}`,
				".piperNpmrc":  testNpmrc.String(),
			},

			packageDescriptors: []string{"package.json"},

			packBeforePublish: true,

			expectedPublishConfigPath: `temp-(?:test|[0-9]+)/\.piperNpmrc`,
			expectedPublishConfig:     testNpmrc,

			registryURL:      "https://my.private.npm.registry/",
			registryUser:     "ThisIsTheUser",
			registryPassword: "AndHereIsThePassword",
		},*/
	}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			utils := newNpmMockUtilsBundle()

			for path, content := range test.files {
				utils.AddFile(path, []byte(content))
			}

			options := ExecutorOptions{}

			exec := &Execute{
				Utils:   &utils,
				Options: options,
			}

			propertiesLoadFile = func(filename string, enc properties.Encoding) (*properties.Properties, error) {
				p := properties.NewProperties()

				b, err := utils.FileRead(filename)

				if err != nil {
					return nil, err
				}

				err = p.Load(b, properties.UTF8)
				return p, err
			}

			propertiesWriteFile = utils.FileWrite
			writeIgnoreFile = utils.FileWrite

			// This stub simulates the behavior of npm pack and puts a tgz into the requested
			utils.execRunner.Stub = func(call string, stdoutReturn map[string]string, shouldFailOnCommand map[string]error, stdout io.Writer) error {
				r := regexp.MustCompile(`npm\s+pack\s+.*--pack-destination\s+(?P<destination>[^\s]+).*`)

				matches := r.FindStringSubmatch(call)

				if len(matches) == 0 {
					return nil
				}

				packDestination := matches[1]

				utils.AddFile(filepath.Join(packDestination, "package.tgz"), []byte("this is a tgz file"))

				return nil
			}

			err := exec.PublishAllPackages(test.packageDescriptors, test.registryURL, test.registryUser, test.registryPassword, test.packBeforePublish)

			if len(test.expectedError) == 0 && assert.NoError(t, err) {
				if assert.NotEmpty(t, utils.execRunner.Calls) {
					// last call is expected to be npm publish
					publishCmd := utils.execRunner.Calls[len(utils.execRunner.Calls)-1]

					assert.Equal(t, "npm", publishCmd.Exec)
					assert.Equal(t, "publish", publishCmd.Params[0])

					if assert.Contains(t, publishCmd.Params, "--userconfig") {
						effectivePublishConfigPath := publishCmd.Params[piperutils.FindString(publishCmd.Params, "--userconfig")+1]

						assert.Regexp(t, test.expectedPublishConfigPath, effectivePublishConfigPath)

						effectiveConfig, err := utils.FileRead(effectivePublishConfigPath)

						if assert.NoError(t, err) {
							assert.Equal(t, test.expectedPublishConfig, string(effectiveConfig))
						}
					}
				}
			} else {
				assert.EqualError(t, err, test.expectedError)
			}
		})
	}
}
