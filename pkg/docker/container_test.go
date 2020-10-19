package docker

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContainerRegistryFromURL(t *testing.T) {
	tt := []struct {
		url           string
		expected      string
		expectedError string
	}{
		{url: "", expected: "", expectedError: "invalid registry url"},
		{url: "invalidUrl", expected: "", expectedError: "invalid registry url"},
		{url: "no.protocol.com", expected: "", expectedError: "invalid registry url"},
		{url: "no.protocol.com:50000", expected: "", expectedError: "invalid registry url"},
		{url: "no.protocol.com:50000/my/path/to/image", expected: "", expectedError: "invalid registry url"},
		{url: "no.protocol.com:50000/my/path/to/image:withTag", expected: "", expectedError: "invalid registry url"},
		{url: "no.protocol.com:50000/my/path/to/image@withDigest", expected: "", expectedError: "invalid registry url"},
		{url: "https://my.registry.com", expected: "my.registry.com"},
		{url: "https://my.registry.com:50000", expected: "my.registry.com:50000"},
		{url: "https://my.registry.com:50000/", expected: "my.registry.com:50000"},
		{url: "https://my.registry.com:50000/my/path/to/image:withTag", expected: "my.registry.com:50000"},
		{url: "https://my.registry.com:50000/my/path/to/image@withDigest", expected: "my.registry.com:50000"},
	}

	for _, test := range tt {
		t.Run(test.url, func(t *testing.T) {
			got, err := ContainerRegistryFromURL(test.url)
			if len(test.expectedError) > 0 {
				assert.Contains(t, fmt.Sprint(err), test.expectedError)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestContainerRegistryFromImage(t *testing.T) {
	tt := []struct {
		image         string
		expected      string
		expectedError string
	}{
		{image: "", expected: "", expectedError: "failed to parse image name"},
		{image: "onlyImage", expected: "index.docker.io"},
		{image: "onlyimage", expected: "index.docker.io"},
		{image: "onlyimage:withTag", expected: "index.docker.io"},
		{image: "onlyimage@sha256:152f65865ae43b143b1e42dacdb5e9c473dd70b3adc5b79af7cf585cc8605205", expected: "index.docker.io"},
		{image: "path/to/image", expected: "index.docker.io"},
		{image: "my.registry.com/onlyimage", expected: "my.registry.com"},
		{image: "my.registry.com:50000/onlyimage", expected: "my.registry.com:50000"},
		{image: "my.registry.com:50000/onlyimage:withTag", expected: "my.registry.com:50000"},
		{image: "my.registry.com:50000/onlyimage@sha256:152f65865ae43b143b1e42dacdb5e9c473dd70b3adc5b79af7cf585cc8605205", expected: "my.registry.com:50000"},
		{image: "my.registry.com:50000/path/to/image:withTag", expected: "my.registry.com:50000"},
		{image: "my.registry.com:50000/path/to/image@sha256:152f65865ae43b143b1e42dacdb5e9c473dd70b3adc5b79af7cf585cc8605205", expected: "my.registry.com:50000"},
	}

	for _, test := range tt {
		t.Run(test.image, func(t *testing.T) {
			got, err := ContainerRegistryFromImage(test.image)
			if len(test.expectedError) > 0 {
				assert.Contains(t, fmt.Sprint(err), test.expectedError)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expected, got)
		})
	}
}

func TestContainerImageNameTagFromImage(t *testing.T) {
	tt := []struct {
		image         string
		expected      string
		expectedError string
	}{
		{image: "", expected: "", expectedError: "failed to parse image name"},
		{image: "onlyImage", expected: "onlyImage"},
		{image: "onlyimage", expected: "onlyimage"},
		{image: "onlyimage:withTag", expected: "onlyimage:withTag"},
		{image: "onlyimage@sha256:152f65865ae43b143b1e42dacdb5e9c473dd70b3adc5b79af7cf585cc8605205", expected: "onlyimage@sha256:152f65865ae43b143b1e42dacdb5e9c473dd70b3adc5b79af7cf585cc8605205"},
		{image: "path/to/image", expected: "path/to/image"},
		{image: "my.registry.com/onlyimage", expected: "onlyimage"},
		{image: "my.registry.com:50000/onlyimage", expected: "onlyimage"},
		{image: "my.registry.com:50000/onlyimage:withTag", expected: "onlyimage:withTag"},
		{image: "my.registry.com:50000/onlyimage@sha256:152f65865ae43b143b1e42dacdb5e9c473dd70b3adc5b79af7cf585cc8605205", expected: "onlyimage@sha256:152f65865ae43b143b1e42dacdb5e9c473dd70b3adc5b79af7cf585cc8605205"},
		{image: "my.registry.com:50000/path/to/image:withTag", expected: "path/to/image:withTag"},
		{image: "my.registry.com:50000/path/to/image@sha256:152f65865ae43b143b1e42dacdb5e9c473dd70b3adc5b79af7cf585cc8605205", expected: "path/to/image@sha256:152f65865ae43b143b1e42dacdb5e9c473dd70b3adc5b79af7cf585cc8605205"},
	}

	for _, test := range tt {
		t.Run(test.image, func(t *testing.T) {
			got, err := ContainerImageNameTagFromImage(test.image)
			if len(test.expectedError) > 0 {
				assert.Contains(t, fmt.Sprint(err), test.expectedError)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expected, got)
		})
	}
}
