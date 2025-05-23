package npm

import (
	"fmt"
	"os"

	"github.com/SAP/jenkins-library/pkg/log"
)

// Tool encapsulates the commands and configuration for a package manager tool.
type Tool struct {
	Name           string
	ExecRunner     ExecRunner
	InstallCmd     []string
	RunCmd         []string
	PublishCmd     []string
	PublishFlags   []string
	PackCmd        []string
	ConfigGetFlags []string
	ConfigSetFlags []string
	configFiles    map[string]string // Map of original file paths to backup paths
	Utils          Utils             // For file operations
}

// ConfigFile represents a configuration file that needs backup/restore
type configFile struct {
	path        string
	backupPath  string
	needsBackup bool
}

// getToolPath returns a consistent path for a tool in the local installation directory
func getToolPath(toolName string) string {
	return npmInstallationFolder + "/node_modules/.bin/" + toolName
}

// getBinaryPath returns the path to the tool's binary
func (t *Tool) GetBinaryPath() string {
	if t.Name == "yarn" || t.Name == "pnpm" {
		return getToolPath(t.Name)
	}
	return t.Name
}

// Install runs the install command for the tool.
func (t *Tool) Install() error {
	return t.ExecRunner.RunExecutable(t.GetBinaryPath(), t.InstallCmd...)
}

// Run runs the run command for the tool with additional arguments.
func (t *Tool) Run(args ...string) error {
	cmd := append(t.RunCmd, args...)
	return t.ExecRunner.RunExecutable(t.GetBinaryPath(), cmd...)
}

// initConfigFiles initializes the configuration files map based on the tool
func (t *Tool) initConfigFiles() {
	if t.configFiles == nil {
		t.configFiles = make(map[string]string)
	}

	// Always check .npmrc
	npmrcPath := ".npmrc"
	if exists(npmrcPath, t.Utils) {
		t.configFiles[npmrcPath] = npmrcPath + ".bak"
	}

	// For pnpm, also check workspace file
	if t.Name == "pnpm" {
		workspacePath := "pnpm-workspace.yaml"
		if exists(workspacePath, t.Utils) {
			t.configFiles[workspacePath] = workspacePath + ".bak"
		}
	}
}

// backupConfigFiles creates backups of all configuration files
func (t *Tool) backupConfigFiles() error {
	t.initConfigFiles()
	for orig, backup := range t.configFiles {
		content, err := t.Utils.FileRead(orig)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", orig, err)
		}
		if err := t.Utils.FileWrite(backup, content, 0644); err != nil {
			return fmt.Errorf("failed to create backup %s: %w", backup, err)
		}
		log.Entry().Debugf("Created backup of %s at %s", orig, backup)
	}
	return nil
}

// restoreConfigFiles restores all configuration files from backups
func (t *Tool) restoreConfigFiles() error {
	for orig, backup := range t.configFiles {
		if exists(backup, t.Utils) {
			content, err := t.Utils.FileRead(backup)
			if err != nil {
				return fmt.Errorf("failed to read backup %s: %w", backup, err)
			}
			if err := t.Utils.FileWrite(orig, content, 0644); err != nil {
				return fmt.Errorf("failed to restore %s: %w", orig, err)
			}
			if err := t.Utils.FileRemove(backup); err != nil {
				log.Entry().Warnf("Failed to remove backup file %s: %v", backup, err)
			}
			log.Entry().Debugf("Restored %s from backup", orig)
		}
	}
	return nil
}

// Set os env
func (t *Tool) setOSEnv(args ...string) {
	env := os.Environ()
	env = append(env, "NPM_CONFIG_USERCONFIG=.npmrc")
	env = append(env, args...)
	t.ExecRunner.SetEnv(env)
}

// Publish runs the publish command for the tool.
func (t *Tool) Publish(args ...string) error {
	t.setOSEnv()
	cmd := append(t.PublishCmd, t.PublishFlags...)
	cmd = append(cmd, args...)
	return t.ExecRunner.RunExecutable(t.GetBinaryPath(), cmd...)
}

// Pack runs the pack command for the tool.
func (t *Tool) Pack(args ...string) error {
	cmd := append(t.PackCmd, args...)
	return t.ExecRunner.RunExecutable(t.GetBinaryPath(), cmd...)
}

// GetRegistry returns the registry URL for the tool.
func (t *Tool) GetRegistry(args ...string) error {
	cmd := []string{"config", "get", "registry"}
	cmd = append(cmd, t.ConfigGetFlags...)
	cmd = append(cmd, args...)
	return t.ExecRunner.RunExecutable(t.GetBinaryPath(), cmd...)
}

// SetRegistry configures the registry URL and authentication for the tool.
func (t *Tool) SetRegistry(registry, username, password, scope string) error {
	if err := t.backupConfigFiles(); err != nil {
		return err
	}
	defer func() {
		if err := t.restoreConfigFiles(); err != nil {
			log.Entry().Warnf("Failed to restore configuration files: %v", err)
		}
	}()

	// Set registry URL
	cmd := []string{"config", "set", "registry", registry}
	cmd = append(cmd, t.ConfigSetFlags...)
	if err := t.ExecRunner.RunExecutable(t.GetBinaryPath(), cmd...); err != nil {
		return fmt.Errorf("failed to set registry: %w", err)
	}

	// Set scoped registry if provided
	if scope != "" {
		cmd := []string{"config", "set", scope + ":registry", registry}
		cmd = append(cmd, t.ConfigSetFlags...)
		if err := t.ExecRunner.RunExecutable(t.GetBinaryPath(), cmd...); err != nil {
			return fmt.Errorf("failed to set scoped registry: %w", err)
		}
	}

	// Set authentication if provided
	if username != "" && password != "" {
		authToken := fmt.Sprintf("%s:%s", username, password)
		if err := t.ExecRunner.RunExecutable(t.GetBinaryPath(), "config", "set", registry+"/_auth", authToken); err != nil {
			return fmt.Errorf("failed to set authentication: %w", err)
		}
		if err := t.ExecRunner.RunExecutable(t.GetBinaryPath(), "config", "set", "always-auth", "true"); err != nil {
			return fmt.Errorf("failed to set always-auth: %w", err)
		}
	}

	// list config
	cmd = []string{"config", "list"}
	cmd = append(cmd, t.ConfigGetFlags...)
	if err := t.ExecRunner.RunExecutable(t.GetBinaryPath(), cmd...); err != nil {
		return fmt.Errorf("failed to list config: %w", err)
	}

	return nil
}

// DetectTool inspects the current directory for lockfiles, auto-installs the tool if needed,
// and returns the ready-to-use Tool struct. For specific tools (yarn/pnpm), it handles installation.
// It warns if a lock file is missing for the selected tool.
func DetectTool(utils Utils, toolName string) (*Tool, error) {
	var (
		ToolNPM = Tool{
			Name:           "npm",
			InstallCmd:     []string{"ci"},
			RunCmd:         []string{"run"},
			PublishCmd:     []string{"publish"},
			PublishFlags:   []string{""},
			PackCmd:        []string{"pack"},
			ConfigGetFlags: []string{"-ws=false", "-iwr"},
			ConfigSetFlags: []string{"-ws=false", "-iwr"},
			Utils:          utils,
			configFiles:    make(map[string]string),
		}
		ToolYarn = Tool{
			Name:           "yarn",
			InstallCmd:     []string{"install", "--frozen-lockfile"},
			RunCmd:         []string{"run"},
			PublishCmd:     []string{"publish"},
			PublishFlags:   []string{"----non-interactive"},
			PackCmd:        []string{"pack"},
			ConfigGetFlags: []string{},
			ConfigSetFlags: []string{},
			Utils:          utils,
			configFiles:    make(map[string]string),
		}
		ToolPNPM = Tool{
			Name:           "pnpm",
			InstallCmd:     []string{"install"},
			RunCmd:         []string{"run"},
			PublishCmd:     []string{"publish"},
			PublishFlags:   []string{"--no-git-checks"},
			PackCmd:        []string{"pack"},
			ConfigGetFlags: []string{"--location", "project"},
			ConfigSetFlags: []string{"--location", "project"},
			Utils:          utils,
			configFiles:    make(map[string]string),
		}
	)
	execRunner := utils.GetExecRunner()
	var tool Tool

	// Handle specific tool requests first
	switch toolName {
	case "pnpm":
		if !exists("pnpm-lock.yaml", utils) {
			log.Entry().Warning("No pnpm-lock.yaml found. Please run pnpm install locally and commit the lock file.")
		}
		if err := autoInstallTool(execRunner, "pnpm"); err != nil {
			return nil, err
		}
		tool = ToolPNPM

	case "yarn":
		if !exists("yarn.lock", utils) {
			log.Entry().Warning("No yarn.lock found. Please run yarn install locally and commit the lock file.")
		}
		if err := autoInstallTool(execRunner, "yarn"); err != nil {
			return nil, err
		}
		tool = ToolYarn

	case "auto":
		// Auto-detect based on lock files
		switch {
		case exists("pnpm-lock.yaml", utils):
			if err := autoInstallTool(execRunner, "pnpm"); err != nil {
				return nil, err
			}
			tool = ToolPNPM
		case exists("yarn.lock", utils):
			if err := autoInstallTool(execRunner, "yarn"); err != nil {
				return nil, err
			}
			tool = ToolYarn
		case exists("package-lock.json", utils):
			tool = ToolNPM
		default:
			log.Entry().Warning("No lock file found. Please run install locally and commit the lock file.")
			tool = ToolNPM
			tool.InstallCmd = []string{"install"}
		}

	default:
		tool = ToolNPM
		if !exists("package-lock.json", utils) {
			log.Entry().Warning("No package-lock.json found. Please run npm install locally and commit the lock file.")
			tool.InstallCmd = []string{"install"}
		}
	}

	tool.ExecRunner = execRunner
	return &tool, nil
}

// autoInstallTool installs the given tool locally in the tmp directory if not already present.
func autoInstallTool(execRunner ExecRunner, toolName string) error {
	// Keep relative path for tests and CI compatibility
	binPath := getToolPath(toolName)
	if _, err := os.Stat(binPath); err == nil {
		return nil
	}

	// Save current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	log.Entry().Infof("Current directory: %s", currentDir)

	// Install tool locally in tmp directory
	err = execRunner.RunExecutable("npm", "install", toolName, "--prefix", npmInstallationFolder)
	if err != nil {
		return fmt.Errorf("failed to install required tool '%s': %w", toolName, err)
	}

	// Return to original directory
	if err := os.Chdir(currentDir); err != nil {
		return fmt.Errorf("failed to return to original directory: %w", err)
	}

	currentDir, err = os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}
	log.Entry().Infof("Tool %s installed successfully, working dir: %s", toolName, currentDir)

	return nil
}

// exists checks if a file exists in the current working directory.
func exists(filename string, utils Utils) bool {
	exists, err := utils.FileExists(filename)
	if err != nil {
		log.Entry().Fatalf("Error checking for file %s: %v\n", filename, err)
		return false
	}
	return exists
}
