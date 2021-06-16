package influx

import (
	"context"
	"errors"
	"testing"

	"github.com/SAP/jenkins-library/pkg/influx/mocks"
	"github.com/stretchr/testify/mock"
)

func TestWriteMetrics(t *testing.T) {
	errWriteMetrics := errors.New("error")
	tests := []struct {
		name          string
		dataMap       map[string]interface{}
		dataMapTags   map[string]interface{}
		writePointErr error
		err           error
	}{
		{
			"Test writing metrics with correct data - success",
			map[string]interface{}{
				"series_1": map[string]interface{}{"field_a": 11, "field_b": 12},
				"series_2": map[string]interface{}{"field_c": 21, "field_d": 22},
			},
			map[string]interface{}{
				"series_1": map[string]interface{}{"tag_a": "a", "tag_b": "b"},
				"series_2": map[string]interface{}{"tag_c": "c", "tag_d": "d"},
			},
			nil,
			nil,
		},
		{
			"Test writing metrics with wrong dataMap - failed",
			map[string]interface{}{"series_1": "something"},
			map[string]interface{}{
				"series_1": map[string]interface{}{"tag_a": "a", "tag_b": "b"},
				"series_2": map[string]interface{}{"tag_c": "c", "tag_d": "d"},
			},
			nil,
			errFieldsMapAssertion,
		},
		{
			"Test writing metrics with wrong dataMapTags - failed",
			map[string]interface{}{
				"series_1": map[string]interface{}{"field_a": 11, "field_b": 12},
				"series_2": map[string]interface{}{"field_c": 21, "field_d": 22},
			},
			map[string]interface{}{"series_1": "something"},
			nil,
			errTagsMapAssertion,
		},
		{
			"Test writing metrics with correct data - failed",
			map[string]interface{}{
				"series_1": map[string]interface{}{"field_a": 11, "field_b": 12},
				"series_2": map[string]interface{}{"field_c": 21, "field_d": 22},
			},
			map[string]interface{}{
				"series_1": map[string]interface{}{"tag_a": "a", "tag_b": "b"},
				"series_2": map[string]interface{}{"tag_c": "c", "tag_d": "d"},
			},
			errWriteMetrics,
			errWriteMetrics,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			influxClientMock := &mocks.Client{}
			client := Client{
				client:       influxClientMock,
				ctx:          context.Background(),
				organization: "organization",
				bucket:       "piper",
			}
			writeAPIBlockingMock := &mocks.WriteAPIBlocking{}
			writeAPIBlockingMock.On("WritePoint", client.ctx, mock.Anything).Return(tt.writePointErr)
			influxClientMock.On("WriteAPIBlocking", client.organization, client.bucket).Return(writeAPIBlockingMock)
			err := client.WriteMetrics(tt.dataMap, tt.dataMapTags)
			if err != tt.err {
				t.Errorf("\nactual: %q\nexpected: %q\n", err, tt.err)
			}
		})
	}

}
