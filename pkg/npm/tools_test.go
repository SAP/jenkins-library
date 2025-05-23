//go:build unit
// +build unit

package npm

import (
	"io"
	"testing"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/mock"
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
				PublishFlags: []string{"--no-git-checks"},
				PackCmd:      []string{"pack"},
			},
			operation:  "publish",
			args:       []string{"--tarball", "package.tgz", "--registry", "https://registry.example.com"},
			wantExec:   "./tmp/node_modules/.bin/pnpm",
			wantParams: []string{"publish", "--no-git-checks", "--tarball", "package.tgz", "--registry", "https://registry.example.com"},
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
				PublishFlags: []string{"--no-git-checks"},
			},
			operation:  "publish",
			args:       []string{"--registry", "https://registry.example.com"},
			wantExec:   "./tmp/node_modules/.bin/pnpm",
			wantParams: []string{"publish", "--no-git-checks", "--registry", "https://registry.example.com"},
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

func TestConfigurationManagement(t *testing.T) {
	t.Run("backup and restore npmrc", func(t *testing.T) {
		// Setup
		filesMock := newNpmMockUtilsBundle()
		filesMock.AddFile(".npmrc", []byte("registry=https://registry.npmjs.org/"))

		tool := Tool{
			Name:  "npm",
			Utils: &filesMock,
		}

		// Test backup
		err := tool.backupConfigFiles()
		assert.NoError(t, err)

		// Check backup exists
		assert.True(t, filesMock.HasFile(".npmrc.bak"))
		backup, err := filesMock.FileRead(".npmrc.bak")
		assert.NoError(t, err)
		assert.Equal(t, []byte("registry=https://registry.npmjs.org/"), backup)

		// Test restore
		err = tool.restoreConfigFiles()
		assert.NoError(t, err)

		// Check original is restored and backup is removed
		assert.True(t, filesMock.HasFile(".npmrc"))
		assert.False(t, filesMock.HasFile(".npmrc.bak"))

		restored, err := filesMock.FileRead(".npmrc")
		assert.NoError(t, err)
		assert.Equal(t, []byte("registry=https://registry.npmjs.org/"), restored)
	})

	t.Run("handle pnpm workspace config", func(t *testing.T) {
		// Setup with both npmrc and workspace file
		filesMock := newNpmMockUtilsBundle()
		filesMock.AddFile(".npmrc", []byte("registry=https://registry.npmjs.org/"))
		filesMock.AddFile("pnpm-workspace.yaml", []byte("packages:\n  - 'packages/*'"))

		tool := Tool{
			Name:  "pnpm",
			Utils: &filesMock,
		}

		// Test backup
		err := tool.backupConfigFiles()
		assert.NoError(t, err)

		// Check both backups exist
		assert.True(t, filesMock.HasFile(".npmrc.bak"))
		assert.True(t, filesMock.HasFile("pnpm-workspace.yaml.bak"))

		npmrcBackup, err := filesMock.FileRead(".npmrc.bak")
		assert.NoError(t, err)
		assert.Equal(t, []byte("registry=https://registry.npmjs.org/"), npmrcBackup)

		workspaceBackup, err := filesMock.FileRead("pnpm-workspace.yaml.bak")
		assert.NoError(t, err)
		assert.Equal(t, []byte("packages:\n  - 'packages/*'"), workspaceBackup)
	})

	t.Run("set registry without auth", func(t *testing.T) {
		// Setup
		execRunner := &mockExecRunner{}
		filesMock := newNpmMockUtilsBundle()
		filesMock.AddFile(".npmrc", []byte("registry=https://old-registry.npmjs.org/"))

		tool := Tool{
			Name:       "npm",
			ExecRunner: execRunner,
			Utils:      &filesMock,
		}

		// Test SetRegistry without auth
		err := tool.SetRegistry(
			"https://test-registry.com",
			"",
			"",
			"@test-scope",
		)
		assert.NoError(t, err)

		// Verify correct npm config commands were called
		expectedCalls := []execCall{
			{executable: "npm", params: []string{"config", "set", "registry", "https://test-registry.com"}},
			{executable: "npm", params: []string{"config", "set", "@test-scope:registry", "https://test-registry.com"}},
		}

		assert.Len(t, execRunner.calls, len(expectedCalls))
		for i, call := range execRunner.calls {
			assert.Equal(t, expectedCalls[i].executable, call.executable)
			assert.Equal(t, expectedCalls[i].params, call.params)
		}
	})

	t.Run("set registry with auth", func(t *testing.T) {
		// Setup
		utils := newNpmMockUtilsBundle()
		utils.AddFile(".npmrc", []byte("registry=https://old-registry.npmjs.org/"))

		tool := Tool{
			Name:       "npm",
			ExecRunner: utils.execRunner,
			Utils:      &utils,
		}

		// Test SetRegistry with auth
		err := tool.SetRegistry(
			"https://test-registry.com",
			"testuser",
			"testpass",
			"@test-scope",
		)
		assert.NoError(t, err)

		// Verify correct npm config commands were called
		expectedCalls := []mock.ExecCall{
			{Exec: "npm", Params: []string{"config", "set", "registry", "https://test-registry.com"}},
			{Exec: "npm", Params: []string{"config", "set", "@test-scope:registry", "https://test-registry.com"}},
			{Exec: "npm", Params: []string{"config", "set", "https://test-registry.com/_auth", "testuser:testpass"}},
			{Exec: "npm", Params: []string{"config", "set", "always-auth", "true"}},
		}

		assert.Len(t, len(utils.execRunner.Calls), len(expectedCalls))
		for i, call := range utils.execRunner.Calls {
			assert.Equal(t, call.Exec, expectedCalls[i].Exec)
		}

		// Verify backup was created and cleaned up
		assert.False(t, utils.HasFile(".npmrc.bak"))
	})
}
