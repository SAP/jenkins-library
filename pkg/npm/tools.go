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
	AddCmd     []string
	RemoveCmd  []string
	TestCmd    []string
	BuildCmd   []string
}

var (
	ToolNPM = Tool{
		Name:       "npm",
		InstallCmd: []string{"ci"},
		RunCmd:     []string{"run"},
		PublishCmd: []string{"publish"},
		AddCmd:     []string{"install"},
		RemoveCmd:  []string{"uninstall"},
		TestCmd:    []string{"run", "test"},
		BuildCmd:   []string{"run", "build"},
	}
	ToolYarn = Tool{
		Name:       "yarn",
		InstallCmd: []string{"install", "--frozen-lockfile"},
		RunCmd:     []string{"run"},
		PublishCmd: []string{"publish"},
		AddCmd:     []string{"add"},
		RemoveCmd:  []string{"remove"},
		TestCmd:    []string{"run", "test"},
		BuildCmd:   []string{"run", "build"},
	}
	ToolPNPM = Tool{
		Name:       "pnpm",
		InstallCmd: []string{"install"},
		RunCmd:     []string{"run"},
		PublishCmd: []string{"publish"},
		AddCmd:     []string{"add"},
		RemoveCmd:  []string{"remove"},
		TestCmd:    []string{"run", "test"},
		BuildCmd:   []string{"run", "build"},
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

// Add runs the add command for the tool.
func (t *Tool) Add(args ...string) error {
	cmd := append(t.AddCmd, args...)
	return t.ExecRunner.RunExecutable(t.Name, cmd...)
}

// Remove runs the remove command for the tool.
func (t *Tool) Remove(args ...string) error {
	cmd := append(t.RemoveCmd, args...)
	return t.ExecRunner.RunExecutable(t.Name, cmd...)
}

// Test runs the test command for the tool.
func (t *Tool) Test(args ...string) error {
	cmd := append(t.TestCmd, args...)
	return t.ExecRunner.RunExecutable(t.Name, cmd...)
}

// Build runs the build command for the tool.
func (t *Tool) Build(args ...string) error {
	cmd := append(t.BuildCmd, args...)
	return t.ExecRunner.RunExecutable(t.Name, cmd...)
}

// DetectTool inspects the current directory for lockfiles, auto-installs the tool if needed,
// runs install, and returns the ready-to-use Tool struct.
func DetectTool(utils Utils, toolName string) (*Tool, error) {
	execRunner := utils.GetExecRunner()
	var tool Tool

	// First handle specific tool requests
	switch toolName {
	case "pnpm":
		if err := autoInstallTool(execRunner, "pnpm"); err != nil {
			return nil, err
		}
		tool = ToolPNPM
	case "yarn":
		if err := autoInstallTool(execRunner, "yarn"); err != nil {
			return nil, err
		}
		tool = ToolYarn
	case "auto":
		// Auto-detect based on lock files
		if exists("pnpm-lock.yaml", utils) {
			if err := autoInstallTool(execRunner, "pnpm"); err != nil {
				return nil, err
			}
			tool = ToolPNPM
		} else if exists("yarn.lock", utils) {
			if err := autoInstallTool(execRunner, "yarn"); err != nil {
				return nil, err
			}
			tool = ToolYarn
		} else if exists("package-lock.json", utils) {
			tool = ToolNPM
		} else {
			tool = ToolNPM
			tool.InstallCmd = []string{"install"}
		}
	default:
		tool = ToolNPM
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
