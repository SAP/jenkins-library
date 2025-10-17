package piperutils

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"strings"

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
	Name string `xml:"name"`
	Purl string `xml:"purl"`
}

// Component represents a software/hardware component
type Component struct {
	Name string `xml:"name"`
	Purl string `xml:"purl"`
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

func GetPurl(bomFilePath string) string {
	bom, err := GetBom(bomFilePath)
	if err != nil {
		log.Entry().Warnf("unable to get bom metadata: %v", err)
		return ""
	}
	return bom.Metadata.Component.Purl
}

// ValidateCycloneDX14 validates that the BOM conforms to CycloneDX 1.4 requirements
// with mandatory PURL as per project specifications
func ValidateCycloneDX14(bomFilePath string) error {
	bom, err := GetBom(bomFilePath)
	if err != nil {
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

	// Validate all components have PURLs (optional components should have PURLs too)
	for i, component := range bom.Components {
		if err := ValidatePurl(component.Purl); err != nil {
			log.Entry().Warnf("component[%d] (%s) purl validation failed: %v", i, component.Name, err)
		}
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
