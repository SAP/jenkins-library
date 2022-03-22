package body

import (
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

func TestReadResponseBody(t *testing.T) {
	tests := []struct {
		name        string
		response    *http.Response
		want        []byte
		wantErrText string
	}{
		{
			name:     "Straight forward",
			response: httpmock.NewStringResponse(200, "test string"),
			want:     []byte("test string"),
		},
		{
			name:        "No response error",
			wantErrText: "did not retrieve an HTTP response",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ReadResponseBody(tt.response)
			if tt.wantErrText != "" {
				require.Error(t, err, "Error expected")
				assert.EqualError(t, err, tt.wantErrText, "Error is not equal")
				return
			}
			require.NoError(t, err, "No error expected")
			assert.Equal(t, tt.want, got, "Did not receive expected body")
		})
	}
}
