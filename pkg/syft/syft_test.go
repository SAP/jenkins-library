//go:build unit
// +build unit

package syft_test

import (
	"net/http"
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/SAP/jenkins-library/pkg/syft"
	"github.com/jarcoal/httpmock"
	"github.com/pkg/errors"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSBOM(t *testing.T) {
	execMock := mock.ExecMockRunner{}
	fileMock := mock.FilesMock{}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	fakeArchive, err := fileMock.CreateArchive(map[string][]byte{"syft": []byte("test")})
	assert.NoError(t, err)

	httpmock.RegisterResponder(http.MethodGet, "http://test-syft-gh-release.com/syft.tar.gz", httpmock.NewBytesResponder(http.StatusOK, fakeArchive))
	httpmock.RegisterResponder(http.MethodGet, "http://not-found.com/syft.tar.gz", httpmock.NewBytesResponder(http.StatusNotFound, nil))
	httpmock.RegisterResponder(http.MethodGet, "http://failure.com/syft.tar.gz", httpmock.NewErrorResponder(errors.New("network error")))
	client := &piperhttp.Client{}
	client.SetOptions(piperhttp.ClientOptions{MaxRetries: -1, UseDefaultTransport: true})

	t.Run("should generate SBOM", func(t *testing.T) {
		err := syft.GenerateSBOM("http://test-syft-gh-release.com/syft.tar.gz", "", &execMock, &fileMock, client, "https://my-registry", []string{"image:latest", "image:1.2.3"})
		assert.NoError(t, err)

		assert.True(t, fileMock.HasFile("/tmp/syfttest/syft"))
		fi, err := fileMock.Stat("/tmp/syfttest/syft")
		assert.NoError(t, err)
		assert.Equal(t, fi.Mode().Perm().String(), "-rwxr-xr-x")

		assert.Len(t, execMock.Calls, 2)
		firstCall := execMock.Calls[0]
		assert.Equal(t, firstCall.Exec, "/tmp/syfttest/syft")
		assert.Equal(t, firstCall.Params, []string{"scan", "registry:my-registry/image:latest", "-o", "cyclonedx-xml@1.4=bom-docker-0.xml", "-q"})

		secondCall := execMock.Calls[1]
		assert.Equal(t, secondCall.Exec, "/tmp/syfttest/syft")
		assert.Equal(t, secondCall.Params, []string{"scan", "registry:my-registry/image:1.2.3", "-o", "cyclonedx-xml@1.4=bom-docker-1.xml", "-q"})
	})

	t.Run("error case: syft execution failed", func(t *testing.T) {
		execMock = mock.ExecMockRunner{}
		execMock.ShouldFailOnCommand = map[string]error{
			"/tmp/syfttest/syft scan registry:my-registry/image:latest -o cyclonedx-xml@1.4=bom-docker-0.xml -q": errors.New("failed"),
		}

		err := syft.GenerateSBOM("http://test-syft-gh-release.com/syft.tar.gz", "", &execMock, &fileMock, client, "https://my-registry", []string{"image:latest"})
		assert.Error(t, err)
		assert.Equal(t, "failed to generate SBOM: failed", err.Error())
	})

	t.Run("error case: no registry", func(t *testing.T) {
		err := syft.GenerateSBOM("http://test-syft-gh-release.com/syft.tar.gz", "", &execMock, &fileMock, client, "", []string{"image:latest"})
		assert.Error(t, err)
		assert.Equal(t, "syft: registry url must not be empty", err.Error())
	})

	t.Run("error case: no images provided", func(t *testing.T) {
		err := syft.GenerateSBOM("http://test-syft-gh-release.com/syft.tar.gz", "", &execMock, &fileMock, client, "my-registry", nil)
		assert.Error(t, err)
		assert.Equal(t, "syft: no images provided", err.Error())
	})

	t.Run("error case: empty image name", func(t *testing.T) {
		err := syft.GenerateSBOM("http://test-syft-gh-release.com/syft.tar.gz", "", &execMock, &fileMock, client, "my-registry", []string{""})
		assert.Error(t, err)
		assert.Equal(t, "syft: image name must not be empty", err.Error())
	})

	t.Run("error case: failed to download archive (not found)", func(t *testing.T) {
		err := syft.GenerateSBOM("http://not-found.com/syft.tar.gz", "", &execMock, &fileMock, client, "my-registry", []string{"img"})
		assert.Error(t, err)
		assert.Equal(t, "failed to install syft: failed to download syft binary: request to http://not-found.com/syft.tar.gz returned with response 404 Not Found", err.Error())
	})

	t.Run("error case: failed to download archive (network error)", func(t *testing.T) {
		err := syft.GenerateSBOM("http://failure.com/syft.tar.gz", "", &execMock, &fileMock, client, "my-registry", []string{"img"})
		assert.Error(t, err)
		assert.Equal(t, "failed to install syft: failed to download syft binary: HTTP GET request to http://failure.com/syft.tar.gz failed: Get \"http://failure.com/syft.tar.gz\": network error", err.Error())
	})
}
