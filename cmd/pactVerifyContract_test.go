package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunPactVerifyContract(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := pactVerifyContractOptions{}

		utils := newPactPublishContractTestsUtils()
		utils.AddFile("file.txt", []byte("dummy content"))

		// test
		err := runPactVerifyContract(&config, nil, &utils)

		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// init
		config := pactVerifyContractOptions{}

		utils := newPactPublishContractTestsUtils()

		// test
		err := runPactVerifyContract(&config, nil, &utils)

		// assert
		assert.EqualError(t, err, "cannot run without important file")
	})
}
