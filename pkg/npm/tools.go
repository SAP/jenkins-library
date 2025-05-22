package npm

import (
	"fmt"
	"os"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/npm/rc"
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
	GetRegistryCmd []string
	SetRegistryCmd []string
	RC             RCManager
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

// Publish runs the publish command for the tool.
func (t *Tool) Publish(args ...string) error {
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
	cmd := append(t.GetRegistryCmd, args...)
	return t.ExecRunner.RunExecutable(t.GetBinaryPath(), cmd...)
}

// SetRegistry sets the registry URL for the tool.
func (t *Tool) SetRegistry(registry string, args ...string) error {
	cmd := append(t.SetRegistryCmd, registry)
	if len(args) > 0 {
		cmd = append(cmd, args...)
	}
	return t.ExecRunner.RunExecutable(t.GetBinaryPath(), cmd...)
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
			PublishFlags:   []string{"--userconfig", ".piperNpmrc"},
			PackCmd:        []string{"pack"},
			GetRegistryCmd: []string{"config", "get", "registry", "-ws=false", "-iwr"},
			SetRegistryCmd: []string{"config", "set", "registry"},
			RC:             rc.NewNPM(".", utils),
		}
		ToolYarn = Tool{
			Name:           "yarn",
			InstallCmd:     []string{"install", "--frozen-lockfile"},
			RunCmd:         []string{"run"},
			PublishCmd:     []string{"publish"},
			PublishFlags:   []string{"--non-interactive"},
			PackCmd:        []string{"pack"},
			GetRegistryCmd: []string{"config", "get", "registry"},
			SetRegistryCmd: []string{"config", "set", "registry"},
			RC:             rc.NewYarn(".", utils),
		}
		ToolPNPM = Tool{
			Name:           "pnpm",
			InstallCmd:     []string{"install"},
			RunCmd:         []string{"run"},
			PublishCmd:     []string{"publish"},
			PublishFlags:   []string{"--no-git-checks"},
			PackCmd:        []string{"pack"},
			GetRegistryCmd: []string{"config", "get", "registry"},
			SetRegistryCmd: []string{"config", "set", "registry"},
			RC:             rc.NewPNPM(".", utils),
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
