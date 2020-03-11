package nexus

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var snapshotMetadataXML = `<?xml version="1.0" encoding="UTF-8"?>
<metadata modelVersion="1.1.0">
  <groupId>com.sap.opensap</groupId>
  <artifactId>employee-browser</artifactId>
  <version>1.3.0-SNAPSHOT</version>
  <versioning>
    <snapshot>
      <timestamp>20200311.082246</timestamp>
      <buildNumber>18</buildNumber>
    </snapshot>
    <lastUpdated>20200311082246</lastUpdated>
    <snapshotVersions>
      <snapshotVersion>
        <extension>pom</extension>
        <value>1.3.0-20200310.091352-17</value>
        <updated>20200310091352</updated>
      </snapshotVersion>
      <snapshotVersion>
        <extension>pom</extension>
        <value>1.3.0-20200311.082246-18</value>
        <updated>20200311082246</updated>
      </snapshotVersion>
    </snapshotVersions>
  </versioning>
</metadata>
`

var artifactMetadataXML = `<?xml version="1.0" encoding="UTF-8"?>
<metadata>
  <groupId>com.sap.opensap</groupId>
  <artifactId>employee-browser-application</artifactId>
  <versioning>
    <versions>
      <version>1.3.0-SNAPSHOT</version>
    </versions>
    <lastUpdated>20200310203442</lastUpdated>
  </versioning>
</metadata>
`

func TestConvertMavenMetadata(t *testing.T) {
	buffer := []byte(snapshotMetadataXML)
	metadata, err := xmlBufferToMavenMetadata(buffer)
	assert.NoError(t, err)

	assert.Equal(t, 18, metadata.Versioning.Snapshot.BuildNumber)

	buffer, err = mavenMetadataToXMLBuffer(metadata)
	assert.NoError(t, err)

	newString := string(buffer[:])
	assert.Equal(t, snapshotMetadataXML, newString, "expected conversion to be loss-less")
}

func TestConvertMavenMetadata(t *testing.T) {
	buffer := []byte(snapshotMetadataXML)
	metadata, err := xmlBufferToMavenMetadata(buffer)
	assert.NoError(t, err)

	assert.Equal(t, 18, metadata.Versioning.Snapshot.BuildNumber)

	buffer, err = mavenMetadataToXMLBuffer(metadata)
	assert.NoError(t, err)

	newString := string(buffer[:])
	assert.Equal(t, snapshotMetadataXML, newString, "expected conversion to be loss-less")
}
