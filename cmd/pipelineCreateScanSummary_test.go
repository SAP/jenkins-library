package cmd

import (
	"fmt"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
)

type pipelineCreateScanSummaryMockUtils struct {
	*mock.FilesMock
}

func newPipelineCreateScanSummaryTestsUtils() pipelineCreateScanSummaryMockUtils {
	utils := pipelineCreateScanSummaryMockUtils{
		FilesMock: &mock.FilesMock{},
	}
	return utils
}

func TestRunPipelineCreateScanSummary(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		config := pipelineCreateScanSummaryOptions{
			OutputFilePath: "scanSummary.md",
		}

		utils := newPipelineCreateScanSummaryTestsUtils()
		utils.AddFile(".pipeline/stepReports/step1.json", []byte(`{"title":"Title Scan 1"}`))
		utils.AddFile(".pipeline/stepReports/step2.json", []byte(`{"title":"Title Scan 2"}`))
		utils.AddFile(".pipeline/stepReports/step3.json", []byte(`{"title":"Title Scan 3"}`))

		err := runPipelineCreateScanSummary(&config, nil, utils)

		assert.NoError(t, err)
		reportExists, _ := utils.FileExists("scanSummary.md")
		assert.True(t, reportExists)
		fileContent, _ := utils.FileRead("scanSummary.md")
		fileContentString := string(fileContent)
		assert.Contains(t, fileContentString, "Title Scan 1")
		assert.Contains(t, fileContentString, "Title Scan 2")
		assert.Contains(t, fileContentString, "Title Scan 3")
	})

	t.Run("success - failed only", func(t *testing.T) {
		t.Parallel()

		config := pipelineCreateScanSummaryOptions{
			OutputFilePath: "scanSummary.md",
			FailedOnly:     true,
		}

		utils := newPipelineCreateScanSummaryTestsUtils()
		utils.AddFile(".pipeline/stepReports/step1.json", []byte(`{"title":"Title Scan 1", "successfulScan": true}`))
		utils.AddFile(".pipeline/stepReports/step2.json", []byte(`{"title":"Title Scan 2", "successfulScan": false}`))
		utils.AddFile(".pipeline/stepReports/step3.json", []byte(`{"title":"Title Scan 3", "successfulScan": false}`))

		err := runPipelineCreateScanSummary(&config, nil, utils)

		assert.NoError(t, err)
		reportExists, _ := utils.FileExists("scanSummary.md")
		assert.True(t, reportExists)
		fileContent, _ := utils.FileRead("scanSummary.md")
		fileContentString := string(fileContent)
		assert.NotContains(t, fileContentString, "Title Scan 1")
		assert.Contains(t, fileContentString, "Title Scan 2")
		assert.Contains(t, fileContentString, "Title Scan 3")
	})

	t.Run("success - with source link", func(t *testing.T) {
		t.Parallel()

		config := pipelineCreateScanSummaryOptions{
			OutputFilePath: "scanSummary.md",
			PipelineLink:   "https://test.com/link",
		}

		utils := newPipelineCreateScanSummaryTestsUtils()

		err := runPipelineCreateScanSummary(&config, nil, utils)

		assert.NoError(t, err)
		reportExists, _ := utils.FileExists("scanSummary.md")
		assert.True(t, reportExists)
		fileContent, _ := utils.FileRead("scanSummary.md")
		fileContentString := string(fileContent)
		assert.Contains(t, fileContentString, "https://test.com/link")
	})

	t.Run("error - read file", func(t *testing.T) {
		t.Skip()
		//ToDo
		// so far mock cannot create error for reading files

		config := pipelineCreateScanSummaryOptions{
			OutputFilePath: "scanSummary.md",
		}

		utils := newPipelineCreateScanSummaryTestsUtils()

		err := runPipelineCreateScanSummary(&config, nil, utils)

		assert.Contains(t, fmt.Sprint(err), "failed to read report")
	})

	t.Run("error - unmarshal json", func(t *testing.T) {
		t.Parallel()

		config := pipelineCreateScanSummaryOptions{
			OutputFilePath: "scanSummary.md",
		}

		utils := newPipelineCreateScanSummaryTestsUtils()
		utils.AddFile(".pipeline/stepReports/step1.json", []byte(`{"title":"Title Scan 1"`))

		err := runPipelineCreateScanSummary(&config, nil, utils)

		assert.Contains(t, fmt.Sprint(err), "failed to parse report")
	})

	t.Run("error - write file", func(t *testing.T) {
		//ToDo
		// so far mock cannot create error
	})

}
