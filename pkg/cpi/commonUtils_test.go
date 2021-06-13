package cpi

import (
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestReadCpiServiceKeyFile(t *testing.T) {
	properServiceKey := `{
			"url": "https://demo",
			"uaa": {
				"clientid": "demouser",
				"clientsecret": "******",
				"url": "https://demo/oauth/token"
				}
			}`
	faultyServiceKey := `this is not json`

	type args struct {
		serviceKeyPath string
		fileUtils      piperutils.FileUtils
	}
	tests := []struct {
		name              string
		args              args
		wantCpiServiceKey CpiServiceKey
		wantedErrorMsg    string
	}{
		{
			"happy path",
			args{
				serviceKeyPath: "positive/outcome/serviceKey.json",
				fileUtils: &FileMock{
					FileReadContent: map[string]string{"positive/outcome/serviceKey.json": properServiceKey},
				},
			},
			CpiServiceKey{
				Host: "https://demo",
				Uaa: OAuth{
					OAuthTokenProviderURL: "https://demo/oauth/token",
					ClientId:              "demouser",
					ClientSecret:          "******",
				},
			},
			"",
		},
		{
			"faulty json",
			args{
				serviceKeyPath: "",
				fileUtils: &FileMock{
					FileReadContent: map[string]string{"faulty/serviceKey.json": faultyServiceKey},
				},
			},
			CpiServiceKey{},
			"error unmarshalling serviceKey: unexpected end of JSON input",
		},
		{
			"read file error",
			args{
				serviceKeyPath: "non/existent/serviceKey.json",
				fileUtils: &FileMock{
					FileReadErr: map[string]error{"non/existent/serviceKey.json": errors.New("this file does not exist")},
				},
			},
			CpiServiceKey{},
			"error reading serviceKey file: this file does not exist",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCpiServiceKey, err := ReadCpiServiceKeyFile(tt.args.serviceKeyPath, tt.args.fileUtils)
			if tt.wantedErrorMsg != "" {
				assert.EqualError(t, err, tt.wantedErrorMsg)
			}
			assert.Equal(t, tt.wantCpiServiceKey, gotCpiServiceKey)
		})
	}
}
