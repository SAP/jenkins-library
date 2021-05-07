package versioning

import (
	"testing"
)

func Test_applyVersioningModel(t *testing.T) {
	type args struct {
		model   string
		version Coordinates
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"maven - major", args{VersioningModelMajor, Coordinates{Version: "1.2.3-7864387648746"}}, "1"},
		{"maven - major-minor", args{VersioningModelMajorMinor, Coordinates{Version: "1.2.3-7864387648746"}}, "1.2"},
		{"maven - semantic", args{VersioningModelSemantic, Coordinates{Version: "1.2.3-7864387648746"}}, "1.2.3"},
		{"maven - full", args{VersioningModelFull, Coordinates{Version: "1.2.3-7864387648746"}}, "1.2.3-7864387648746"},
		{"python - major-minor", args{VersioningModelMajorMinor, Coordinates{Version: "2.2.3.20200101"}}, "2.2"},
		{"leading zero", args{VersioningModelMajor, Coordinates{Version: "0.0.1"}}, "0"},
		{"trailing zero", args{VersioningModelMajorMinor, Coordinates{Version: "2.0"}}, "2.0"},
		{"invalid - unknown versioning model", args{"snapshot", Coordinates{Version: "1.2.3-SNAPSHOT"}}, ""},
		{"invalid - incorrect version", args{VersioningModelMajor, Coordinates{Version: ".2.3"}}, ""},
		{"invalid - version to short", args{VersioningModelSemantic, Coordinates{Version: "1.2"}}, "1.2.<no value>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ApplyVersioningModel(tt.args.model, tt.args.version)
			if got != tt.want {
				t.Errorf("ApplyVersioningModel() = %v, want %v", got, tt.want)
			}
		})
	}
}
