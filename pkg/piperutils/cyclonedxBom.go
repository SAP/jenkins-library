package piperutils

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/SAP/jenkins-library/pkg/log"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/moby/buildkit/util/purl"
	purlParser "github.com/package-url/packageurl-go"
)

// CycloneDX 1.4 BOM structure
// Spec: https://cyclonedx.org/docs/1.4/xml/

// Bom represents the root BOM element
type Bom struct {
	Xmlns      string         `xml:"xmlns,attr"`
	Metadata   Metadata       `xml:"metadata"`
	Components []BomComponent `xml:"components>component,omitempty"`
}

// Metadata provides additional information about the BOM
type Metadata struct {
	Component BomComponent `xml:"component"`
}

// BomComponent represents the main component (application/project)
type BomComponent struct {
	Name    string `xml:"name"`
	Version string `xml:"version"`
	Purl    string `xml:"purl"`
}

func GetBom(absoluteBomPath string) (Bom, error) {
	xmlFile, err := os.Open(absoluteBomPath)
	if err != nil {
		log.Entry().Debugf("failed to open bom file %s", absoluteBomPath)
		return Bom{}, err
	}
	defer xmlFile.Close()
	byteValue, err := io.ReadAll(xmlFile)
	if err != nil {
		log.Entry().Debugf("failed to read bom file %s", absoluteBomPath)
		return Bom{}, err
	}
	var bom Bom
	err = xml.Unmarshal(byteValue, &bom)
	if err != nil {
		log.Entry().Debugf("failed to unmarshal bom file %s", absoluteBomPath)
		return Bom{}, err
	}
	return bom, nil
}

func GetComponent(bomFilePath string) BomComponent {
	bom, err := GetBom(bomFilePath)
	if err != nil {
		log.Entry().Warnf("unable to get bom metadata: %v", err)
		return BomComponent{}
	}
	return bom.Metadata.Component
}

// GetBomSchemaVersion extracts the CycloneDX schema version from the BOM
func GetBomSchemaVersion(bomFilePath string) (string, error) {
	bom, err := GetBom(bomFilePath)
	if err != nil {
		return "", err
	}
	return bomSchemaVersionFromXmlns(bom.Xmlns)
}

// GetBomSchemaVersionFromContent extracts the CycloneDX schema version from BOM content bytes
func GetBomSchemaVersionFromContent(bomContent []byte) (string, error) {
	var bom Bom
	if err := xml.Unmarshal(bomContent, &bom); err != nil {
		return "", fmt.Errorf("failed to parse BOM: %w", err)
	}
	return bomSchemaVersionFromXmlns(bom.Xmlns)
}

func bomSchemaVersionFromXmlns(xmlns string) (string, error) {
	if strings.Contains(xmlns, "/1.4") {
		return "1.4", nil
	}
	if strings.Contains(xmlns, "/1.5") {
		return "1.5", nil
	}
	if strings.Contains(xmlns, "/1.6") {
		return "1.6", nil
	}
	return "", fmt.Errorf("unable to determine CycloneDX version from BOM")
}

// ValidateBOM validates that the BOM conforms to CycloneDX 1.4 requirements
// with mandatory PURL as per project specifications
func ValidateBOM(bomContent []byte) error {
	var bom Bom
	if err := xml.Unmarshal(bomContent, &bom); err != nil {
		return fmt.Errorf("failed to parse BOM: %w", err)
	}

	// Validate xmlns is correct for CycloneDX
	if bom.Xmlns != "" && !strings.Contains(bom.Xmlns, "cyclonedx.org/schema/bom") {
		return fmt.Errorf("invalid xmlns: expected cyclonedx schema, got %s", bom.Xmlns)
	}

	// Validate that metadata component exists
	if bom.Metadata.Component.Name == "" {
		return fmt.Errorf("metadata.component.name is required but missing")
	}

	// MANDATORY: Validate that PURL is present in metadata component
	if err := ValidatePurl(bom.Metadata.Component.Purl); err != nil {
		return fmt.Errorf("metadata.component.purl validation failed: %w", err)
	}

	return nil
}

// ValidatePurl validates that a PURL is present and follows the Package URL spec
// PURL format: pkg:type/namespace/name@version
// Spec: https://github.com/package-url/purl-spec
func ValidatePurl(purl string) error {
	if purl == "" {
		return fmt.Errorf("purl is mandatory but was empty")
	}

	if !strings.HasPrefix(purl, "pkg:") {
		return fmt.Errorf("purl must start with 'pkg:' but got: %s", purl)
	}

	parts := strings.SplitN(purl, ":", 2)
	if len(parts) < 2 || parts[1] == "" {
		return fmt.Errorf("purl has invalid format: %s", purl)
	}

	return nil
}

func GetName(bomFilePath string) string {
	bom, err := GetBom(bomFilePath)
	if err != nil {
		log.Entry().Warnf("unable to get bom metadata name: %v", err)
		return ""
	}
	return bom.Metadata.Component.Name
}

// UpdatePurl updates the PURL in the BOM metadata component
// This uses the official CycloneDX library for robust XML handling
func UpdatePurl(sbomPath string, newPurl string) error {
	// Open SBOM file
	file, err := os.Open(sbomPath)
	if err != nil {
		return fmt.Errorf("failed to open SBOM file: %w", err)
	}
	defer file.Close()

	// Decode the SBOM
	var bom cdx.BOM
	decoder := cdx.NewBOMDecoder(file, cdx.BOMFileFormatXML)
	if err := decoder.Decode(&bom); err != nil {
		return fmt.Errorf("failed to decode SBOM: %w", err)
	}

	// Check and update Parent Component
	if bom.Metadata != nil && bom.Metadata.Component != nil {
		parent := bom.Metadata.Component

		if parent.PackageURL == "" {
			parent.PackageURL = newPurl
		} else {
			log.Entry().Debugf("purl already present in parent component hence not updating for: %s", sbomPath)
		}

	} else {
		return fmt.Errorf("no parent component found in SBOM metadata")
	}

	// Reopen the file for writing (truncate)
	outFile, err := os.Create(sbomPath)
	if err != nil {
		return fmt.Errorf("failed to open SBOM file for writing: %w", err)
	}
	defer outFile.Close()

	// Encode back to SBOM format
	encoder := cdx.NewBOMEncoder(outFile, cdx.BOMFileFormatXML)
	encoder.SetPretty(true)
	if err := encoder.Encode(&bom); err != nil {
		return fmt.Errorf("failed to encode updated SBOM: %w", err)
	}

	log.Entry().Debugf("SBOM updated successfully for: %s", sbomPath)
	return nil
}

// BuildRegistryFreeDockerPurl constructs a clean, registry-free docker PURL of
// the form pkg:docker/<name>@<version> for the given image name and version.
//
// Syft omits the PURL for the parent/root component of a container-image BOM
// (https://github.com/anchore/syft/issues/1408); callers use this to construct
// a replacement and inject it via UpdatePurl. The registry (host:port) is
// stripped so the emitted PURL is deterministic and does not leak the registry
// host. It also returns the parsed registry, name and version (from the
// intermediate PURL) so callers that need them — e.g. to build build-artifact
// coordinates — do not have to parse again.
func BuildRegistryFreeDockerPurl(name, version string) (registryFreePurl, registry, parsedName, parsedVersion string, err error) {
	constructedPurl, err := purl.RefToPURL("docker", fmt.Sprintf("%s:%s", name, version), nil)
	if err != nil {
		return "", "", "", "", fmt.Errorf("unable to create purl from reference: %w", err)
	}

	registry, parsedName, parsedVersion, err = ParsePurl(constructedPurl)
	// Defensive: constructedPurl was just produced by RefToPURL, so ParsePurl
	// cannot realistically fail on it — not reachable by input in a unit test.
	if err != nil {
		return "", "", "", "", fmt.Errorf("unable to parse purl: %w", err)
	}

	registryFreePurl, err = purl.RefToPURL("docker", fmt.Sprintf("%s:%s", parsedName, parsedVersion), nil)
	// Defensive: parsedName/parsedVersion come from an already-validated PURL, so
	// this second RefToPURL cannot realistically fail — not reachable by input.
	if err != nil {
		return "", "", "", "", fmt.Errorf("unable to create registry-free purl from reference: %w", err)
	}

	return registryFreePurl, registry, parsedName, parsedVersion, nil
}

// ParsePurl splits a docker PURL string into its registry, name and version.
// A PURL without a registry-like namespace defaults the registry to docker.io.
func ParsePurl(purlStr string) (registry, name, version string, err error) {
	p, err := purlParser.FromString(purlStr)
	if err != nil {
		return "", "", "", err
	}

	// Split namespace to extract registry
	// E.g., namespace = "ghcr.io/my-org"
	namespace := p.Namespace
	if namespace == "" {
		registry = "docker.io"
	} else {
		nsParts := strings.Split(namespace, "/")
		if strings.Contains(nsParts[0], ".") {
			registry = nsParts[0]
		} else {
			registry = "docker.io"
		}
	}

	name = p.Name
	version = p.Version
	return
}
