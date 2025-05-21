//go:build unit
// +build unit

package npm

import (
	"io"
	"testing"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/stretchr/testify/assert"
)

type mockExecRunner struct {
	calls            []execCall
	lookPathResponse string
	lookPathError    error
}

type execCall struct {
	executable string
	params     []string
}

func (m *mockExecRunner) RunExecutable(executable string, params ...string) error {
	m.calls = append(m.calls, execCall{executable: executable, params: params})
	return nil
}

func (m *mockExecRunner) RunExecutableInBackground(executable string, params ...string) (command.Execution, error) {
	return nil, nil
}

func (m *mockExecRunner) LookPath(bin string) (string, error) {
	return m.lookPathResponse, m.lookPathError
}

func (m *mockExecRunner) SetEnv(e []string)    {}
func (m *mockExecRunner) Stdout(out io.Writer) {}
func (m *mockExecRunner) Stderr(out io.Writer) {}

func TestToolInstallation(t *testing.T) {
	tests := []struct {
		name     string
		toolName string
		want     []execCall
	}{
		{
			name:     "install yarn locally",
			toolName: "yarn",
			want: []execCall{
				{
					executable: "npm",
					params:     []string{"install", "yarn", "--prefix", "./tmp"},
				},
			},
		},
		{
			name:     "install pnpm locally",
			toolName: "pnpm",
			want: []execCall{
				{
					executable: "npm",
					params:     []string{"install", "pnpm", "--prefix", "./tmp"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execRunner := &mockExecRunner{}
			err := autoInstallTool(execRunner, tt.toolName)
			assert.NoError(t, err)
			assert.Len(t, execRunner.calls, len(tt.want))

			for i, call := range execRunner.calls {
				assert.Equal(t, tt.want[i].executable, call.executable)
				assert.Equal(t, tt.want[i].params, call.params)
			}
		})
	}
}

func TestToolExecution(t *testing.T) {
	tests := []struct {
		name       string
		tool       Tool
		operation  string
		args       []string
		wantExec   string
		wantParams []string
	}{
		{
			name:       "yarn install uses local binary",
			tool:       Tool{Name: "yarn", InstallCmd: []string{"install", "--frozen-lockfile"}},
			operation:  "install",
			wantExec:   "./tmp/node_modules/.bin/yarn",
			wantParams: []string{"install", "--frozen-lockfile"},
		},
		{
			name: "pnpm publish with tarball",
			tool: Tool{
				Name:         "pnpm",
				PublishCmd:   []string{"publish"},
				PublishFlags: []string{"--config", ".piperNpmrc"},
				PackCmd:      []string{"pack"},
			},
			operation:  "publish",
			args:       []string{"--tarball", "package.tgz", "--registry", "https://registry.example.com"},
			wantExec:   "./tmp/node_modules/.bin/pnpm",
			wantParams: []string{"publish", "--config", ".piperNpmrc", "--tarball", "package.tgz", "--registry", "https://registry.example.com"},
		},
		{
			name: "pnpm pack command",
			tool: Tool{
				Name:    "pnpm",
				PackCmd: []string{"pack"},
			},
			operation:  "pack",
			wantExec:   "./tmp/node_modules/.bin/pnpm",
			wantParams: []string{"pack"},
		},
		{
			name: "pnpm publish uses local binary and flags",
			tool: Tool{
				Name:         "pnpm",
				PublishCmd:   []string{"publish"},
				PublishFlags: []string{"--config", ".piperNpmrc"},
			},
			operation:  "publish",
			args:       []string{"--registry", "https://registry.example.com"},
			wantExec:   "./tmp/node_modules/.bin/pnpm",
			wantParams: []string{"publish", "--config", ".piperNpmrc", "--registry", "https://registry.example.com"},
		},
		{
			name:       "pnpm run uses local binary",
			tool:       Tool{Name: "pnpm", RunCmd: []string{"run"}},
			operation:  "run",
			args:       []string{"build"},
			wantExec:   "./tmp/node_modules/.bin/pnpm",
			wantParams: []string{"run", "build"},
		},
		{
			name:       "npm uses global binary",
			tool:       Tool{Name: "npm", RunCmd: []string{"run"}},
			operation:  "run",
			args:       []string{"test"},
			wantExec:   "npm",
			wantParams: []string{"run", "test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			execRunner := &mockExecRunner{}
			tt.tool.ExecRunner = execRunner

			var err error
			switch tt.operation {
			case "install":
				err = tt.tool.Install()
			case "run":
				err = tt.tool.Run(tt.args...)
			case "publish":
				err = tt.tool.Publish(tt.args...)
			case "pack":
				err = tt.tool.Pack()
			}

			assert.NoError(t, err)
			assert.Len(t, execRunner.calls, 1)

			call := execRunner.calls[0]
			assert.Equal(t, tt.wantExec, call.executable)
			assert.Equal(t, tt.wantParams, call.params)
		})
	}
}
