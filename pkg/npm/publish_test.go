package npm

import (
	"io"
	"path/filepath"
	"slices"
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/piperutils"

	"github.com/SAP/jenkins-library/pkg/versioning"
	"github.com/stretchr/testify/assert"
)

type npmMockUtilsBundleRelativeGlob struct {
	*mock.FilesMockRelativeGlob
	execRunner     *mock.ExecMockRunner
	downloadClient piperhttp.Downloader
}

func (u *npmMockUtilsBundleRelativeGlob) GetExecRunner() ExecRunner {
	return u.execRunner
}

func (u *npmMockUtilsBundleRelativeGlob) GetFileUtils() piperutils.FileUtils {
	return u.FilesMock
}

func (u *npmMockUtilsBundleRelativeGlob) GetDownloadUtils() piperhttp.Downloader {
	return u.downloadClient
}

func newNpmMockUtilsBundleRelativeGlob() npmMockUtilsBundleRelativeGlob {
	return npmMockUtilsBundleRelativeGlob{
		FilesMockRelativeGlob: &mock.FilesMockRelativeGlob{FilesMock: &mock.FilesMock{}},
		execRunner:            &mock.ExecMockRunner{},
	}
}

func TestNpmPublish(t *testing.T) {
	type wants struct {
		publishConfigPath string
		publishConfig     string

		tarballPath string

		err string
	}

	tt := []struct {
		name string

		files map[string]string

		packageDescriptors []string
		registryURL        string
		registryUser       string
		registryPassword   string
		packBeforePublish  bool

		wants wants
	}{
		// project in root folder
		{
			name: "success - single project, publish normal, unpacked package - target registry in npmrc",

			files: map[string]string{
				"package.json": `{"name": "piper-project", "version": "0.0.1"}`,
				".piperNpmrc":  "_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nregistry=https://my.private.npm.registry/\n",
			},

			packageDescriptors: []string{"package.json"},

			wants: wants{
				publishConfigPath: `\.piperNpmrc`,
				publishConfig:     "_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nregistry=https://my.private.npm.registry/\n",
			},
		},
		{
			name: "success - single project, publish normal, unpacked package - target registry from pipeline",

			files: map[string]string{
				"package.json": `{"name": "piper-project", "version": "0.0.1"}`,
			},

			packageDescriptors: []string{"package.json"},

			registryURL:      "https://my.private.npm.registry/",
			registryUser:     "ThisIsTheUser",
			registryPassword: "AndHereIsThePassword",

			wants: wants{
				publishConfigPath: `\.piperNpmrc`,
				publishConfig:     "registry=https://my.private.npm.registry/\n//my.private.npm.registry/:_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nalways-auth=true\n",
			},
		},
		{
			name: "success - single project, publish normal, unpacked package - target registry from pipeline (precedence over npmrc)",

			files: map[string]string{
				"package.json": `{"name": "piper-project", "version": "0.0.1"}`,
				".piperNpmrc":  "//my.private.npm.registry/:_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nregistry=https://my.private.npm.registry/\n",
			},

			packageDescriptors: []string{"package.json"},

			registryURL:      "https://my.other.private.npm.registry/",
			registryUser:     "ThisIsTheOtherUser",
			registryPassword: "AndHereIsTheOtherPassword",

			wants: wants{
				publishConfigPath: `\.piperNpmrc`,
				publishConfig:     "//my.private.npm.registry/:_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nregistry=https://my.other.private.npm.registry/\n//my.other.private.npm.registry/:_auth=VGhpc0lzVGhlT3RoZXJVc2VyOkFuZEhlcmVJc1RoZU90aGVyUGFzc3dvcmQ=\nalways-auth=true\n",
			},
		},
		{
			name: "success - single project, publish normal, packed - target registry in npmrc",

			files: map[string]string{
				"package.json": `{"name": "piper-project", "version": "0.0.1"}`,
				".piperNpmrc":  "_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nregistry=https://my.private.npm.registry/\n",
			},

			packageDescriptors: []string{"package.json"},

			packBeforePublish: true,

			wants: wants{
				publishConfigPath: `\.piperNpmrc`,
				publishConfig:     "_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nregistry=https://my.private.npm.registry/\n",
				tarballPath:       "/package.tgz",
			},
		},
		{
			name: "success - single project, publish normal, packed - target registry from pipeline",

			files: map[string]string{
				"package.json": `{"name": "piper-project", "version": "0.0.1"}`,
			},

			packageDescriptors: []string{"package.json"},

			packBeforePublish: true,

			registryURL:      "https://my.private.npm.registry/",
			registryUser:     "ThisIsTheUser",
			registryPassword: "AndHereIsThePassword",

			wants: wants{
				publishConfigPath: `\.piperNpmrc`,
				publishConfig:     "registry=https://my.private.npm.registry/\n//my.private.npm.registry/:_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nalways-auth=true\n",
				tarballPath:       "/package.tgz",
			},
		},
		{
			name: "success - single project, publish normal, packed - target registry from pipeline (precedence over npmrc)",

			files: map[string]string{
				"package.json": `{"name": "piper-project", "version": "0.0.1"}`,
				".piperNpmrc":  "//my.private.npm.registry/:_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nregistry=https://my.private.npm.registry/\n",
			},

			packageDescriptors: []string{"package.json"},

			packBeforePublish: true,

			registryURL:      "https://my.other.private.npm.registry/",
			registryUser:     "ThisIsTheOtherUser",
			registryPassword: "AndHereIsTheOtherPassword",

			wants: wants{
				publishConfigPath: `\.piperNpmrc`,
				publishConfig:     "//my.private.npm.registry/:_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nregistry=https://my.other.private.npm.registry/\n//my.other.private.npm.registry/:_auth=VGhpc0lzVGhlT3RoZXJVc2VyOkFuZEhlcmVJc1RoZU90aGVyUGFzc3dvcmQ=\nalways-auth=true\n",
				tarballPath:       "/package.tgz",
			},
		},
		// scoped project
		{
			name: "success - single project, publish scoped, unpacked package - target registry in npmrc",

			files: map[string]string{
				"package.json": `{"name": "@piper/project", "version": "0.0.1"}`,
				".piperNpmrc":  "_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nregistry=https://my.private.npm.registry/\n@piper:registry=https://my.private.npm.registry/\n",
			},

			packageDescriptors: []string{"package.json"},

			wants: wants{
				publishConfigPath: `\.piperNpmrc`,
				publishConfig:     "_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nregistry=https://my.private.npm.registry/\n@piper:registry=https://my.private.npm.registry/\n",
			},
		},
		{
			name: "success - single project, publish scoped, unpacked package - target registry from pipeline",

			files: map[string]string{
				"package.json": `{"name": "@piper/project", "version": "0.0.1"}`,
			},

			packageDescriptors: []string{"package.json"},

			registryURL:      "https://my.private.npm.registry/",
			registryUser:     "ThisIsTheUser",
			registryPassword: "AndHereIsThePassword",

			wants: wants{
				publishConfigPath: `\.piperNpmrc`,
				publishConfig:     "registry=https://my.private.npm.registry/\n@piper:registry=https://my.private.npm.registry/\n//my.private.npm.registry/:_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nalways-auth=true\n",
			},
		},
		{
			name: "success - single project, publish scoped, unpacked package - target registry from pipeline (precedence over npmrc)",

			files: map[string]string{
				"package.json": `{"name": "@piper/project", "version": "0.0.1"}`,
				".piperNpmrc":  "//my.private.npm.registry/:_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nregistry=https://my.private.npm.registry/\n",
			},

			packageDescriptors: []string{"package.json"},

			registryURL:      "https://my.other.private.npm.registry/",
			registryUser:     "ThisIsTheOtherUser",
			registryPassword: "AndHereIsTheOtherPassword",

			wants: wants{
				publishConfigPath: `\.piperNpmrc`,
				publishConfig:     "//my.private.npm.registry/:_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nregistry=https://my.other.private.npm.registry/\n@piper:registry=https://my.other.private.npm.registry/\n//my.other.private.npm.registry/:_auth=VGhpc0lzVGhlT3RoZXJVc2VyOkFuZEhlcmVJc1RoZU90aGVyUGFzc3dvcmQ=\nalways-auth=true\n",
			},
		},
		{
			name: "success - single project, publish scoped, packed - target registry in npmrc",

			files: map[string]string{
				"package.json": `{"name": "@piper/project", "version": "0.0.1"}`,
				".piperNpmrc":  "_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\n@piper:registry=https://my.private.npm.registry/\n",
			},

			packageDescriptors: []string{"package.json"},

			packBeforePublish: true,

			wants: wants{
				publishConfigPath: `\.piperNpmrc`,
				publishConfig:     "_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\n@piper:registry=https://my.private.npm.registry/\n",
				tarballPath:       "/package.tgz",
			},
		},
		{
			name: "success - single project, publish scoped, packed - target registry from pipeline",

			files: map[string]string{
				"package.json": `{"name": "@piper/project", "version": "0.0.1"}`,
			},

			packageDescriptors: []string{"package.json"},

			packBeforePublish: true,

			registryURL:      "https://my.private.npm.registry/",
			registryUser:     "ThisIsTheUser",
			registryPassword: "AndHereIsThePassword",

			wants: wants{
				publishConfigPath: `\.piperNpmrc`,
				publishConfig:     "registry=https://my.private.npm.registry/\n@piper:registry=https://my.private.npm.registry/\n//my.private.npm.registry/:_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nalways-auth=true\n",
				tarballPath:       "/package.tgz",
			},
		},
		{
			name: "success - single project, publish scoped, packed - target registry from pipeline (precedence over npmrc)",

			files: map[string]string{
				"package.json": `{"name": "@piper/project", "version": "0.0.1"}`,
				".piperNpmrc":  "//my.private.npm.registry/:_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nregistry=https://my.private.npm.registry/\n",
			},

			packageDescriptors: []string{"package.json"},

			packBeforePublish: true,

			registryURL:      "https://my.other.private.npm.registry/",
			registryUser:     "ThisIsTheOtherUser",
			registryPassword: "AndHereIsTheOtherPassword",

			wants: wants{
				publishConfigPath: `\.piperNpmrc`,
				publishConfig:     "//my.private.npm.registry/:_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nregistry=https://my.other.private.npm.registry/\n@piper:registry=https://my.other.private.npm.registry/\n//my.other.private.npm.registry/:_auth=VGhpc0lzVGhlT3RoZXJVc2VyOkFuZEhlcmVJc1RoZU90aGVyUGFzc3dvcmQ=\nalways-auth=true\n",
				tarballPath:       "/package.tgz",
			},
		},
		// project in a subfolder
		{
			name: "success - single project in subfolder, publish normal, unpacked package - target registry in npmrc",

			files: map[string]string{
				"sub/package.json": `{"name": "piper-project", "version": "0.0.1"}`,
				"sub/.piperNpmrc":  "_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nregistry=https://my.private.npm.registry/\n",
			},

			packageDescriptors: []string{"sub/package.json"},

			wants: wants{
				publishConfigPath: `sub/\.piperNpmrc`,
				publishConfig:     "_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nregistry=https://my.private.npm.registry/\n",
			},
		},
		{
			name: "success - single project in subfolder, publish normal, unpacked package - target registry from pipeline",

			files: map[string]string{
				"sub/package.json": `{"name": "piper-project", "version": "0.0.1"}`,
			},

			packageDescriptors: []string{"sub/package.json"},

			registryURL:      "https://my.private.npm.registry/",
			registryUser:     "ThisIsTheUser",
			registryPassword: "AndHereIsThePassword",

			wants: wants{
				publishConfigPath: `sub/\.piperNpmrc`,
				publishConfig:     "registry=https://my.private.npm.registry/\n//my.private.npm.registry/:_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nalways-auth=true\n",
			},
		},
		{
			name: "success - single project in subfolder, publish normal, unpacked package - target registry from pipeline (precedence over npmrc)",

			files: map[string]string{
				"sub/package.json": `{"name": "piper-project", "version": "0.0.1"}`,
				"sub/.piperNpmrc":  "//my.private.npm.registry/:_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nregistry=https://my.private.npm.registry/\n",
			},

			packageDescriptors: []string{"sub/package.json"},

			registryURL:      "https://my.other.private.npm.registry/",
			registryUser:     "ThisIsTheOtherUser",
			registryPassword: "AndHereIsTheOtherPassword",

			wants: wants{
				publishConfigPath: `sub/\.piperNpmrc`,
				publishConfig:     "//my.private.npm.registry/:_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nregistry=https://my.other.private.npm.registry/\n//my.other.private.npm.registry/:_auth=VGhpc0lzVGhlT3RoZXJVc2VyOkFuZEhlcmVJc1RoZU90aGVyUGFzc3dvcmQ=\nalways-auth=true\n",
			},
		},
		{
			name: "success - single project in subfolder, publish normal, packed - target registry in npmrc",

			files: map[string]string{
				"sub/package.json": `{"name": "piper-project", "version": "0.0.1"}`,
				"sub/.piperNpmrc":  "_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nregistry=https://my.private.npm.registry/\n",
			},

			packageDescriptors: []string{"sub/package.json"},

			packBeforePublish: true,

			wants: wants{
				publishConfigPath: `\.piperNpmrc`,
				publishConfig:     "_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nregistry=https://my.private.npm.registry/\n",
				tarballPath:       "/sub/package.tgz",
			},
		},
		{
			name: "success - single project in subfolder, publish normal, packed - target registry from pipeline",

			files: map[string]string{
				"sub/package.json": `{"name": "piper-project", "version": "0.0.1"}`,
			},

			packageDescriptors: []string{"sub/package.json"},

			packBeforePublish: true,

			registryURL:      "https://my.private.npm.registry/",
			registryUser:     "ThisIsTheUser",
			registryPassword: "AndHereIsThePassword",

			wants: wants{
				publishConfigPath: `\.piperNpmrc`,
				publishConfig:     "registry=https://my.private.npm.registry/\n//my.private.npm.registry/:_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nalways-auth=true\n",
				tarballPath:       "/sub/package.tgz",
			},
		},
		{
			name: "success - single project in subfolder, publish normal, packed - target registry from pipeline (precedence over npmrc)",

			files: map[string]string{
				"sub/package.json": `{"name": "piper-project", "version": "0.0.1"}`,
				"sub/.piperNpmrc":  "//my.private.npm.registry/:_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nregistry=https://my.private.npm.registry/\n",
			},

			packageDescriptors: []string{"sub/package.json"},

			packBeforePublish: true,

			registryURL:      "https://my.other.private.npm.registry/",
			registryUser:     "ThisIsTheOtherUser",
			registryPassword: "AndHereIsTheOtherPassword",

			wants: wants{
				publishConfigPath: `\.piperNpmrc`,
				publishConfig:     "//my.private.npm.registry/:_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nregistry=https://my.other.private.npm.registry/\n//my.other.private.npm.registry/:_auth=VGhpc0lzVGhlT3RoZXJVc2VyOkFuZEhlcmVJc1RoZU90aGVyUGFzc3dvcmQ=\nalways-auth=true\n",
				tarballPath:       "/sub/package.tgz",
			},
		},
		// scoped project in a subfolder
		{
			name: "success - single project in subfolder, publish scoped, unpacked package - target registry in npmrc",

			files: map[string]string{
				"sub/package.json": `{"name": "@piper/project", "version": "0.0.1"}`,
				"sub/.piperNpmrc":  "_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nregistry=https://my.private.npm.registry/\n@piper:registry=https://my.private.npm.registry/\n",
			},

			packageDescriptors: []string{"sub/package.json"},

			wants: wants{
				publishConfigPath: `\.piperNpmrc`,
				publishConfig:     "_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nregistry=https://my.private.npm.registry/\n@piper:registry=https://my.private.npm.registry/\n",
			},
		},
		{
			name: "success - single project in subfolder, publish scoped, unpacked package - target registry from pipeline",

			files: map[string]string{
				"sub/package.json": `{"name": "@piper/project", "version": "0.0.1"}`,
			},

			packageDescriptors: []string{"sub/package.json"},

			registryURL:      "https://my.private.npm.registry/",
			registryUser:     "ThisIsTheUser",
			registryPassword: "AndHereIsThePassword",

			wants: wants{
				publishConfigPath: `\.piperNpmrc`,
				publishConfig:     "registry=https://my.private.npm.registry/\n@piper:registry=https://my.private.npm.registry/\n//my.private.npm.registry/:_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nalways-auth=true\n",
			},
		},
		{
			name: "success - single project in subfolder, publish scoped, unpacked package - target registry from pipeline (precedence over npmrc)",

			files: map[string]string{
				"sub/package.json": `{"name": "@piper/project", "version": "0.0.1"}`,
				"sub/.piperNpmrc":  "//my.private.npm.registry/:_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nregistry=https://my.private.npm.registry/\n",
			},

			packageDescriptors: []string{"sub/package.json"},

			registryURL:      "https://my.other.private.npm.registry/",
			registryUser:     "ThisIsTheOtherUser",
			registryPassword: "AndHereIsTheOtherPassword",

			wants: wants{
				publishConfigPath: `\.piperNpmrc`,
				publishConfig:     "//my.private.npm.registry/:_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nregistry=https://my.other.private.npm.registry/\n@piper:registry=https://my.other.private.npm.registry/\n//my.other.private.npm.registry/:_auth=VGhpc0lzVGhlT3RoZXJVc2VyOkFuZEhlcmVJc1RoZU90aGVyUGFzc3dvcmQ=\nalways-auth=true\n",
			},
		},
		{
			name: "success - single project in subfolder, publish scoped, packed - target registry in npmrc",

			files: map[string]string{
				"sub/package.json": `{"name": "@piper/project", "version": "0.0.1"}`,
				"sub/.piperNpmrc":  "_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\n@piper:registry=https://my.private.npm.registry/\n",
			},

			packageDescriptors: []string{"sub/package.json"},

			packBeforePublish: true,

			wants: wants{
				publishConfigPath: `\.piperNpmrc`,
				publishConfig:     "_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\n@piper:registry=https://my.private.npm.registry/\n",
				tarballPath:       "/sub/package.tgz",
			},
		},
		{
			name: "success - single project in subfolder, publish scoped, packed - target registry from pipeline",

			files: map[string]string{
				"sub/package.json": `{"name": "@piper/project", "version": "0.0.1"}`,
			},

			packageDescriptors: []string{"sub/package.json"},

			packBeforePublish: true,

			registryURL:      "https://my.private.npm.registry/",
			registryUser:     "ThisIsTheUser",
			registryPassword: "AndHereIsThePassword",

			wants: wants{
				publishConfigPath: `\.piperNpmrc`,
				publishConfig:     "registry=https://my.private.npm.registry/\n@piper:registry=https://my.private.npm.registry/\n//my.private.npm.registry/:_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nalways-auth=true\n",
				tarballPath:       "/sub/package.tgz",
			},
		},
		{
			name: "success - single project in subfolder, publish scoped, packed - target registry from pipeline (precedence over npmrc)",

			files: map[string]string{
				"sub/package.json": `{"name": "@piper/project", "version": "0.0.1"}`,
				"sub/.piperNpmrc":  "_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nregistry=https://my.private.npm.registry/\n",
			},

			packageDescriptors: []string{"sub/package.json"},

			packBeforePublish: true,

			registryURL:      "https://my.other.private.npm.registry/",
			registryUser:     "ThisIsTheOtherUser",
			registryPassword: "AndHereIsTheOtherPassword",

			wants: wants{
				publishConfigPath: `\.piperNpmrc`,
				publishConfig:     "_auth=VGhpc0lzVGhlVXNlcjpBbmRIZXJlSXNUaGVQYXNzd29yZA==\nregistry=https://my.other.private.npm.registry/\n@piper:registry=https://my.other.private.npm.registry/\n//my.other.private.npm.registry/:_auth=VGhpc0lzVGhlT3RoZXJVc2VyOkFuZEhlcmVJc1RoZU90aGVyUGFzc3dvcmQ=\nalways-auth=true\n",
				tarballPath:       "/sub/package.tgz",
			},
		},
		// TODO multiple projects
	}

	for _, test := range tt {
		t.Run(test.name, func(t *testing.T) {
			utils := newNpmMockUtilsBundleRelativeGlob()
			for path, content := range test.files {
				utils.AddFile(path, []byte(content))
			}
			utils.Separator = string(filepath.Separator)

			exec := &Execute{
				Utils: &utils,
			}

			propertiesLoadFile = utils.FileRead
			propertiesWriteFile = utils.FileWrite
			writeIgnoreFile = utils.FileWrite

			// This stub simulates the behavior of npm pack and puts a tgz into the requested
			utils.execRunner.Stub = func(call string, stdoutReturn map[string]string, shouldFailOnCommand map[string]error, stdout io.Writer) error {
				// tgzTargetPath := filepath.Dir(test.packageDescriptors[0])
				utils.AddFile(filepath.Join(".", "package.tgz"), []byte("this is a tgz file"))
				return nil
			}

			coordinates := []versioning.Coordinates{}
			err := exec.PublishAllPackages(test.packageDescriptors, test.registryURL, test.registryUser, test.registryPassword, test.packBeforePublish, &coordinates)

			if len(test.wants.err) == 0 && assert.NoError(t, err) {
				if assert.NotEmpty(t, utils.execRunner.Calls) {
					// last call is expected to be npm publish
					publishCmd := utils.execRunner.Calls[len(utils.execRunner.Calls)-1]

					assert.Equal(t, "npm", publishCmd.Exec)
					assert.Equal(t, "publish", publishCmd.Params[0])

					if len(test.wants.tarballPath) > 0 && assert.Contains(t, publishCmd.Params, "--tarball") {
						tarballPath := publishCmd.Params[slices.Index(publishCmd.Params, "--tarball")+1]
						assert.Equal(t, test.wants.tarballPath, filepath.ToSlash(tarballPath))
					}

					if assert.Contains(t, publishCmd.Params, "--userconfig") {
						effectivePublishConfigPath := publishCmd.Params[slices.Index(publishCmd.Params, "--userconfig")+1]

						assert.Regexp(t, test.wants.publishConfigPath, filepath.ToSlash(effectivePublishConfigPath))

						if test.packBeforePublish {
							subPath := filepath.Dir(test.packageDescriptors[0])
							effectivePublishConfigPath = filepath.Join(subPath, effectivePublishConfigPath)
						}

						effectiveConfig, err := utils.FileRead(effectivePublishConfigPath)
						if assert.NoError(t, err) {
							assert.Equal(t, test.wants.publishConfig, string(effectiveConfig))
						}
					}
				}
			} else {
				assert.EqualError(t, err, test.wants.err)
			}
		})
	}
}
