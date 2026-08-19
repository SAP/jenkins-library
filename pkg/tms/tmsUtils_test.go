package tms

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_unmarshalServiceKey(t *testing.T) {
	tests := []struct {
		name           string
		serviceKeyJson string
		wantTmsUrl     string
		errMessage     string
	}{
		{
			name:           "standard cTMS service key uri works",
			serviceKeyJson: `{"uri": "https://my.tms.endpoint.sap.com"}`,
			wantTmsUrl:     "https://my.tms.endpoint.sap.com",
		},
		{
			name:           "standard cALM service key uri has expected postfix",
			serviceKeyJson: `{"endpoints": {"Api": "https://my.alm.endpoint.sap.com"}}`,
			wantTmsUrl:     "https://my.alm.endpoint.sap.com/imp-cdm-transport-management-api/v1",
		},
		{
			name:           "no uri or endpoints in service key leads to error",
			serviceKeyJson: `{"missing key options": "leads to error"}`,
			errMessage:     "neither uri nor endpoints.Api is set in service key json string",
		},
		{
			name:           "faulty json leads to error",
			serviceKeyJson: `"this is not correct json"`,
			errMessage:     "json: cannot unmarshal string into Go value of type tms.serviceKey",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotServiceKey, err := unmarshalServiceKey(tt.serviceKeyJson)
			if tt.errMessage == "" {
				assert.NoError(t, err, "No error was expected")
				assert.Equal(t, tt.wantTmsUrl, gotServiceKey.Uri, "Expected tms url does not match the uri in the service key")
			} else {
				assert.EqualError(t, err, tt.errMessage, "Error message not as expected")
			}
		})
	}
}
