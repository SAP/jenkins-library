//go:build unit
// +build unit

package log

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSetErrorCategory(t *testing.T) {
	SetErrorCategory(ErrorCustom)
	assert.Equal(t, errorCategory, ErrorCustom)
	assert.Equal(t, "custom", fmt.Sprint(errorCategory))
}

func TestGetErrorCategory(t *testing.T) {
	errorCategory = ErrorCompliance
	assert.Equal(t, GetErrorCategory(), errorCategory)
}

func TestSetFatalErrorDetail(t *testing.T) {
	sampleError := logrus.Fields{"Message": "Error happened"}
	errDetails, _ := json.Marshal(&sampleError)

	tests := []struct {
		name       string
		error      []byte
		want       []byte
		errPresent bool
	}{
		{
			name:       "set fatal error",
			error:      errDetails,
			want:       errDetails,
			errPresent: false,
		},
		{
			name:       "set fatal error - override",
			error:      errDetails,
			want:       errDetails,
			errPresent: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.errPresent {
				SetFatalErrorDetail(tt.error)
			}
			SetFatalErrorDetail(tt.error)
			assert.Equalf(t, tt.want, fatalError, "GetFatalErrorDetail()")
			fatalError = nil // reset error
		})
	}
}

func TestGetFatalErrorDetail(t *testing.T) {
	sampleError := logrus.Fields{"Message": "Error happened"}
	errDetails, _ := json.Marshal(&sampleError)

	tests := []struct {
		name       string
		errDetails []byte
		want       string
	}{
		{
			name:       "returns fatal error",
			errDetails: errDetails,
			want:       "{\"Message\":\"Error happened\"}",
		},
		{
			name:       "no fatal error set - returns empty",
			errDetails: nil,
			want:       "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.errDetails != nil {
				SetFatalErrorDetail(tt.errDetails)
			}
			str := ""
			data := []byte(str)
			fmt.Println(data)
			fatalErrorDetail := string(GetFatalErrorDetail())
			assert.Equalf(t, tt.want, fatalErrorDetail, "GetFatalErrorDetail()")
			fatalError = nil // resets fatal error
		})
	}
}
