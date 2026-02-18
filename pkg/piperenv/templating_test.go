//go:build unit

package piperenv

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseTemplate(t *testing.T) {
	tt := []struct {
		template      string
		cpe           CPEMap
		expected      string
		expectedError error
	}{
		{template: `version: {{index .CPE "artifactVersion"}}, sha: {{git "commitId"}}`, expected: "version: 1.2.3, sha: thisIsMyTestSha"},
		{template: "version: {{", expectedError: fmt.Errorf("failed to parse cpe template 'version: {{'")},
	}

	cpe := CPEMap{
		"artifactVersion": "1.2.3",
		"git/commitId":    "thisIsMyTestSha",
	}

	for _, test := range tt {
		res, err := cpe.ParseTemplate(test.template)
		if test.expectedError != nil {
			assert.Contains(t, fmt.Sprint(err), fmt.Sprint(test.expectedError))
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.expected, (*res).String())
		}

	}
}

func TestParseTemplateWithDelimiter(t *testing.T) {
	tt := []struct {
		template      string
		cpe           CPEMap
		expected      string
		expectedError error
	}{
		{template: `version: [[index .CPE "artifactVersion"]], sha: [[git "commitId"]]`, expected: "version: 1.2.3, sha: thisIsMyTestSha"},
		{template: "version: [[", expectedError: fmt.Errorf("failed to parse cpe template 'version: [['")},
		{template: `version: [[index .CPE "artifactVersion"]], release: {{ .RELEASE }}`, expected: "version: 1.2.3, release: {{ .RELEASE }}"},
	}

	cpe := CPEMap{
		"artifactVersion": "1.2.3",
		"git/commitId":    "thisIsMyTestSha",
	}

	for _, test := range tt {
		res, err := cpe.ParseTemplateWithDelimiter(test.template, "[[", "]]")
		if test.expectedError != nil {
			assert.Contains(t, fmt.Sprint(err), fmt.Sprint(test.expectedError))
		} else {
			assert.NoError(t, err)
			assert.Equal(t, test.expected, (*res).String())
		}

	}
}

func TestTemplateFunctionCpe(t *testing.T) {
	t.Run("CPE from object", func(t *testing.T) {
		tt := []struct {
			element  string
			expected string
		}{
			{element: "artifactVersion", expected: "1.2.3"},
			{element: "git/commitId", expected: "thisIsMyTestSha"},
		}

		cpe := CPEMap{
			"artifactVersion": "1.2.3",
			"git/commitId":    "thisIsMyTestSha",
		}

		for _, test := range tt {
			assert.Equal(t, test.expected, cpe.cpe(test.element))
		}
	})

	t.Run("CPE from files", func(t *testing.T) {
		theVersion := "1.2.3"
		dir := t.TempDir()
		assert.NoError(t, os.WriteFile(filepath.Join(dir, "artifactVersion"), []byte(theVersion), 0o666))
		cpe := CPEMap{}
		assert.NoError(t, cpe.LoadFromDisk(dir))

		res, err := cpe.ParseTemplate(`{{cpe "artifactVersion"}}`)
		assert.NoError(t, err)
		assert.Equal(t, theVersion, (*res).String())
	})
}

func TestTemplateFunctionCustom(t *testing.T) {
	tt := []struct {
		element  string
		expected string
	}{
		{element: "repositoryUrl", expected: "https://this.is.the.repo.url"},
		{element: "repositoryId", expected: "repoTestId"},
	}

	cpe := CPEMap{
		"custom/repositoryUrl": "https://this.is.the.repo.url",
		"custom/repositoryId":  "repoTestId",
	}

	for _, test := range tt {
		assert.Equal(t, test.expected, cpe.custom(test.element))
	}
}

func TestTemplateFunctionGit(t *testing.T) {
	tt := []struct {
		element  string
		expected string
	}{
		{element: "commitId", expected: "thisIsMyTestSha"},
		{element: "repository", expected: "testRepo"},
	}

	cpe := CPEMap{
		"git/commitId":      "thisIsMyTestSha",
		"github/repository": "testRepo",
	}

	for _, test := range tt {
		assert.Equal(t, test.expected, cpe.git(test.element))
	}
}

func TestTemplateFunctionImageDigest(t *testing.T) {
	t.Run("CPE from object", func(t *testing.T) {
		tt := []struct {
			imageName string
			cpe       CPEMap
			expected  string
		}{
			{
				imageName: "image1",
				cpe:       CPEMap{},
				expected:  "",
			},
			{
				imageName: "image2",
				cpe: CPEMap{
					"container/imageDigests": []interface{}{"digest1", "digest2", "digest3"},
					"container/imageNames":   []interface{}{"image1", "image2", "image3"},
				},
				expected: "digest2",
			},
			{
				imageName: "image4",
				cpe: CPEMap{
					"container/imageDigests": []interface{}{"digest1", "digest2", "digest3"},
					"container/imageNames":   []interface{}{"image1", "image2", "image3"},
				},
				expected: "",
			},
			{
				imageName: "image1",
				cpe: CPEMap{
					"container/imageDigests": []interface{}{"digest1", "digest3"},
					"container/imageNames":   []interface{}{"image1", "image2", "image3"},
				},
				expected: "",
			},
		}

		for _, test := range tt {
			assert.Equal(t, test.expected, test.cpe.imageDigest(test.imageName))
		}
	})

	t.Run("CPE from files", func(t *testing.T) {
		dir := t.TempDir()

		imageDigests := []string{"digest1", "digest2", "digest3"}
		imageNames := []string{"image1", "image2", "image3"}
		cpeOut := CPEMap{"container/imageDigests": imageDigests, "container/imageNames": imageNames}
		assert.NoError(t, cpeOut.WriteToDisk(dir))

		cpe := CPEMap{}
		assert.NoError(t, cpe.LoadFromDisk(dir))

		res, err := cpe.ParseTemplate(`{{imageDigest "image2"}}`)
		assert.NoError(t, err)
		assert.Equal(t, "digest2", (*res).String())
	})
}

func TestTemplateFunctionImageTag(t *testing.T) {
	t.Run("CPE from object", func(t *testing.T) {
		tt := []struct {
			imageName string
			cpe       CPEMap
			expected  string
		}{
			{
				imageName: "image1",
				cpe:       CPEMap{},
				expected:  "",
			},
			{
				imageName: "image2",
				cpe: CPEMap{
					"container/imageNameTags": []interface{}{"image1:tag1", "image2:tag2", "image3:tag3"},
				},
				expected: "tag2",
			},
			{
				imageName: "image4",
				cpe: CPEMap{
					"container/imageNameTags": []interface{}{"image1:tag1", "image2:tag2", "image3:tag3"},
				},
				expected: "",
			},
		}

		for _, test := range tt {
			assert.Equal(t, test.expected, test.cpe.imageTag(test.imageName))
		}
	})

	t.Run("CPE from files", func(t *testing.T) {
		dir := t.TempDir()

		imageNameTags := []string{"image1:tag1", "image2:tag2", "image3:tag3"}
		imageNames := []string{"image1", "image2", "image3"}
		cpeOut := CPEMap{"container/imageNameTags": imageNameTags, "container/imageNames": imageNames}
		assert.NoError(t, cpeOut.WriteToDisk(dir))

		cpe := CPEMap{}
		assert.NoError(t, cpe.LoadFromDisk(dir))

		res, err := cpe.ParseTemplate(`{{imageTag "image2"}}`)
		assert.NoError(t, err)
		assert.Equal(t, "tag2", (*res).String())
	})
}
