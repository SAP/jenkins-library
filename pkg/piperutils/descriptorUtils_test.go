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

	result, err := GetMavenGAV(file.Name())
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
