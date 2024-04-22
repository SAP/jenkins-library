package codeql

import (
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

type CodeqlSarifUploaderMock struct {
	counter int
}

func (c *CodeqlSarifUploaderMock) GetSarifStatus() (SarifFileInfo, error) {
	if c.counter == 0 {
		return SarifFileInfo{
			ProcessingStatus: "complete",
			Errors:           nil,
		}, nil
	}
	if c.counter == -1 {
		return SarifFileInfo{
			ProcessingStatus: "failed",
			Errors:           []string{"upload error"},
		}, nil
	}
	c.counter--
	return SarifFileInfo{
		ProcessingStatus: "pending",
		Errors:           nil,
	}, nil
}

type CodeqlSarifUploaderErrorMock struct {
	counter int
}

func (c *CodeqlSarifUploaderErrorMock) GetSarifStatus() (SarifFileInfo, error) {
	if c.counter == -1 {
		return SarifFileInfo{}, errors.New("test error")
	}
	if c.counter == 0 {
		return SarifFileInfo{
			ProcessingStatus: "complete",
			Errors:           nil,
		}, nil
	}
	c.counter--
	return SarifFileInfo{ProcessingStatus: "Service unavailable"}, nil
}

func TestWaitSarifUploaded(t *testing.T) {
	t.Parallel()
	sarifCheckRetryInterval := 1
	sarifCheckMaxRetries := 5
	t.Run("Fast complete upload", func(t *testing.T) {
		codeqlScanAuditMock := CodeqlSarifUploaderMock{counter: 0}
		timerStart := time.Now()
		err := WaitSarifUploaded(sarifCheckMaxRetries, sarifCheckRetryInterval, &codeqlScanAuditMock)
		assert.Less(t, time.Now().Sub(timerStart), time.Second)
		assert.NoError(t, err)
	})
	t.Run("Long completed upload", func(t *testing.T) {
		codeqlScanAuditMock := CodeqlSarifUploaderMock{counter: 2}
		timerStart := time.Now()
		err := WaitSarifUploaded(sarifCheckMaxRetries, sarifCheckRetryInterval, &codeqlScanAuditMock)
		assert.GreaterOrEqual(t, time.Now().Sub(timerStart), time.Second*2)
		assert.NoError(t, err)
	})
	t.Run("Failed upload", func(t *testing.T) {
		codeqlScanAuditMock := CodeqlSarifUploaderMock{counter: -1}
		err := WaitSarifUploaded(sarifCheckMaxRetries, sarifCheckRetryInterval, &codeqlScanAuditMock)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "failed to upload sarif file")
	})
	t.Run("Error while checking sarif uploading", func(t *testing.T) {
		codeqlScanAuditErrorMock := CodeqlSarifUploaderErrorMock{counter: -1}
		err := WaitSarifUploaded(sarifCheckMaxRetries, sarifCheckRetryInterval, &codeqlScanAuditErrorMock)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "test error")
	})
	t.Run("Completed upload after getting errors from server", func(t *testing.T) {
		codeqlScanAuditErrorMock := CodeqlSarifUploaderErrorMock{counter: 3}
		err := WaitSarifUploaded(sarifCheckMaxRetries, sarifCheckRetryInterval, &codeqlScanAuditErrorMock)
		assert.NoError(t, err)
	})
	t.Run("Max retries reached", func(t *testing.T) {
		codeqlScanAuditErrorMock := CodeqlSarifUploaderErrorMock{counter: 6}
		err := WaitSarifUploaded(sarifCheckMaxRetries, sarifCheckRetryInterval, &codeqlScanAuditErrorMock)
		assert.Error(t, err)
		assert.ErrorContains(t, err, "max retries reached")
	})
}
