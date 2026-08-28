package helm

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"path/filepath"
	"strings"
	"time"
	"uuid"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/SAP/jenkins-library/pkg/docker"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/versioning"
	"go.yaml.in/yaml/v3"
	"helm.sh/helm/v3/pkg/chart"
)

// cycloneDxSchemaVersion is the CycloneDX schema version the chart SBOM is
// emitted in. helmExecute standardizes on 1.4 to match the container-image
// SBOMs produced by Syft (see pkg/syft/syft.go cyclonedxFormatForSyft). It is
// set explicitly on the BOM rather than relying on the cyclonedx-go default so
// the version cannot silently change on a library upgrade.
const cycloneDxSchemaVersion = "1.4"

// GenerateChartSBOM builds a chart-level CycloneDX SBOM (bom-helm.xml) that
// describes the helm chart artifact itself: the chart as the root component
// (with a pkg:helm PURL), its sub-chart dependencies, and the referenced
// container images. Unlike the container-image SBOMs, this is hand-built with
// the CycloneDX library since Syft has no helm-chart support.
func GenerateChartSBOM(chartPath, outputPath string, images []string, files piperutils.FileUtils) error {
	meta, err := readChartMetadata(chartPath, files)
	if err != nil {
		return err
	}

	bom := newChartBOM(meta, chartPath, images, files)
	return writeBOM(bom, outputPath, files)
}

// newChartBOM assembles the chart-level CycloneDX BOM: the chart as the root
// component, its sub-chart dependencies (Chart.lock resolved versions winning
// over Chart.yaml ranges), and the referenced container images.
func newChartBOM(meta *versioning.Metadata, chartPath string, images []string, files piperutils.FileUtils) *cdx.BOM {
	chartRef := fmt.Sprintf("pkg:helm/%s@%s", meta.Name, meta.Version)

	bom := cdx.NewBOM()
	// Pin the CycloneDX schema version explicitly (1.4) rather than relying on
	// the cyclonedx-go default, and set the matching XML namespace.
	bom.SpecVersion = cycloneDxSchemaVersion
	bom.XMLNS = "http://cyclonedx.org/schema/bom/" + cycloneDxSchemaVersion
	// The SBOM gateway requires a BOM serialNumber, a metadata timestamp, and a
	// bom-ref on the root component. cdx.NewBOM sets none of these (Syft-produced
	// BOMs carry them, which is why only this hand-built BOM was rejected).
	bom.SerialNumber = "urn:uuid:" + uuid.NewV4().String()
	bom.Metadata = &cdx.Metadata{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Component: &cdx.Component{
			BOMRef:     chartRef,
			Type:       cdx.ComponentTypeApplication,
			Name:       meta.Name,
			Version:    meta.Version,
			PackageURL: chartRef,
		},
	}

	var components []cdx.Component

	// Each sub-chart dependency becomes a component with a clean pkg:helm PURL.
	// The dependency Repository is a URL and is intentionally omitted from the
	// PURL to keep it purl-spec-valid and deterministic.
	//
	// Chart.lock, when present, pins the resolved dependency versions; it wins
	// over the (possibly ranged) versions declared in Chart.yaml so the SBOM
	// reflects what is actually packaged.
	dependencies := meta.Dependencies
	if locked := readChartLock(chartPath, files); len(locked) > 0 {
		dependencies = locked
	}
	for _, dep := range dependencies {
		depPurl := fmt.Sprintf("pkg:helm/%s@%s", dep.Name, dep.Version)
		components = append(components, cdx.Component{
			BOMRef:     depPurl,
			Type:       cdx.ComponentTypeLibrary,
			Name:       dep.Name,
			Version:    dep.Version,
			PackageURL: depPurl,
		})
	}

	// Each referenced container image becomes a container component with a
	// pkg:oci PURL of the form pkg:oci/<name>@<tag-or-digest>.
	for _, image := range images {
		name, err := docker.ContainerImageNameFromImage(image)
		if err != nil {
			log.Entry().Warnf("helm SBOM: skipping unparseable image %q: %v", image, err)
			continue
		}
		// ContainerImageNameFromImage already validated the reference (it parses
		// via ContainerImageNameTagFromImage internally), so this cannot fail.
		nameTag, _ := docker.ContainerImageNameTagFromImage(image)
		// nameTag is "<name>:<tag>" or "<name>@<digest>"; take the part after
		// the name as the PURL version.
		version := strings.TrimLeft(strings.TrimPrefix(nameTag, name), ":@")
		imagePurl := fmt.Sprintf("pkg:oci/%s@%s", name, version)
		components = append(components, cdx.Component{
			BOMRef:     imagePurl,
			Type:       cdx.ComponentTypeContainer,
			Name:       name,
			Version:    version,
			PackageURL: imagePurl,
		})
	}

	if len(components) > 0 {
		bom.Components = &components
	}
	return bom
}

// writeBOM encodes the BOM as pretty CycloneDX XML, validates it against the
// CycloneDX 1.4 required-field rules, and writes it to outputPath. Validation
// runs on the marshaled bytes so a non-conforming BOM never reaches disk.
func writeBOM(bom *cdx.BOM, outputPath string, files piperutils.FileUtils) error {
	var buf bytes.Buffer
	encoder := cdx.NewBOMEncoder(&buf, cdx.BOMFileFormatXML)
	encoder.SetPretty(true)
	if err := encoder.Encode(bom); err != nil {
		return fmt.Errorf("failed to encode chart SBOM: %w", err)
	}

	if err := validateChartBOM(buf.Bytes()); err != nil {
		return fmt.Errorf("generated chart SBOM is not schema-conformant: %w", err)
	}

	return files.FileWrite(outputPath, buf.Bytes(), 0o644)
}

// validateChartBOM checks the marshaled BOM against the CycloneDX 1.4
// required-field rules enforced by the downstream SBOM gateway, failing fast at
// build time instead of at upload. This is not a full XSD validation (no
// pure-Go XSD validator exists and libxml2/cgo would break the CGO_ENABLED=0
// static build); it codifies the specific constraints the gateway requires:
// a urn:uuid serialNumber, a metadata timestamp, the 1.4 namespace, and a
// bom-ref + valid PURL on the root component and every listed component.
func validateChartBOM(content []byte) error {
	var bom struct {
		XMLNS        string `xml:"xmlns,attr"`
		SerialNumber string `xml:"serialNumber,attr"`
		Metadata     struct {
			Timestamp string `xml:"timestamp"`
			Component struct {
				BOMRef string `xml:"bom-ref,attr"`
				Name   string `xml:"name"`
				Purl   string `xml:"purl"`
			} `xml:"component"`
		} `xml:"metadata"`
		Components []struct {
			BOMRef string `xml:"bom-ref,attr"`
			Name   string `xml:"name"`
			Purl   string `xml:"purl"`
		} `xml:"components>component"`
	}
	if err := xml.Unmarshal(content, &bom); err != nil {
		return fmt.Errorf("failed to parse generated BOM: %w", err)
	}

	if want := "http://cyclonedx.org/schema/bom/" + cycloneDxSchemaVersion; bom.XMLNS != want {
		return fmt.Errorf("unexpected xmlns %q, want %q", bom.XMLNS, want)
	}
	if !strings.HasPrefix(bom.SerialNumber, "urn:uuid:") {
		return fmt.Errorf("serialNumber %q is missing or not a urn:uuid", bom.SerialNumber)
	}
	if bom.Metadata.Timestamp == "" {
		return fmt.Errorf("metadata timestamp is required")
	}
	root := bom.Metadata.Component
	if root.BOMRef == "" || root.Name == "" {
		return fmt.Errorf("root component must have a bom-ref and a name")
	}
	if err := piperutils.ValidatePurl(root.Purl); err != nil {
		return fmt.Errorf("root component purl: %w", err)
	}
	for i, c := range bom.Components {
		if c.BOMRef == "" || c.Name == "" {
			return fmt.Errorf("component %d (%q) must have a bom-ref and a name", i, c.Name)
		}
		if err := piperutils.ValidatePurl(c.Purl); err != nil {
			return fmt.Errorf("component %d (%q) purl: %w", i, c.Name, err)
		}
	}
	return nil
}

// readChartMetadata reads and parses <chartPath>/Chart.yaml into the shared
// versioning.Metadata model.
func readChartMetadata(chartPath string, files piperutils.FileUtils) (*versioning.Metadata, error) {
	chartYamlPath := filepath.Join(chartPath, "Chart.yaml")
	content, err := files.FileRead(chartYamlPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", chartYamlPath, err)
	}

	var meta versioning.Metadata
	if err := yaml.Unmarshal(content, &meta); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", chartYamlPath, err)
	}

	return &meta, nil
}

// chartLock models the subset of Chart.lock needed for the SBOM: the resolved
// dependency list. Chart.lock has the same dependency shape as Chart.yaml.
type chartLock struct {
	Dependencies []*chart.Dependency `yaml:"dependencies,omitempty"`
}

// readChartLock reads <chartPath>/Chart.lock and returns its resolved
// dependency list. It is best-effort: a missing or unparseable lock yields nil
// so the caller falls back to the Chart.yaml dependencies.
func readChartLock(chartPath string, files piperutils.FileUtils) []*chart.Dependency {
	lockPath := filepath.Join(chartPath, "Chart.lock")
	content, err := files.FileRead(lockPath)
	if err != nil {
		return nil
	}

	var lock chartLock
	if err := yaml.Unmarshal(content, &lock); err != nil {
		return nil
	}
	return lock.Dependencies
}
