package piperutils

import (
	"encoding/xml"
	"fmt"
	"io"
	"os"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/SAP/jenkins-library/pkg/log"
)

// To serialize the cyclonedx BOM file
type Bom struct {
	Metadata Metadata `xml:"metadata"`
}

type Metadata struct {
	Component  BomComponent  `xml:"component"`
	Properties []BomProperty `xml:"properties>property"`
}

type BomProperty struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
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

func GetPurl(bomFilePath string) string {
	bom, err := GetBom(bomFilePath)
	if err != nil {
		log.Entry().Warnf("unable to get bom metadata: %v", err)
		return ""
	}
	return bom.Metadata.Component.Purl
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
} // UpdateOrInsertPurl updates or inserts the PURL into the parent component of an SBOM

func UpdatePurl(sbomPath string, newPurl string) error {
	// Open SBOM file
	file, err := os.Open(sbomPath)
	if err != nil {
		return fmt.Errorf("failed to open SBOM file: %w", err)
	}
	defer file.Close()

	// Decode the SBOM
	var bom cdx.BOM
	decoder := cdx.NewBOMDecoder(file, cdx.BOMFileFormatJSON)
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
