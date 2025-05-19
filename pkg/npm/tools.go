package npm

import (
	"fmt"

	"github.com/SAP/jenkins-library/pkg/log"
)

// Tool encapsulates the commands and configuration for a package manager tool.
type Tool struct {
	Name       string
	ExecRunner ExecRunner
	InstallCmd []string
	RunCmd     []string
	PublishCmd []string
}

var (
	ToolNPM = Tool{
		Name:       "npm",
		InstallCmd: []string{"ci"},
		RunCmd:     []string{"run"},
		PublishCmd: []string{"publish"},
	}
	ToolYarn = Tool{
		Name:       "yarn",
		InstallCmd: []string{"install", "--frozen-lockfile"},
		RunCmd:     []string{"run"},
		PublishCmd: []string{"publish"},
	}
	ToolPNPM = Tool{
		Name:       "pnpm",
		InstallCmd: []string{"install"},
		RunCmd:     []string{"run"},
		PublishCmd: []string{"publish"},
	}
)

// Install runs the install command for the tool.
func (t *Tool) Install() error {
	return t.ExecRunner.RunExecutable(t.Name, t.InstallCmd...)
}

// Run runs the run command for the tool with additional arguments.
func (t *Tool) Run(args ...string) error {
	cmd := append(t.RunCmd, args...)
	return t.ExecRunner.RunExecutable(t.Name, cmd...)
}

// Publish runs the publish command for the tool.
func (t *Tool) Publish(args ...string) error {
	cmd := append(t.PublishCmd, args...)
	return t.ExecRunner.RunExecutable(t.Name, cmd...)
}

// DetectTool inspects the current directory for lockfiles, auto-installs the tool if needed,
// and returns the ready-to-use Tool struct. For specific tools (yarn/pnpm), it handles installation.
// It warns if a lock file is missing for the selected tool.
func DetectTool(utils Utils, toolName string) (*Tool, error) {
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

// autoInstallTool installs the given tool globally if not already present (for yarn/pnpm).
func autoInstallTool(execRunner ExecRunner, toolName string) error {
	_, err := execRunner.LookPath(toolName)
	if err == nil {
		return nil
	}
	err = execRunner.RunExecutable("npm", "install", "-g", toolName)
	if err != nil {
		return fmt.Errorf("failed to install required tool '%s': %w", toolName, err)
	}
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
