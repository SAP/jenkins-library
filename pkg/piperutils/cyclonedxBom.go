package piperutils

import (
	"encoding/xml"
	"io"
	"os"
	"path/filepath"

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

func GetPurl(filePath, bomFilename string) string {
	bomFilePath := filepath.Join(filepath.Dir(filePath), bomFilename)
	exists, err := FileExists(bomFilePath)
	if err != nil {
		log.Entry().Warnf("unable to check if bom file exists: %v", err)
		return ""
	}
	if !exists {
		log.Entry().Debugf("bom file doesn't exist and hence no pURL info: %v", bomFilePath)
		return ""
	}
	bom, err := GetBom(bomFilePath)
	if err != nil {
		log.Entry().Warnf("unable to get bom metadata: %v", err)
		return ""
	}
	return bom.Metadata.Component.Purl
}
