package nexus

import (
	"bytes"
	"encoding/xml"
)

// Represent a Repository maven-metadata.xml file
type MavenMetadata struct {
	XMLName      xml.Name   `xml:"metadata"`
	ModelVersion string     `xml:"modelVersion,attr"`
	GroupId      string     `xml:"groupId"`
	ArtifactId   string     `xml:"artifactId"`
	Version      string     `xml:"version"`
	Versioning   Versioning `xml:"versioning"`
}

// Represent a versioning
type Versioning struct {
	XMLName          xml.Name          `xml:"versioning"`
	Snapshot         Snapshot          `xml:"snapshot"`
	LastUpdated      string            `xml:"lastUpdated"`
	SnapshotVersions []SnapshotVersion `xml:"snapshotVersions>snapshotVersion"`
}

// Represent a shapshot
type Snapshot struct {
	XMLName     xml.Name `xml:"snapshot"`
	TimeStamp   string   `xml:"timestamp"`
	BuildNumber int      `xml:"buildNumber"`
}

// Represent a shapshotVersions
type SnapshotVersion struct {
	XMLName    xml.Name `xml:"snapshotVersion"`
	Classifier string   `xml:"classifier"`
	Extension  string   `xml:"extension"`
	Value      string   `xml:"value"`
	Updated    string   `xml:"updated"`
}

func xmlBufferToMavenMetadata(buffer []byte) (*MavenMetadata, error) {
	var metadata MavenMetadata
	err := xml.Unmarshal(buffer, &metadata)
	if err != nil {
		return nil, err
	}
	return &metadata, nil
}

func mavenMetadataToXMLBuffer(metadata *MavenMetadata) ([]byte, error) {
	buffer := new(bytes.Buffer)
	_, err := buffer.WriteString("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	var blob []byte
	if err == nil {
		blob, err = xml.MarshalIndent(metadata, "", "  ")
	}
	if err == nil {
		_, err = buffer.Write(blob)
	}
	if err == nil {
		_, err = buffer.WriteString("\n")
	}
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), err
}
