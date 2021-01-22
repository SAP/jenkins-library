package reporting

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToHTML(t *testing.T) {
	report := ScanReport{}
	expected := ``
	assert.Equal(t, expected, string(report.ToHTML()))
}
