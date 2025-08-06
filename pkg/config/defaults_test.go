package config

import (
	"io"
	"strings"
	"testing"
)

func TestReadPipelineDefaults(t *testing.T) {

	var d PipelineDefaults

	t.Run("Success case", func(t *testing.T) {
		d0 := strings.NewReader("general:\n  testStepKey1: testStepValue1")
		d1 := strings.NewReader("general:\n  testStepKey2: testStepValue2")
		err := d.ReadPipelineDefaults([]io.ReadCloser{io.NopCloser(d0), io.NopCloser(d1)})

		if err != nil {
			t.Errorf("Got error although no error expected: %v", err)
		}

		t.Run("Defaults 0", func(t *testing.T) {
			expected := "testStepValue1"
			if d.Defaults[0].General["testStepKey1"] != expected {
				t.Errorf("got: %v, expected: %v", d.Defaults[0].General["testStepKey1"], expected)
			}
		})

		t.Run("Defaults 1", func(t *testing.T) {
			expected := "testStepValue2"
			if d.Defaults[1].General["testStepKey2"] != expected {
				t.Errorf("got: %v, expected: %v", d.Defaults[1].General["testStepKey2"], expected)
			}
		})
	})

	t.Run("Read failure", func(t *testing.T) {
		var rc errReadCloser
		err := d.ReadPipelineDefaults([]io.ReadCloser{rc})
		if err == nil {
			t.Errorf("Got no error although error expected.")
		}
	})

	t.Run("Unmarshalling failure", func(t *testing.T) {
		myConfig := strings.NewReader("general:\n\ttestStepKey: testStepValue")
		err := d.ReadPipelineDefaults([]io.ReadCloser{io.NopCloser(myConfig)})
		if err == nil {
			t.Errorf("Got no error although error expected.")
		}
	})
}
