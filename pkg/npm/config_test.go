//go:build unit
// +build unit

package npm

import (
	"path/filepath"
	"testing"

	filesmock "github.com/SAP/jenkins-library/pkg/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewNPMRC(t *testing.T) {
	type args struct {
		path string
	}
	type want struct {
		path     string
		errLoad  string
		errWrite string
		files    map[string]string
	}

	tests := []struct {
		name  string
		args  args
		files map[string]string
		want  want
	}{
		{
			name: "success - path pointing to no dir - no content",
			args: args{""},
			files: map[string]string{
				defaultConfigFilename: "",
			},
			want: want{
				path: defaultConfigFilename,
				files: map[string]string{
					defaultConfigFilename: "",
				},
			},
		},
		{
			name: "success - path pointing to cwd - no content",
			args: args{"."},
			files: map[string]string{
				defaultConfigFilename: "",
			},
			want: want{
				path: defaultConfigFilename,
				files: map[string]string{
					defaultConfigFilename: "",
				},
			},
		},
		{
			name: "success - path pointing to sub dir - no content",
			args: args{mock.Anything},
			files: map[string]string{
				filepath.Join(mock.Anything, ".piperNpmrc"): "",
			},
			want: want{
				path: filepath.Join(mock.Anything, ".piperNpmrc"),
				files: map[string]string{
					filepath.Join(mock.Anything, ".piperNpmrc"): "",
				},
			},
		},
		{
			name: "success - path pointing to file in current folder - no content",
			args: args{".piperNpmrc"},
			files: map[string]string{
				".piperNpmrc": "",
			},
			want: want{
				path: ".piperNpmrc",
				files: map[string]string{
					".piperNpmrc": "",
				},
			},
		},
		{
			name: "success - path pointing to file in sub folder - no content",
			args: args{filepath.Join(mock.Anything, ".piperNpmrc")},
			files: map[string]string{
				filepath.Join(mock.Anything, ".piperNpmrc"): "",
			},
			want: want{
				path: filepath.Join(mock.Anything, ".piperNpmrc"),
				files: map[string]string{
					filepath.Join(mock.Anything, ".piperNpmrc"): "",
				},
			},
		},
		{
			name: "success - doesn't modify existing content",
			args: args{filepath.Join(mock.Anything, ".piperNpmrc")},
			files: map[string]string{
				filepath.Join(mock.Anything, ".piperNpmrc"): `
_auth=dGVzdDp0ZXN0
registry=https://my.private.registry/
@piper:registry=https://my.scoped.private.registry/
//my.private.registry/:_auth=dGVzdDp0ZXN0
//my.scoped.private.registry/:_auth=dGVzdDp0ZXN0
`,
			},
			want: want{
				path: filepath.Join(mock.Anything, ".piperNpmrc"),
				files: map[string]string{
					filepath.Join(mock.Anything, ".piperNpmrc"): `
_auth=dGVzdDp0ZXN0
registry=https://my.private.registry/
@piper:registry=https://my.scoped.private.registry/
//my.private.registry/:_auth=dGVzdDp0ZXN0
//my.scoped.private.registry/:_auth=dGVzdDp0ZXN0
`,
				},
			},
		},
		{
			name:  "failure - path pointing to sth which doesnt exist",
			args:  args{"./this/path/doesnt/exist/.piperNpmrc"},
			files: map[string]string{},
			want: want{
				path:    "./this/path/doesnt/exist/.piperNpmrc",
				errLoad: "could not read './this/path/doesnt/exist/.piperNpmrc'",
			},
		},
	}
	for _, tt := range tests {
		files := filesmock.FilesMock{}

		for path, content := range tt.files {
			files.AddFile(path, []byte(content))
		}

		propertiesLoadFile = files.FileRead
		propertiesWriteFile = files.FileWrite

		t.Run(tt.name, func(t *testing.T) {
			uut := NewNPMRC(tt.args.path)

			assert.Equal(t, tt.want.path, uut.filepath)

			if err := uut.Load(); len(tt.want.errLoad) == 0 {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.want.errLoad)
			}

			if err := uut.Write(); len(tt.want.errWrite) == 0 {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.want.errWrite)
			}

			for wantFile, wantContent := range tt.want.files {
				if actualContent, err := files.FileRead(wantFile); assert.NoError(t, err) {
					assert.Equal(t, wantContent, string(actualContent))
				}
			}
		})
	}
}
