//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestHelmIntegration ./integration/...

package main

import (
	"context"
	"strings"
	"testing"

	"uuid"

	"github.com/SAP/jenkins-library/pkg/piperutils"
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
