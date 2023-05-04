//go:build unit
// +build unit

package registry

import (
	"net/http"
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestSearchBuildpack(t *testing.T) {
	t.Run("returns image URL for specific version", func(t *testing.T) {

		fakeResponse := "{\"latest\":{\"version\":\"1.1.1\",\"namespace\":\"test\",\"name\":\"test\",\"description\":\"\",\"homepage\":\"\",\"licenses\":null,\"stacks\":[\"test\",\"test\"],\"id\":\"test\"},\"versions\":[{\"version\":\"1.1.1\",\"_link\":\"https://test/1.1.1\"}]}"

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(http.MethodGet, "https://registry.buildpacks.io/api/v1/buildpacks/test", httpmock.NewStringResponder(200, fakeResponse))
		httpmock.RegisterResponder(http.MethodGet, "https://test/1.1.1", httpmock.NewStringResponder(200, "{\"addr\": \"index.docker.io/test@1.1.1\"}"))
		client := &piperhttp.Client{}
		client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})

		img, err := SearchBuildpack("test", "1.1.1", client, "")

		assert.NoError(t, err)
		assert.Equal(t, "index.docker.io/test@1.1.1", img)
	})

	t.Run("returns image URL for the latest", func(t *testing.T) {
		fakeResponse := "{\"latest\":{\"version\":\"1.1.1\",\"namespace\":\"test\",\"name\":\"test\",\"description\":\"\",\"homepage\":\"\",\"licenses\":null,\"stacks\":[\"test\",\"test\"],\"id\":\"test\"},\"versions\":[{\"version\":\"1.1.1\",\"_link\":\"https://test/1.1.1\"}]}"

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(http.MethodGet, "https://registry.buildpacks.io/api/v1/buildpacks/test", httpmock.NewStringResponder(200, fakeResponse))
		httpmock.RegisterResponder(http.MethodGet, "https://test/1.1.1", httpmock.NewStringResponder(200, "{\"addr\": \"index.docker.io/test@1.1.1\"}"))
		client := &piperhttp.Client{}
		client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})

		img, err := SearchBuildpack("test", "", client, "")

		assert.NoError(t, err)
		assert.Equal(t, "index.docker.io/test@1.1.1", img)
	})

	t.Run("fails with the version not found", func(t *testing.T) {
		fakeResponse := "{\"latest\":{\"version\":\"1.1.1\",\"namespace\":\"test\",\"name\":\"test\",\"description\":\"\",\"homepage\":\"\",\"licenses\":null,\"stacks\":[\"test\",\"test\"],\"id\":\"test\"},\"versions\":[{\"version\":\"1.1.1\",\"_link\":\"https://test/1.1.1\"}]}"

		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(http.MethodGet, "https://registry.buildpacks.io/api/v1/buildpacks/test", httpmock.NewStringResponder(200, fakeResponse))
		httpmock.RegisterResponder(http.MethodGet, "https://test/1.1.1", httpmock.NewStringResponder(200, "{\"addr\": \"index.docker.io/test@1.1.1\"}"))
		client := &piperhttp.Client{}
		client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})

		img, err := SearchBuildpack("test", "1.1.2", client, "")

		assert.Error(t, err)
		assert.Equal(t, "version '1.1.2' was not found for the buildpack 'test'", err.Error())
		assert.Equal(t, "", img)
	})

	t.Run("fails with the HTTP error", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(http.MethodGet, "https://registry.buildpacks.io/api/v1/buildpacks/test", httpmock.NewStringResponder(404, "not_found"))
		client := &piperhttp.Client{}
		client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})

		img, err := SearchBuildpack("test", "1.1.2", client, "")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "returned with response 404")
		assert.Equal(t, "", img)
	})

	t.Run("fails with the invalid response object", func(t *testing.T) {
		httpmock.Activate()
		defer httpmock.DeactivateAndReset()
		httpmock.RegisterResponder(http.MethodGet, "https://registry.buildpacks.io/api/v1/buildpacks/test", httpmock.NewStringResponder(200, "not_a_json"))
		client := &piperhttp.Client{}
		client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})

		img, err := SearchBuildpack("test", "1.1.2", client, "")

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unable to parse response from the https://registry.buildpacks.io/api/v1/buildpacks/test, error: invalid character")
		assert.Equal(t, "", img)
	})
}
