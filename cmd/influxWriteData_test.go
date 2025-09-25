package cmd

import (
	"errors"
	"testing"

	"github.com/SAP/jenkins-library/pkg/influx/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWriteData(t *testing.T) {
	options := &influxWriteDataOptions{
		ServerURL:    "http://localhost:8086",
		AuthToken:    "authToken",
		Bucket:       "piper",
		Organization: "org",
	}
	errString := "some error"
	errWriteData := errors.New(errString)
	tests := []struct {
		name          string
		dataMap       string
		dataMapTags   string
		writeDataErr  error
		errExpected   bool
		errIncludeStr string
	}{
		{
			"Test writing metrics with correct json data - success",
			`{"series_1": {"field_a": 11, "field_b": 12}, "series_2": {"field_c": 21, "field_d": 22}}`,
			`{"series_1": {"tag_a": "a", "tag_b": "b"}, "series_2": {"tag_c": "c", "tag_d": "d"}}`,
			nil,
			false,
			"",
		},
		{
			"Test writing metrics with invalid dataMap",
			`"series_1": {"field_a": 11, "field_b": 12}, "series_2": {"field_c": 21, "field_d": 22}`,
			`{"series_1": {"tag_a": "a", "tag_b": "b"}, "series_2": {"tag_c": "c", "tag_d": "d"}}`,
			nil,
			false,
			"Failed to unmarshal dataMap:",
		},
		{
			"Test writing metrics with invalid dataMapTags",
			`{"series_1": {"field_a": 11, "field_b": 12}, "series_2": {"field_c": 21, "field_d": 22}}`,
			`{"series_1": {"tag_a": 2, "tag_b": "b"}, "series_2": {"tag_c": "c", "tag_d": "d"}}`,
			nil,
			false,
			"Failed to unmarshal dataMapTags:",
		},
		{
			"Test writing metrics with correct json data - failed",
			`{"series_1": {"field_a": 11, "field_b": 12}, "series_2": {"field_c": 21, "field_d": 22}}`,
			`{"series_1": {"tag_a": "a", "tag_b": "b"}, "series_2": {"tag_c": "c", "tag_d": "d"}}`,
			errWriteData,
			true,
			errString,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			influxClientMock := &mocks.Client{}
			writeAPIBlockingMock := &mocks.WriteAPIBlocking{}
			writeAPIBlockingMock.On("WritePoint", mock.Anything, mock.Anything).Return(tt.writeDataErr)
			influxClientMock.On("WriteAPIBlocking", mock.Anything, mock.Anything).Return(writeAPIBlockingMock)
			options.DataMap = tt.dataMap
			options.DataMapTags = tt.dataMapTags
			err := writeData(options, influxClientMock)
			if err != nil {
				assert.Contains(t, err.Error(), tt.errIncludeStr)
			}
		})
	}
}
