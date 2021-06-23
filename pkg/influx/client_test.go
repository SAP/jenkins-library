package influx

import (
	"errors"
	"testing"

	"github.com/SAP/jenkins-library/pkg/influx/mocks"
	"github.com/stretchr/testify/mock"
)

func TestWriteMetrics(t *testing.T) {
	errWriteMetrics := errors.New("error")
	tests := []struct {
		name          string
		dataMap       map[string]map[string]interface{}
		dataMapTags   map[string]map[string]string
		writePointErr error
		err           error
	}{
		{
			"Test writing metrics - success",
			map[string]map[string]interface{}{
				"series_1": {"field_a": 11, "field_b": 12},
				"series_2": {"field_c": 21, "field_d": 22},
			},
			map[string]map[string]string{
				"series_1": {"tag_a": "a", "tag_b": "b"},
				"series_2": {"tag_c": "c", "tag_d": "d"},
			},
			nil,
			nil,
		},
		{
			"Test writing metrics - failed",
			map[string]map[string]interface{}{
				"series_1": {"field_a": 11, "field_b": 12},
				"series_2": {"field_c": 21, "field_d": 22},
			},
			map[string]map[string]string{
				"series_1": {"tag_a": "a", "tag_b": "b"},
				"series_2": {"tag_c": "c", "tag_d": "d"},
			},
			errWriteMetrics,
			errWriteMetrics,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			influxClientMock := &mocks.Client{}
			client := NewClient(influxClientMock, "org", "piper")
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
