package piperutils

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/SAP/jenkins-library/pkg/log"
)

// CycloneDX 1.4 BOM structure
// Spec: https://cyclonedx.org/docs/1.4/xml/

// Bom represents the root BOM element
type Bom struct {
	Xmlns      string      `xml:"xmlns,attr"`
	Metadata   Metadata    `xml:"metadata"`
	Components []Component `xml:"components>component,omitempty"`
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

type BomComponent struct {
	Purl    string `xml:"purl"`
	Name    string `xml:"name"`
	Version string `xml:"version"`
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

// GetBomVersion extracts the CycloneDX schema version from the BOM
func GetBomVersion(bomFilePath string) (string, error) {
	bom, err := GetBom(bomFilePath)
	if err != nil {
		return "", err
	}

	if strings.Contains(bom.Xmlns, "/1.4") {
		return "1.4", nil
	}
	if strings.Contains(bom.Xmlns, "/1.5") {
		return "1.5", nil
	}
	if strings.Contains(bom.Xmlns, "/1.6") {
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

// UpdatePurl updates the PURL in the BOM metadata component
// This uses the official CycloneDX library for robust XML handling
func UpdatePurl(sbomPath string, newPurl string) error {
	// Open SBOM file
	file, err := os.Open(sbomPath)
	if err != nil {
		return fmt.Errorf("failed to open SBOM file: %w", err)
	}
	defer file.Close()

	// Decode the SBOM using official CycloneDX library
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
			log.Entry().Debugf("adding purl in BOM: %s", newPurl)
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

	// Encode back to SBOM format with pretty printing
	encoder := cdx.NewBOMEncoder(outFile, cdx.BOMFileFormatXML)
	encoder.SetPretty(true)
	if err := encoder.Encode(&bom); err != nil {
		return fmt.Errorf("failed to encode updated SBOM: %w", err)
	}

	log.Entry().Debugf("SBOM updated successfully for: %s", sbomPath)
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

func GetVersion(bomFilePath string) string {
	bom, err := GetBom(bomFilePath)
	if err != nil {
		log.Entry().Warnf("unable to get bom metadata version: %v", err)
		return ""
	}
	return bom.Metadata.Component.Version
}

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
