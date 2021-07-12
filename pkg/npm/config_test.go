package npm

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/mock"
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
		{name: "sub dir", args: args{mock.Anything}, want: mock.Anything + "/.npmrc"},
		{name: "file path in current dir", args: args{".npmrc"}, want: ".npmrc"},
		{name: "file path in sub dir", args: args{mock.Anything + "/.npmrc"}, want: mock.Anything + "/.npmrc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewNPMRC(tt.args.path); !reflect.DeepEqual(got.filepath, tt.want) {
				t.Errorf("NewNPMRC().filepath = %v, want %v", got.filepath, tt.want)
			}
		})
	}
}
