//go:build unit
// +build unit

package cmd

import (
	"errors"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type credentialdiggerScanMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
	noerr bool
}

func newCDTestsUtils() credentialdiggerScanMockUtils {
	utils := credentialdiggerScanMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
		noerr:          true, // flag for return value of MockRunner
	}
	return utils
}
func (c credentialdiggerScanMockUtils) RunExecutable(executable string, params ...string) error {
	if c.noerr {
		return nil
	} else {
		return errors.New("Some custom error")
	}
}

func TestCredentialdiggerFullScan(t *testing.T) {
	t.Run("Valid full scan without discoveries", func(t *testing.T) {
		config := credentialdiggerScanOptions{Repository: "testRepo", Token: "validToken"}
		utils := newCDTestsUtils()
		assert.Equal(t, nil, credentialdiggerFullScan(&config, nil, utils))

	})
	t.Run("Full scan with discoveries or wrong arguments", func(t *testing.T) {
		config := credentialdiggerScanOptions{Repository: "testRepo", Token: "validToken"}
		utils := newCDTestsUtils()
		utils.noerr = false
		assert.EqualError(t, credentialdiggerFullScan(&config, nil, utils), "Some custom error")
	})
}

func TestCredentialdiggerScanSnapshot(t *testing.T) {
	t.Run("Valid scan snapshot without discoveries", func(t *testing.T) {
		config := credentialdiggerScanOptions{Repository: "testRepo", Token: "validToken", Snapshot: "main"}
		utils := newCDTestsUtils()
		assert.Equal(t, nil, credentialdiggerScanSnapshot(&config, nil, utils))
	})
	t.Run("Scan snapshot with discoveries or wrong arguments", func(t *testing.T) {
		config := credentialdiggerScanOptions{Repository: "testRepo", Token: "validToken", Snapshot: "main"}
		utils := newCDTestsUtils()
		utils.noerr = false
		assert.EqualError(t, credentialdiggerScanSnapshot(&config, nil, utils), "Some custom error")
	})
}

func TestCredentialdiggerScanPR(t *testing.T) {
	t.Run("Valid scan pull request without discoveries", func(t *testing.T) {
		config := credentialdiggerScanOptions{Repository: "testRepo", Token: "validToken", PrNumber: 1}
		utils := newCDTestsUtils()
		assert.Equal(t, nil, credentialdiggerScanPR(&config, nil, utils))
	})
	t.Run("Scan pull request with discoveries or wrong arguments", func(t *testing.T) {
		config := credentialdiggerScanOptions{Repository: "testRepo", Token: "validToken", PrNumber: 1}
		utils := newCDTestsUtils()
		utils.noerr = false
		assert.EqualError(t, credentialdiggerScanPR(&config, nil, utils), "Some custom error")
	})
}

func TestCredentialdiggerAddRules(t *testing.T) {
	t.Run("Valid standard or remote rules", func(t *testing.T) {
		config := credentialdiggerScanOptions{}
		utils := newCDTestsUtils()
		assert.Equal(t, nil, credentialdiggerAddRules(&config, nil, utils))
	})
	t.Run("Broken add rules", func(t *testing.T) {
		config := credentialdiggerScanOptions{}
		utils := newCDTestsUtils()
		utils.noerr = false
		assert.EqualError(t, credentialdiggerAddRules(&config, nil, utils), "Some custom error")
	})
	/*
		// In case we want to test the error raised by piperhttp
		t.Run("Invalid external rules link", func(t *testing.T) {
			rulesExt := "https://broken-link.com/fakerules"
			config := credentialdiggerScanOptions{RulesDownloadURL: rulesExt}
			utils := newCDTestsUtils()
			assert.Equal(t, nil, credentialdiggerAddRules(&config, nil, utils))
		})
	*/
}

func TestCredentialdiggerGetDiscoveries(t *testing.T) {
	t.Run("Empty discoveries", func(t *testing.T) {
		config := credentialdiggerScanOptions{Repository: "testRepo"}
		utils := newCDTestsUtils()
		assert.Equal(t, nil, credentialdiggerGetDiscoveries(&config, nil, utils))
	})
	t.Run("Get discoveries non-empty", func(t *testing.T) {
		config := credentialdiggerScanOptions{Repository: "testRepo"}
		utils := newCDTestsUtils()
		utils.noerr = false
		assert.EqualError(t, credentialdiggerGetDiscoveries(&config, nil, utils), "Some custom error")
	})
}

func TestCredentialdiggerBuildCommonArgs(t *testing.T) {
	t.Run("Valid build common args", func(t *testing.T) {
		arguments := []string{"repoURL", "--sqlite", "piper_step_db.db", "--git_token", "validToken",
			"--debug", "--models", "model1", "model2"}
		config := credentialdiggerScanOptions{Repository: "repoURL", Token: "validToken", Snapshot: "main",
			Debug: true, PrNumber: 1,
			Models: []string{"model1", "model2"},
		}
		assert.Equal(t, arguments, credentialdiggerBuildCommonArgs(&config))
	})

}
