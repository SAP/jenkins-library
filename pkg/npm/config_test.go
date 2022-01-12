package npm

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/magiconair/properties"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewNPMRC(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{name: "current dir", args: args{""}, want: configFilename},
		{name: "sub dir", args: args{mock.Anything}, want: filepath.Join(mock.Anything, ".piperStagingNpmrc")},
		{name: "file path in current dir", args: args{".piperStagingNpmrc"}, want: ".piperStagingNpmrc"},
		{name: "file path in sub dir", args: args{filepath.Join(mock.Anything, ".piperStagingNpmrc")}, want: filepath.Join(mock.Anything, ".piperStagingNpmrc")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewNPMRC(tt.args.path); !reflect.DeepEqual(got.filepath, tt.want) {
				t.Errorf("NewNPMRC().filepath = %v, want %v", got.filepath, tt.want)
			}
		})
	}
}

func mockLoadProperties(t *testing.T, result *properties.Properties, err error) func(filename string, enc properties.Encoding) (*properties.Properties, error) {
	return func(filename string, enc properties.Encoding) (*properties.Properties, error) {
		return result, err
	}
}

func TestLoad(t *testing.T) {
	// init
	config := NewNPMRC("")

	new := properties.NewProperties()
	new.Set("test", "anything")
	propertiesLoadFile = mockLoadProperties(t, new, nil)
	require.NotEmpty(t, new.Keys())

	require.Empty(t, config.values.Keys())
	// test
	err := config.Load()
	// assert
	assert.NoError(t, err)
	assert.NotEmpty(t, config.values.Keys())

}
