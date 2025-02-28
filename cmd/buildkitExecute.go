package cmd

import (
	"fmt"
	"os"

	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/syft"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func buildkitExecute(config buildkitExecuteOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *buildkitExecuteCommonPipelineEnvironment) {
	// for command execution use Command
	c := command.Command{
		ErrorCategoryMapping: map[string][]string{
			log.ErrorBuild.String(): {
				"failed to execute buildctl",
			},
		},
		StepName: "buildkitExecute",
	}

	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	client := &piperhttp.Client{}
	fileUtils := &piperutils.Files{}

	err := runBuildkitExecute(&config, telemetryData, commonPipelineEnvironment, &c, client, fileUtils)
	if err != nil {
		log.Entry().WithError(err).Fatal("Buildkit execution failed")
	}
}

func runBuildkitExecute(config *buildkitExecuteOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *buildkitExecuteCommonPipelineEnvironment, execRunner command.ExecRunner, httpClient piperhttp.Sender, fileUtils piperutils.FileUtils) error {
	log.Entry().Info("Starting buildkit execution in rootful daemonless mode...")
	log.Entry().Infof("Using Dockerfile at: %s", config.DockerfilePath)

	// Set environment for rootful operation
	os.Setenv("BUILDKIT_CLI_MODE", "daemonless")
	os.Setenv("BUILDKIT_PROGRESS", "plain")
	os.Setenv("BUILDKIT_SNAPSHOTTER", "native") // Use native snapshotter instead of overlayfs
	os.Setenv("BUILDKIT_DEBUG", "1")

	// Setup paths and create directories
	basePath := "/home/user/.local/share/buildkit"
	cachePath := fmt.Sprintf("%s/cache", basePath)
	tmpPath := "/tmp"

	// Create directories with proper permissions
	for _, path := range []string{basePath, cachePath, tmpPath, "/var/lib/buildkit", "/var/run/buildkit"} {
		if err := fileUtils.MkdirAll(path, 0777); err != nil {
			log.Entry().Warnf("Failed to create directory %s: %v", path, err)
		}
		// Force permissions even if directory exists
		if err := os.Chmod(path, 0777); err != nil {
			log.Entry().Warnf("Failed to chmod directory %s: %v", path, err)
		}
	}

	// Debug info collection
	log.Entry().Info("Collecting debug information...")

	// System and runtime info
	_ = execRunner.RunExecutable("cat", "/etc/*release")
	_ = execRunner.RunExecutable("uname", "-a")
	_ = execRunner.RunExecutable("cat", "/etc/containers/storage.conf")
	_ = execRunner.RunExecutable("cat", "/proc/self/mountinfo")

	// User and permission checks
	_ = execRunner.RunExecutable("id")
	_ = execRunner.RunExecutable("ls", "-la", "/var/lib/containers")
	_ = execRunner.RunExecutable("ls", "-la", "/")
	_ = execRunner.RunExecutable("mount")

	// Container runtime checks
	_ = execRunner.RunExecutable("capsh", "--print")
	_ = execRunner.RunExecutable("sysctl", "kernel.unprivileged_userns_clone")
	_ = execRunner.RunExecutable("cat", "/proc/sys/user/max_user_namespaces")

	// Security profile checks
	_ = execRunner.RunExecutable("cat", "/proc/self/status")
	_ = execRunner.RunExecutable("cat", "/sys/kernel/security/apparmor/profiles")
	_ = execRunner.RunExecutable("grep", "Seccomp:", "/proc/self/status")
	_ = execRunner.RunExecutable("cat", "/proc/self/attr/current")

	// Storage driver info
	_ = execRunner.RunExecutable("df", "-h")
	_ = execRunner.RunExecutable("findmnt")

	// Network storage checks
	_ = execRunner.RunExecutable("stat", "-f", "/var/lib/containers")
	_ = execRunner.RunExecutable("ls", "-la", "/var/lib/containers/.???*")
	_ = execRunner.RunExecutable("cat", "/proc/mounts")

	// Check buildah version and info
	_ = execRunner.RunExecutable("buildah", "info")
	_ = execRunner.RunExecutable("ps", "aux")

	// Handle Docker authentication
	dockerConfigDir := "/home/user/.docker"
	if len(config.DockerConfigJSON) > 0 {
		dockerConfigJSON, err := fileUtils.FileRead(config.DockerConfigJSON)
		if err != nil {
			return fmt.Errorf("failed to read Docker config: %w", err)
		}
		if err := fileUtils.MkdirAll(dockerConfigDir, 0755); err != nil {
			return fmt.Errorf("failed to create .docker directory: %w", err)
		}
		if err := fileUtils.FileWrite(fmt.Sprintf("%s/config.json", dockerConfigDir), dockerConfigJSON, 0644); err != nil {
			return fmt.Errorf("failed to write Docker config: %w", err)
		}
	}

	// Build with buildkit using direct buildctl
	log.Entry().Info("BuildKit Configuration:")
	log.Entry().Info("- Using privileged mode")
	log.Entry().Infof("- Cache location: %s", cachePath)
	log.Entry().Info("- Temp directory: /tmp")
	log.Entry().Info("- Environment variables:")
	for _, env := range os.Environ() {
		if len(env) > 9 && env[:9] == "BUILDKIT_" {
			log.Entry().Infof("  %s", env)
		}
	}

	// Add buildkit version check
	if err := execRunner.RunExecutable("buildctl-daemonless.sh", "--version"); err != nil {
		log.Entry().Warnf("Failed to get buildkit version: %v", err)
	}

	// Set buildkit-specific env vars for mount and cache handling
	os.Setenv("BUILDKIT_SANDBOX_MOUNT_PATH", "/tmp")
	os.Setenv("BUILDKIT_STEP_MOUNT_PATH", "/tmp")
	os.Setenv("BUILDKIT_SANDBOX_MODE", "0777")
	os.Setenv("BUILDKIT_WORKDIR_MODE", "0777")
	os.Setenv("BUILDKIT_MOUNT_MODE", "0777")

	buildOpts := []string{
		"build",
		"--frontend", "dockerfile.v0",
		"--local", "context=.",
		"--local", fmt.Sprintf("dockerfile=%s", config.DockerfilePath),
		"--progress", "plain",
		"--export-cache", fmt.Sprintf("type=local,dest=%s", cachePath),
		"--import-cache", fmt.Sprintf("type=local,src=%s", cachePath),
	}

	log.Entry().Info("Using build options:", buildOpts)

	// Add build options from config
	buildOpts = append(buildOpts, config.BuildOptions...)

	imageTag := "latest"
	if config.ContainerImageTag != "" {
		imageTag = config.ContainerImageTag
	}
	if config.ContainerImageName != "" && config.ContainerRegistryURL != "" {
		destination := fmt.Sprintf("%s/%s:%s", config.ContainerRegistryURL, config.ContainerImageName, imageTag)
		buildOpts = append(buildOpts, "--output", fmt.Sprintf("type=image,name=%s", destination))

		commonPipelineEnvironment.container.registryURL = config.ContainerRegistryURL
		commonPipelineEnvironment.container.imageNameTag = fmt.Sprintf("%s:%s", config.ContainerImageName, imageTag)
		commonPipelineEnvironment.container.imageNameTags = append(commonPipelineEnvironment.container.imageNameTags, fmt.Sprintf("%s:%s", config.ContainerImageName, imageTag))
		commonPipelineEnvironment.container.imageNames = append(commonPipelineEnvironment.container.imageNames, config.ContainerImageName)
	} else {
		// Build without pushing if no registry/name provided
		buildOpts = append(buildOpts, "--output", "type=docker")
	}

	log.Entry().Info("Executing buildkit build using daemonless mode...")
	if err := execRunner.RunExecutable("buildctl-daemonless.sh", buildOpts...); err != nil {
		return fmt.Errorf("buildkit build failed: %w", err)
	}

	log.Entry().Info("Buildkit execution completed")

	if config.CreateBOM {
		log.Entry().Info("Generating bill of materials using syft...")
		if err := syft.GenerateSBOM(config.SyftDownloadURL, dockerConfigDir, execRunner, fileUtils, httpClient, commonPipelineEnvironment.container.registryURL, commonPipelineEnvironment.container.imageNameTags); err != nil {
			return fmt.Errorf("failed to generate BOM: %w", err)
		}
		log.Entry().Info("BOM generation completed")
	}

	return nil
}
