package asc

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeployAppReleaseDateValidation(t *testing.T) {
	// The release date is validated before any file or HTTP interaction. To
	// exercise the validation branch without a real backend, valid dates are
	// expected to proceed past validation and fail at file open instead, while
	// invalid dates must fail with the validation error up front.
	sys := &SystemInstance{serverURL: "https://asc.example.com"}

	tests := []struct {
		name        string
		releaseDate string
		wantErr     string
	}{
		{name: "valid YYYY-MM-DD passes validation", releaseDate: "2026-09-18", wantErr: "unable to locate file"},
		{name: "legacy MM/DD/YYYY passes validation", releaseDate: "09/18/2026", wantErr: "unable to locate file"},
		{name: "empty defaults to today and passes validation", releaseDate: "", wantErr: "unable to locate file"},
		{name: "invalid month is rejected", releaseDate: "2026-13-01", wantErr: "invalid release date"},
		{name: "invalid day is rejected", releaseDate: "2026-09-45", wantErr: "invalid release date"},
		{name: "unpadded date is rejected", releaseDate: "2026-9-18", wantErr: "invalid release date"},
		{name: "garbage is rejected", releaseDate: "not-a-date", wantErr: "invalid release date"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := sys.DeployApp(DeployRequest{
				FilePath:    "/nonexistent/does-not-exist.ipa",
				ReleaseDate: test.releaseDate,
			})
			assert.Error(t, err)
			assert.True(t, strings.Contains(err.Error(), test.wantErr),
				"expected error containing %q, got %q", test.wantErr, err.Error())
		})
	}
}
