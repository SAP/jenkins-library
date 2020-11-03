package whitesource

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestExecuteScanMTA(t *testing.T) {
	const pomXML = `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <artifactId>my-artifact-id</artifactId>
    <packaging>jar</packaging>
</project>
`
	config := ScanOptions{
		ScanType:    "mta",
		OrgToken:    "org-token",
		UserToken:   "user-token",
		ProductName: "mock-product",
		ProjectName: "mock-project",
	}

	t.Parallel()
	t.Run("happy path MTA", func(t *testing.T) {
		// init
		utilsMock := NewScanUtilsMock()
		utilsMock.AddFile("pom.xml", []byte(pomXML))
		utilsMock.AddFile("package.json", []byte(`{"name":"my-module-name"}`))
		scan := newTestScan(&config)
		// test
		err := scan.ExecuteMTAScan(&config, utilsMock)
		// assert
		require.NoError(t, err)
		expectedCalls := []mock.ExecCall{
			{
				Exec: "mvn",
				Params: []string{
					"--file",
					"pom.xml",
					"-Dorg.whitesource.orgToken=org-token",
					"-Dorg.whitesource.product=mock-product",
					"-Dorg.whitesource.checkPolicies=true",
					"-Dorg.whitesource.failOnError=true",
					"-Dorg.whitesource.aggregateProjectName=mock-project",
					"-Dorg.whitesource.aggregateModules=true",
					"-Dorg.whitesource.userKey=user-token",
					"-Dorg.whitesource.productVersion=product-version",
					"-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn",
					"--batch-mode",
					"org.whitesource:whitesource-maven-plugin:19.5.1:update",
				},
			},
			{
				Exec: "npm",
				Params: []string{
					"ls",
				},
			},
			{
				Exec: "npx",
				Params: []string{
					"whitesource",
					"run",
				},
			},
		}
		assert.Equal(t, expectedCalls, utilsMock.Calls)
		assert.True(t, utilsMock.HasWrittenFile(whiteSourceConfig))
		assert.True(t, utilsMock.HasRemovedFile(whiteSourceConfig))
		assert.Equal(t, expectedCalls, utilsMock.Calls)
	})
	t.Run("MTA with only maven modules", func(t *testing.T) {
		// init
		utilsMock := NewScanUtilsMock()
		utilsMock.AddFile("pom.xml", []byte(pomXML))
		scan := newTestScan(&config)
		// test
		err := scan.ExecuteMTAScan(&config, utilsMock)
		// assert
		require.NoError(t, err)
		expectedCalls := []mock.ExecCall{
			{
				Exec: "mvn",
				Params: []string{
					"--file",
					"pom.xml",
					"-Dorg.whitesource.orgToken=org-token",
					"-Dorg.whitesource.product=mock-product",
					"-Dorg.whitesource.checkPolicies=true",
					"-Dorg.whitesource.failOnError=true",
					"-Dorg.whitesource.aggregateProjectName=mock-project",
					"-Dorg.whitesource.aggregateModules=true",
					"-Dorg.whitesource.userKey=user-token",
					"-Dorg.whitesource.productVersion=product-version",
					"-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn",
					"--batch-mode",
					"org.whitesource:whitesource-maven-plugin:19.5.1:update",
				},
			},
		}
		assert.Equal(t, expectedCalls, utilsMock.Calls)
		assert.False(t, utilsMock.HasWrittenFile(whiteSourceConfig))
		assert.Equal(t, expectedCalls, utilsMock.Calls)
	})
	t.Run("MTA with only NPM modules", func(t *testing.T) {
		// init
		utilsMock := NewScanUtilsMock()
		utilsMock.AddFile("package.json", []byte(`{"name":"my-module-name"}`))
		scan := newTestScan(&config)
		// test
		err := scan.ExecuteMTAScan(&config, utilsMock)
		// assert
		require.NoError(t, err)
		expectedCalls := []mock.ExecCall{
			{
				Exec: "npm",
				Params: []string{
					"ls",
				},
			},
			{
				Exec: "npx",
				Params: []string{
					"whitesource",
					"run",
				},
			},
		}
		assert.Equal(t, expectedCalls, utilsMock.Calls)
		assert.True(t, utilsMock.HasWrittenFile(whiteSourceConfig))
		assert.True(t, utilsMock.HasRemovedFile(whiteSourceConfig))
		assert.Equal(t, expectedCalls, utilsMock.Calls)
	})
	t.Run("MTA with neither Maven nor NPM modules results in error", func(t *testing.T) {
		// init
		utilsMock := NewScanUtilsMock()
		scan := newTestScan(&config)
		// test
		err := scan.ExecuteMTAScan(&config, utilsMock)
		// assert
		assert.EqualError(t, err, "neither Maven nor NPM modules found, no scan performed")
	})
}
