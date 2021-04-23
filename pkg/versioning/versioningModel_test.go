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
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{"maven - major", args{VersioningModelMajor, Coordinates{Version: "1.2.3-7864387648746"}}, "1", false},
		{"maven - major-minor", args{VersioningModelMajorMinor, Coordinates{Version: "1.2.3-7864387648746"}}, "1.2", false},
		{"maven - semantic", args{VersioningModelSemantic, Coordinates{Version: "1.2.3-7864387648746"}}, "1.2.3", false},
		{"maven - full", args{VersioningModelFull, Coordinates{Version: "1.2.3-7864387648746"}}, "1.2.3-7864387648746", false},
		{"python - major-minor", args{VersioningModelMajorMinor, Coordinates{Version: "2.2.3.20200101"}}, "2.2", false},
		{"leading zero", args{VersioningModelMajor, Coordinates{Version: "0.0.1"}}, "0", false},
		{"trailing zero", args{VersioningModelMajorMinor, Coordinates{Version: "2.0"}}, "2.0", false},
		{"invalid - unknown versioning model", args{"snapshot", Coordinates{Version: "1.2.3-SNAPSHOT"}}, "", false},
		{"invalid - incorrect version", args{VersioningModelMajor, Coordinates{Version: ".2.3"}}, "", false},
		{"invalid - version to short", args{VersioningModelSemantic, Coordinates{Version: "1.2"}}, "1.2.<no value>", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ApplyVersioningModel(tt.args.model, tt.args.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("ApplyVersioningModel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ApplyVersioningModel() = %v, want %v", got, tt.want)
			}
		})
	}
}
