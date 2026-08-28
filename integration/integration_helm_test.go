//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestHelmIntegration ./integration/...

package main

import (
	"context"
	"strings"
	"testing"

	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/google/uuid"
	"github.com/testcontainers/testcontainers-go"
)

// TestHelmIntegrationPublishWithSBOM exercises the full helmExecute publish path
// end-to-end (Tier 1): package the chart, publish it via HTTP PUT to a local
// WebDAV sink, and generate the chart-level SBOM (bom-helm.xml) plus the
// buildSettingsInfo commonPipelineEnvironment artifact.
//
// The fixture chart is intentionally image-free so helmExecute's container-SBOM
// path skips the syft scan (which would require a network download of the syft
// binary); only the pure-Go chart BOM is produced.
func TestHelmIntegrationPublishWithSBOM(t *testing.T) {
	t.Parallel()
	assert := NewContainerAssert(t)
	ctx := context.Background()

	// A user-defined network lets the piper container reach the chart-repo sink
	// by its alias (targetRepositoryURL: http://chartrepo:80 in the fixture
	// config). Mirrors the wiring in integration_karma_test.go.
	networkName := "helm-sbom-" + uuid.New().String()
	provider, err := testcontainers.ProviderDocker.GetProvider()
	if err != nil {
		t.Fatalf("Failed to get docker provider: %v", err)
	}
	network, err := provider.CreateNetwork(ctx, testcontainers.NetworkRequest{Name: networkName, CheckDuplicate: true})
	if err != nil {
		t.Fatalf("Failed to create network: %v", err)
	}
	defer network.Remove(ctx)

	// bytemark/webdav is a tiny PUT-accepting server; helm publish uploads the
	// chart .tgz via HTTP PUT, matching the Artifactory/Nexus contract.
	chartRepo, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:          "bytemark/webdav",
			Env:            map[string]string{"USERNAME": "repouser", "PASSWORD": "repopass"},
			Networks:       []string{networkName},
			NetworkAliases: map[string][]string{networkName: {"chartrepo"}},
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("Failed to start chart repo container: %v", err)
	}
	defer chartRepo.Terminate(ctx)

	// dtzar/helm-kubectl ships the helm CLI that helmExecute shells out to.
	container := StartPiperContainer(t, ContainerConfig{
		Image:          "dtzar/helm-kubectl",
		TestData:       "TestHelmIntegration/chart",
		WorkDir:        "/chart",
		Networks:       []string{networkName},
		NetworkAliases: []string{"helm-runner"},
	})

	output := RunPiper(t, container, "/chart", "helmExecute")

	// The chart is packaged, published, and the chart SBOM is generated.
	assert.Contains(output, "publishing artifact:")
	assert.Contains(output, "generated chart SBOM bom-helm.xml")
	// The image-free chart means syft is never invoked (no network download).
	assert.Contains(output, "no container images available, skipping syft scan")

	// Publish artifact, chart SBOM, and buildSettingsInfo CPE file all exist.
	assert.FileExists(container,
		"/chart/helm-int-hello-0.1.0.tgz",
		"/chart/bom-helm.xml",
		"/chart/.pipeline/commonPipelineEnvironment/custom/buildSettingsInfo",
	)

	// The chart SBOM is schema-valid CycloneDX 1.4 and describes the chart.
	bomContent := ReadFile(t, container, "/chart/bom-helm.xml")
	assert.NoError(piperutils.ValidateBOM(bomContent), "bom-helm.xml should be a valid CycloneDX BOM")
	schemaVersion, err := piperutils.GetBomSchemaVersionFromContent(bomContent)
	assert.NoError(err, "bom-helm.xml should carry a CycloneDX schema version")
	assert.Equal("1.4", schemaVersion, "chart SBOM must be emitted in CycloneDX schema version 1.4")
	assert.Contains(string(bomContent), "pkg:helm/helm-int-hello@0.1.0", "chart SBOM root component must carry the helm PURL")

	// The container-image SBOM path was skipped: no bom-docker-*.xml produced.
	code, lsOutput := ExecCommandExpectNonZero(t, container, "/chart", []string{"sh", "-c", "ls bom-docker-*.xml"})
	assert.NotEqual(0, code, "no bom-docker-*.xml should be produced for an image-free chart; found:\n%s", lsOutput)

	// buildSettingsInfo (SLC-29) is persisted and records the build flags.
	buildSettings := ReadFile(t, container, "/chart/.pipeline/commonPipelineEnvironment/custom/buildSettingsInfo")
	assert.NotEmpty(strings.TrimSpace(string(buildSettings)), "buildSettingsInfo must not be empty")
	assert.Contains(string(buildSettings), "\"createBOM\":true", "buildSettingsInfo should record createBOM=true")
	assert.Contains(string(buildSettings), "\"publish\":true", "buildSettingsInfo should record publish=true")
}

// TestHelmIntegrationPublishWithSigning exercises the full chart-signing flow
// end-to-end (Tier 1): a passphrase-free GPG key is written to a temp keyring
// inside the container, helmExecute invokes helm package --sign --key --keyring,
// and both the .tgz and the .prov provenance file are uploaded to the WebDAV
// sink. A regression in the vault-secret-file → keyring-path wiring would fail
// the helm package call or omit the second upload, both of which are caught here.
func TestHelmIntegrationPublishWithSigning(t *testing.T) {
	t.Parallel()
	assert := NewContainerAssert(t)
	ctx := context.Background()

	networkName := "helm-signing-" + uuid.New().String()
	provider, err := testcontainers.ProviderDocker.GetProvider()
	if err != nil {
		t.Fatalf("Failed to get docker provider: %v", err)
	}
	network, err := provider.CreateNetwork(ctx, testcontainers.NetworkRequest{Name: networkName, CheckDuplicate: true})
	if err != nil {
		t.Fatalf("Failed to create network: %v", err)
	}
	defer network.Remove(ctx)

	chartRepo, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:          "bytemark/webdav",
			Env:            map[string]string{"USERNAME": "repouser", "PASSWORD": "repopass"},
			Networks:       []string{networkName},
			NetworkAliases: map[string][]string{networkName: {"chartrepo"}},
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("Failed to start chart repo container: %v", err)
	}
	defer chartRepo.Terminate(ctx)

	// The signing fixture carries a piper config with signingKey/signingKeyRing
	// pointing to /chart/signing-keyring.gpg, which is created below.
	container := StartPiperContainer(t, ContainerConfig{
		Image:          "dtzar/helm-kubectl",
		TestData:       "TestHelmIntegration/signing/chart",
		WorkDir:        "/chart",
		Networks:       []string{networkName},
		NetworkAliases: []string{"helm-runner"},
	})

	// Install gnupg (Alpine), generate a passphrase-free RSA key, and export the
	// secret keyring to the path that the piper config references as signingKeyRing.
	// helm package --sign reads the secret keyring directly via the openpgp library
	// (no interactive GPG prompt needed when there is no passphrase).
	ExecCommand(t, container, "/chart", []string{"sh", "-c",
		"apk add --no-cache gnupg && " +
			"printf '%s\\n' 'Key-Type: RSA' 'Key-Length: 2048' 'Subkey-Type: RSA' 'Subkey-Length: 2048' " +
			"'Name-Real: Test Helm Signer' 'Name-Email: test@example.com' 'Expire-Date: 0' " +
			"'%no-passphrase' '%commit' > /tmp/gpg-batch && " +
			"gpg --batch --gen-key /tmp/gpg-batch && " +
			"gpg --export-secret-keys > /chart/signing-keyring.gpg",
	})

	output := RunPiper(t, container, "/chart", "helmExecute")

	// Both the chart archive and its provenance file must be uploaded.
	assert.Contains(output, "publishing artifact:")
	assert.Contains(output, "publishing provenance file:")

	// helm package --sign writes both files to the working directory before upload.
	assert.FileExists(container,
		"/chart/helm-int-hello-0.1.0.tgz",
		"/chart/helm-int-hello-0.1.0.tgz.prov",
	)
}
