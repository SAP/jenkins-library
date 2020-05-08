package piperutils

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestGetMavenGAV(t *testing.T) {
	file, err := ioutil.TempFile("", "pom.xml")
	if err != nil {
		t.Fatal("Failed to create temporary workspace directory")
	}
	// clean up tmp dir
	defer os.RemoveAll(file.Name())

	data := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<project
	xmlns="http://maven.apache.org/POM/4.0.0"
	xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
	xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
	<modelVersion>4.0.0</modelVersion>
	<properties>
		<revision>1.2.3</revision>
	</properties>	<groupId>com.sap.cp.jenkins</groupId>
	<artifactId>jenkins-library</artifactId>
	<version>${revision}</version>
</project>
`)
	ioutil.WriteFile(file.Name(), data, 777)

	result, err := GetMavenCoordinates(file.Name())
	assert.NoError(t, err, "Didn't expert error but got one")
	assert.Equal(t, "com.sap.cp.jenkins", result.GroupID, "Expected different groupId value")
	assert.Equal(t, "jenkins-library", result.ArtifactID, "Expected different artifactId value")
	assert.Equal(t, "1.2.3", result.Version, "Expected different version value")
}

func TestFilter(t *testing.T) {
	text := `[INFO] Scanning for projects...
[INFO] 
[INFO] -----------------< com.sap.cp.jenkins:jenkins-library >-----------------
[INFO] Building SAP CP Piper Library 0-SNAPSHOT
[INFO] --------------------------------[ jar ]---------------------------------
[INFO]
[INFO] --- maven-help-plugin:3.2.0:evaluate (default-cli) @ jenkins-library ---
[INFO] No artifact parameter specified, using 'com.sap.cp.jenkins:jenkins-library:jar:0-SNAPSHOT' as project.
[INFO]
com.sap.cp.jenkins
[INFO] ------------------------------------------------------------------------
[INFO] BUILD SUCCESS
[INFO] ------------------------------------------------------------------------
[INFO] Total time:  4.912 s
[INFO] Finished at: 2020-02-26T12:08:06+01:00
[INFO] ------------------------------------------------------------------------`
	result := filter(text, `(?m)^[\s*\w+\.]+`)
	assert.Equal(t, "com.sap.cp.jenkins", result, "Expected different value")
}

func TestGetMavenGAVFromFile(t *testing.T) {

	t.Run("test success", func(t *testing.T) {
		descriptor, err := GetMavenCoordinates("./testdata/test_pom.xml")

		assert.Nil(t, err)
		assert.Equal(t, "test.groupID", descriptor.GroupID)
		assert.Equal(t, "test-articatID", descriptor.ArtifactID)
		assert.Equal(t, "1.0.0", descriptor.Version)
	})
}

func TestGetMavenGAVFromFile2(t *testing.T) {

	t.Run("test success", func(t *testing.T) {
		descriptor, err := GetMavenCoordinates("./testdata/test2_pom.xml")

		assert.Nil(t, err)
		assert.Equal(t, "com.sap.ldi", descriptor.GroupID)
		assert.Equal(t, "parent-inherit-test", descriptor.ArtifactID)
		assert.Equal(t, "1.0.0", descriptor.Version)
		assert.Equal(t, "jar", descriptor.Packaging)
	})
}

func TestGetMavenGAVVersionViaInterface(t *testing.T) {

	t.Run("test success", func(t *testing.T) {
		var descriptor BuildDescriptor
		descriptor, err := GetMavenCoordinates("./testdata/test2_pom.xml")

		assert.Nil(t, err)
		assert.Equal(t, "1.0.0", descriptor.GetVersion())
	})
}

func TestGetPipGAV(t *testing.T) {

	t.Run("test success", func(t *testing.T) {

		descriptor, err := GetPipCoordinates("./testdata/setup.py")

		assert.Nil(t, err)
		assert.Equal(t, "some-test", descriptor.ArtifactID)
		assert.Equal(t, "1.0.0-SNAPSHOT", descriptor.Version)
	})
}

func TestGetPipGAVWithVersionString(t *testing.T) {

	t.Run("test success", func(t *testing.T) {
		var descriptor BuildDescriptor
		descriptor, err := GetPipCoordinates("./testdata/2_setup.py")

		assert.Nil(t, err)
		assert.Equal(t, "1.0.0", descriptor.GetVersion())
	})
}

func TestGetPipGAVVersionViaInterface(t *testing.T) {

	t.Run("test success", func(t *testing.T) {
		descriptor, err := GetPipCoordinates("./testdata/2_setup.py")

		assert.Nil(t, err)
		assert.Equal(t, "some-test", descriptor.ArtifactID)
		assert.Equal(t, "1.0.0", descriptor.Version)
	})
}

func TestDetermineProjectCoordinates(t *testing.T) {
	nameTemplate := `{{list .GroupID .ArtifactID | join "-" | trimAll "-"}}`

	t.Run("maven", func(t *testing.T) {
		var gav BuildDescriptor
		gav = &MavenDescriptor{GroupID: "com.test.pkg", ArtifactID: "analyzer", Version: "1.2.3"}
		name, version := DetermineProjectCoordinates(nameTemplate, "major", gav)
		assert.Equal(t, "com.test.pkg-analyzer", name, "Expected different project name")
		assert.Equal(t, "1", version, "Expected different project version")
	})

	t.Run("maven major-minor", func(t *testing.T) {
		var gav BuildDescriptor
		gav = &MavenDescriptor{GroupID: "com.test.pkg", ArtifactID: "analyzer", Version: "1.2.3"}
		name, version := DetermineProjectCoordinates(nameTemplate, "major-minor", gav)
		assert.Equal(t, "com.test.pkg-analyzer", name, "Expected different project name")
		assert.Equal(t, "1.2", version, "Expected different project version")
	})

	t.Run("maven full", func(t *testing.T) {
		var gav BuildDescriptor
		gav = &MavenDescriptor{GroupID: "com.test.pkg", ArtifactID: "analyzer", Version: "1.2.3-7864387648746"}
		name, version := DetermineProjectCoordinates(nameTemplate, "full", gav)
		assert.Equal(t, "com.test.pkg-analyzer", name, "Expected different project name")
		assert.Equal(t, "1.2.3-7864387648746", version, "Expected different project version")
	})

	t.Run("maven semantic", func(t *testing.T) {
		var gav BuildDescriptor
		gav = &MavenDescriptor{GroupID: "com.test.pkg", ArtifactID: "analyzer", Version: "1.2.3-7864387648746"}
		name, version := DetermineProjectCoordinates(nameTemplate, "semantic", gav)
		assert.Equal(t, "com.test.pkg-analyzer", name, "Expected different project name")
		assert.Equal(t, "1.2.3", version, "Expected different project version")
	})

	t.Run("maven original", func(t *testing.T) {
		var gav BuildDescriptor
		gav = &MavenDescriptor{GroupID: "com.test.pkg", ArtifactID: "analyzer", Version: "0-SNAPSHOT"}
		name, version := DetermineProjectCoordinates(nameTemplate, "snapshot", gav)
		assert.Equal(t, "com.test.pkg-analyzer", name, "Expected different project name")
		assert.Equal(t, "0-SNAPSHOT", version, "Expected different project version")
	})

	t.Run("python", func(t *testing.T) {
		var gav BuildDescriptor
		gav = &PipDescriptor{GroupID: "", ArtifactID: "python-test", Version: "2.2.3"}
		name, version := DetermineProjectCoordinates(nameTemplate, "major", gav)
		assert.Equal(t, "python-test", name, "Expected different project name")
		assert.Equal(t, "2", version, "Expected different project version")
	})
}
