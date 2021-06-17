package cmd

import (
	"errors"
	"testing"

	"github.com/SAP/jenkins-library/pkg/influx/mocks"
	"github.com/stretchr/testify/mock"
)

func TestWriteData(t *testing.T) {
	options := &influxWriteDataOptions{
		"http://localhost:8086",
		"authToken",
		"piper",
		"org",
		map[string]interface{}{
			"series_1": map[string]interface{}{"field_a": 11, "field_b": 12},
			"series_2": map[string]interface{}{"field_c": 21, "field_d": 22},
		},
		map[string]interface{}{
			"series_1": map[string]interface{}{"tag_a": "a", "tag_b": "b"},
			"series_2": map[string]interface{}{"tag_c": "c", "tag_d": "d"},
		},
	}
	errWriteData := errors.New("error")
	tests := []struct {
		name         string
		writeDataErr error
		expectedErr  error
	}{
		{
			"Test writing data - success",
			nil,
			nil,
		},
		{
			"Test writing data - failed",
			errWriteData,
			errWriteData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			influxClientMock := &mocks.Client{}
			writeAPIBlockingMock := &mocks.WriteAPIBlocking{}
			writeAPIBlockingMock.On("WritePoint", mock.Anything, mock.Anything).Return(tt.writeDataErr)
			influxClientMock.On("WriteAPIBlocking", mock.Anything, mock.Anything).Return(writeAPIBlockingMock)
			err := writeData(options, influxClientMock)
			if err != tt.expectedErr {
				t.Errorf("\nactual: %q\nexpected: %q\n", err, tt.expectedErr)
			}
		})
	}
}
