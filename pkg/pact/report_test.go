package pact

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSaveReport(t *testing.T) {
	t.Run("success", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		report := Report{}
		reportData := ReportData{OrgOrigin: "theOrg"}
		err := report.SaveReport(&reportData, "theReport.json", "someText", "theName", "theValue", mockUtils)
		assert.NoError(t, err)
		c, _ :=mockUtils.ReadFile("theReport.json")
		resReport := Report{}
		_ = json.Unmarshal(c, &resReport)
		assert.Equal(t, "theOrg", resReport.Data.OrgOrigin)
		assert.Equal(t, "someText", resReport.Metrics[0].Metrics[0].Text)
		assert.Equal(t, "theName", resReport.Metrics[0].Metrics[0].Name)
		assert.Equal(t, "theValue", resReport.Metrics[0].Metrics[0].Value)
	})

	t.Run("failure - write file", func(t *testing.T){
		mockUtils := NewPactUtilsMock()
		mockUtils.FileWriteError = fmt.Errorf("write failed")
		report := Report{}
		reportData := ReportData{OrgOrigin: "theOrg"}
		err := report.SaveReport(&reportData, "theReport.json", "someText", "theName", "theValue", mockUtils)
		assert.EqualError(t, err, "write failed")
	})
}