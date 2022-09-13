package cmd

import (
	"fmt"
	"io"
	"net/http"
	"testing"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type pactMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func (p *pactMockUtils) SendRequest(method string, url string, r io.Reader, header http.Header, cookies []*http.Cookie) (*http.Response, error) {
	return nil, fmt.Errorf("not implemented")
}

func (p *pactMockUtils) SetOptions(options piperhttp.ClientOptions)

func (p *pactMockUtils) GetExitCode() int {
	return 0
}

func newPactPublishContractTestsUtils() pactMockUtils {
	utils := pactMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

func TestRunPactPublishContract(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := pactPublishContractOptions{}

		utils := newPactPublishContractTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runPactPublishContract(&config, nil, &utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := pactPublishContractOptions{}

		utils := newPactPublishContractTestsUtils()

		// test
		err := runPactPublishContract(&config, nil, &utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
