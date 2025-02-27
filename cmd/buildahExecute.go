package cmd

import (
	"fmt"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/syft"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"
)

func buildahExecute(config buildahExecuteOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *buildahExecuteCommonPipelineEnvironment) {
	c := command.Command{
		ErrorCategoryMapping: map[string][]string{
			log.ErrorBuild.String(): {
				"failed to execute buildah",
			},
		},
		StepName: "buildahExecute",
	}

	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	client := &piperhttp.Client{}
	fileUtils := &piperutils.Files{}

	err := runBuildahExecute(&config, telemetryData, commonPipelineEnvironment, &c, client, fileUtils)
	if err != nil {
		log.Entry().WithError(err).Fatal("Buildah execution failed")
	}
}

func runBuildahExecute(config *buildahExecuteOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *buildahExecuteCommonPipelineEnvironment, execRunner command.ExecRunner, httpClient piperhttp.Sender, fileUtils piperutils.FileUtils) error {
	log.Entry().Info("Starting buildah execution...")
	log.Entry().Infof("Using Dockerfile at: %s", config.DockerfilePath)

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

	log.Entry().Info("Checking system capabilities and configuration...")

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
	err := execRunner.RunExecutable("buildah", "--version")
	if err != nil {
		return errors.Wrap(err, "Failed to execute buildah command")
	}

	// Prepare buildah command with options for container operation
	cmdOpts := []string{
		"bud",                                                      // Using bud (build-using-dockerfile) for Dockerfile builds
		"--format=docker",                                          // Use Docker format for compatibility
		"--security-opt=apparmor=unconfined",                       // Required for container operation
		"--security-opt=seccomp=unconfined",                        // Required for container operation
		"--storage-driver=vfs",                                     // Use vfs storage driver explicitly
		"--pull=true",                                             // Allow pulling base images
		"--layers=true",                                           // Enable layer optimization
		"--volume", "/var/lib/containers:/var/lib/containers:rw,z", // Mount container storage with proper SELinux context
	}

	// Additional build arguments for troubleshooting
	cmdOpts = append(cmdOpts,
		"--log-level=debug",  // Enable debug logging
		"--isolation=chroot", // Use chroot isolation
		"--cap-add=all",      // Grant all capabilities for debugging
	)

	// Add Dockerfile location if specified and different from context
	if config.DockerfilePath != "." && config.DockerfilePath != "" {
		cmdOpts = append(cmdOpts, fmt.Sprintf("-f=%s", config.DockerfilePath))
	}

	// Set up image tag
	imageTag := "latest"
	if config.ContainerImageTag != "" {
		imageTag = config.ContainerImageTag
	}

	// Handle registry and tagging
	if config.ContainerImageName != "" {
		if config.ContainerRegistryURL != "" {
			destination := fmt.Sprintf("%s/%s:%s", config.ContainerRegistryURL, config.ContainerImageName, imageTag)
			cmdOpts = append(cmdOpts, fmt.Sprintf("--tag=%s", destination))

			commonPipelineEnvironment.container.registryURL = config.ContainerRegistryURL
			commonPipelineEnvironment.container.imageNameTag = fmt.Sprintf("%s:%s", config.ContainerImageName, imageTag)
			commonPipelineEnvironment.container.imageNameTags = append(commonPipelineEnvironment.container.imageNameTags, fmt.Sprintf("%s:%s", config.ContainerImageName, imageTag))
			commonPipelineEnvironment.container.imageNames = append(commonPipelineEnvironment.container.imageNames, config.ContainerImageName)

			// Add auth file for registry
			cmdOpts = append(cmdOpts, fmt.Sprintf("--authfile=%s", fmt.Sprintf("%s/config.json", dockerConfigDir)))
		} else {
			cmdOpts = append(cmdOpts, fmt.Sprintf("--tag=%s", fmt.Sprintf("%s:%s", config.ContainerImageName, imageTag)))
		}
	}

	// Add any custom build options
	if len(config.BuildOptions) > 0 {
		cmdOpts = append(cmdOpts, config.BuildOptions...)
	}

	// Log the command being executed (with sensitive data masked)
	displayCmd := []string{}
	for i, arg := range cmdOpts {
		if i > 0 && strings.Contains(arg, "--authfile=") {
			displayCmd = append(displayCmd, "--authfile=****")
		} else {
			displayCmd = append(displayCmd, arg)
		}
	}
	log.Entry().Infof("Executing buildah command: buildah %v", displayCmd)
	err = execRunner.RunExecutable("buildah", cmdOpts...)
	if err != nil {
		log.Entry().Warn("Initial buildah attempt failed, trying fallback configuration...")

		// Fallback options with essential settings
		cmdOpts = []string{
			"bud",
			"--format=docker",
			"--storage-driver=vfs",
			"--isolation=oci", // Try OCI isolation instead of chroot
			"--layers=true",   // Enable layer optimization
			"--pull=true",     // Allow pulling images when needed
		}

		if config.DockerfilePath != "." && config.DockerfilePath != "" {
			cmdOpts = append(cmdOpts, fmt.Sprintf("-f=%s", config.DockerfilePath))
		}

		// Try with default tag
		cmdOpts = append(cmdOpts, "--tag=test-image:latest")

		log.Entry().Infof("Trying fallback buildah command: buildah %v", cmdOpts)
		err = execRunner.RunExecutable("buildah", cmdOpts...)
		if err != nil {
			return fmt.Errorf("buildah build failed with both default and fallback configurations: %w", err)
		}
	}

	// If registry is configured, push the image
	if config.ContainerImageName != "" && config.ContainerRegistryURL != "" {
		log.Entry().Info("Pushing image to registry...")
		pushOpts := []string{
			"push",
			"--authfile", fmt.Sprintf("%s/config.json", dockerConfigDir),
			fmt.Sprintf("%s/%s:%s", config.ContainerRegistryURL, config.ContainerImageName, imageTag),
		}
		err = execRunner.RunExecutable("buildah", pushOpts...)
		if err != nil {
			return fmt.Errorf("failed to push image: %w", err)
		}
	}

	log.Entry().Info("Buildah execution completed")

	if config.CreateBOM {
		log.Entry().Info("Generating bill of materials using syft...")
		if err := syft.GenerateSBOM(config.SyftDownloadURL, dockerConfigDir, execRunner, fileUtils, httpClient, commonPipelineEnvironment.container.registryURL, commonPipelineEnvironment.container.imageNameTags); err != nil {
			return fmt.Errorf("failed to generate BOM: %w", err)
		}
		log.Entry().Info("BOM generation completed")
	}

	return nil
}
