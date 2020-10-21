package whitesource

import (
	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"testing"
)

func TestExecuteScanMaven(t *testing.T) {
	t.Parallel()
	t.Run("maven modules are aggregated", func(t *testing.T) {
		// init
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
			ScanType:    "maven",
			OrgToken:    "org-token",
			UserToken:   "user-token",
			ProductName: "mock-product",
			ProjectName: "mock-project",
		}
		utilsMock := NewScanUtilsMock()
		utilsMock.AddFile("pom.xml", []byte(pomXML))
		scan := newTestScan(&config)
		// test
		err := scan.ExecuteMavenScan(&config, utilsMock)
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
	})
	t.Run("maven modules are separate projects", func(t *testing.T) {
		// init
		const rootPomXML = `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <artifactId>my-artifact-id</artifactId>
    <packaging>jar</packaging>
	<modules>
		<module>sub</module>
	</modules>
</project>
`
		const modulePomXML = `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    <artifactId>my-artifact-id-sub</artifactId>
    <packaging>jar</packaging>
</project>
`
		config := ScanOptions{
			ScanType:    "maven",
			OrgToken:    "org-token",
			UserToken:   "user-token",
			ProductName: "mock-product",
		}
		utilsMock := NewScanUtilsMock()
		utilsMock.AddFile("pom.xml", []byte(rootPomXML))
		utilsMock.AddFile(filepath.Join("sub", "pom.xml"), []byte(modulePomXML))
		scan := newTestScan(&config)
		// test
		err := scan.ExecuteMavenScan(&config, utilsMock)
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
					"-Dorg.whitesource.userKey=user-token",
					"-Dorg.whitesource.productVersion=product-version",
					"-Dorg.slf4j.simpleLogger.log.org.apache.maven.cli.transfer.Slf4jMavenTransferListener=warn",
					"--batch-mode",
					"org.whitesource:whitesource-maven-plugin:19.5.1:update",
				},
			},
		}
		assert.Equal(t, expectedCalls, utilsMock.Calls)
		require.Len(t, scan.ScannedProjects(), 2)
		_, existsRoot := scan.ProjectByName("my-artifact-id - product-version")
		_, existsModule := scan.ProjectByName("my-artifact-id-sub - product-version")
		assert.True(t, existsRoot)
		assert.True(t, existsModule)
	})
	t.Run("pom.xml does not exist", func(t *testing.T) {
		// init
		config := ScanOptions{
			ScanType:    "maven",
			OrgToken:    "org-token",
			UserToken:   "user-token",
			ProductName: "mock-product",
		}
		utilsMock := NewScanUtilsMock()
		scan := newTestScan(&config)
		// test
		err := scan.ExecuteMavenScan(&config, utilsMock)
		// assert
		assert.EqualError(t, err,
			"for scanning with type 'maven', the file 'pom.xml' must exist in the project root")
		assert.Len(t, utilsMock.Calls, 0)
	})
}
