//go:build integration
// +build integration

// can be executed with
// go test -v -tags integration -run TestNexusIntegration ./integration/...

package main

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func assertFileCanBeDownloaded(t *testing.T, container IntegrationTestDockerExecRunner, url string) {
	err := container.runScriptInsideContainer("curl -O " + url)
	if err != nil {
		t.Fatalf("Attempting to download file %s failed: %s", url, err)
	}
	container.assertHasFiles(t, "/project/"+path.Base(url))
}

func TestNexusIntegrationV3UploadMta(t *testing.T) {
	// t.Parallel()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:       "sonatype/nexus3:3.25.1",
		User:        "nexus",
		TestDir:     []string{"testdata", "TestNexusIntegration", "mta"},
		Environment: map[string]string{"NEXUS_SECURITY_RANDOMPASSWORD": "false"},
		Setup: []string{
			"/opt/sonatype/start-nexus-repository-manager.sh &",
			"curl https://ftp.fau.de/apache/maven/maven-3/3.6.3/binaries/apache-maven-3.6.3-bin.tar.gz | tar xz -C /tmp",
			"echo PATH=/tmp/apache-maven-3.6.3/bin:$PATH >> ~/.profile",
			"until curl --fail --silent http://localhost:8081/service/rest/v1/status; do sleep 5; done",
		},
	})
	defer container.terminate(t)

	err := container.whenRunningPiperCommand("nexusUpload", "--groupId=mygroup", "--artifactId=mymta",
		"--username=admin", "--password=admin123", "--mavenRepository=maven-releases", "--url=http://localhost:8081")
	if err != nil {
		t.Fatalf("Piper command failed %s", err)
	}

	container.assertHasOutput(t, "BUILD SUCCESS")
	assertFileCanBeDownloaded(t, container, "http://localhost:8081/repository/maven-releases/mygroup/mymta/0.3.0/mymta-0.3.0.mtar")
	assertFileCanBeDownloaded(t, container, "http://localhost:8081/repository/maven-releases/mygroup/mymta/0.3.0/mymta-0.3.0.yaml")
}

func TestNexusIntegrationV3UploadMaven(t *testing.T) {
	// t.Parallel()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:       "sonatype/nexus3:3.25.1",
		User:        "nexus",
		TestDir:     []string{"testdata", "TestNexusIntegration", "maven"},
		Environment: map[string]string{"NEXUS_SECURITY_RANDOMPASSWORD": "false"},
		Setup: []string{
			"/opt/sonatype/start-nexus-repository-manager.sh &",
			"curl https://ftp.fau.de/apache/maven/maven-3/3.6.3/binaries/apache-maven-3.6.3-bin.tar.gz | tar xz -C /tmp",
			"echo PATH=/tmp/apache-maven-3.6.3/bin:$PATH >> ~/.profile",
			"until curl --fail --silent http://localhost:8081/service/rest/v1/status; do sleep 5; done",
		},
	})
	defer container.terminate(t)

	err := container.whenRunningPiperCommand("nexusUpload", "--username=admin", "--password=admin123",
		"--mavenRepository=maven-releases", "--url=http://localhost:8081")
	if err != nil {
		t.Fatalf("Piper command failed %s", err)
	}

	container.assertHasOutput(t, "BUILD SUCCESS")
	assertFileCanBeDownloaded(t, container, "http://localhost:8081/repository/maven-releases/com/mycompany/app/my-app/1.0/my-app-1.0.pom")
	assertFileCanBeDownloaded(t, container, "http://localhost:8081/repository/maven-releases/com/mycompany/app/my-app/1.0/my-app-1.0.jar")
}

func TestNexusIntegrationV3UploadNpm(t *testing.T) {
	// t.Parallel()
	container := givenThisContainer(t, IntegrationTestDockerExecRunnerBundle{
		Image:       "sonatype/nexus3:3.25.1",
		User:        "nexus",
		TestDir:     []string{"testdata", "TestNexusIntegration", "npm"},
		Environment: map[string]string{"NEXUS_SECURITY_RANDOMPASSWORD": "false"},
		Setup: []string{
			"/opt/sonatype/start-nexus-repository-manager.sh &",
			"curl https://nodejs.org/dist/v12.18.3/node-v12.18.3-linux-x64.tar.gz | tar xz -C /tmp",
			"echo PATH=/tmp/node-v12.18.3-linux-x64/bin:$PATH >> ~/.profile",
			"until curl --fail --silent http://localhost:8081/service/rest/v1/status; do sleep 5; done",
			// Create npm repo because nexus does not bring one by default
			"curl -u admin:admin123 -d '{\"name\": \"npm-repo\", \"online\": true, \"storage\": {\"blobStoreName\": \"default\", \"strictContentTypeValidation\": true, \"writePolicy\": \"ALLOW_ONCE\"}}' --header \"Content-Type: application/json\" -X POST http://localhost:8081/service/rest/beta/repositories/npm/hosted",
		},
	})
	defer container.terminate(t)

	err := container.whenRunningPiperCommand("nexusUpload", "--username=admin", "--password=admin123",
		"--npmRepository=npm-repo", "--url=http://localhost:8081")
	if err != nil {
		t.Fatalf("Piper command failed %s", err)
	}

	container.assertHasOutput(t, "npm notice total files:   1")
	assertFileCanBeDownloaded(t, container, "http://localhost:8081/repository/npm-repo/npm-nexus-upload-test/-/npm-nexus-upload-test-1.0.0.tgz")
}

func TestNexusIntegrationV2Upload(t *testing.T) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "sonatype/nexus:2.14.18-01",
		ExposedPorts: []string{"8081/tcp"},
		WaitingFor:   wait.ForLog("org.sonatype.nexus.bootstrap.jetty.JettyServer - Running").WithStartupTimeout(5 * time.Minute), // Nexus takes more than one minute to boot
	}
	nexusContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	assert.NoError(t, err)
	defer nexusContainer.Terminate(ctx)
	ip, err := nexusContainer.Host(ctx)
	assert.NoError(t, err)
	port, err := nexusContainer.MappedPort(ctx, "8081")
	assert.NoError(t, err, "Could not map port for nexus container")
	nexusIpAndPort := fmt.Sprintf("%s:%s", ip, port.Port())
	url := "http://" + nexusIpAndPort + "/nexus/"

	cmd := command.Command{}
	cmd.SetDir("testdata/TestNexusIntegration/mta")

	piperOptions := []string{
		"nexusUpload",
		"--groupId=mygroup",
		"--artifactId=mymta",
		"--username=admin",
		"--password=admin123",
		"--mavenRepository=releases",
		"--version=nexus2",
		"--url=" + nexusIpAndPort + "/nexus/",
	}

	err = cmd.RunExecutable(getPiperExecutable(), piperOptions...)
	assert.NoError(t, err, "Calling piper with arguments %v failed.", piperOptions)

	cmd = command.Command{}
	cmd.SetDir("testdata/TestNexusIntegration/maven")

	piperOptions = []string{
		"nexusUpload",
		"--username=admin",
		"--password=admin123",
		"--mavenRepository=releases",
		"--version=nexus2",
		"--url=" + nexusIpAndPort + "/nexus/",
	}

	err = cmd.RunExecutable(getPiperExecutable(), piperOptions...)
	assert.NoError(t, err, "Calling piper with arguments %v failed.", piperOptions)

	cmd = command.Command{}
	cmd.SetDir("testdata/TestNexusIntegration/npm")

	piperOptions = []string{
		"nexusUpload",
		"--username=admin",
		"--password=admin123",
		"--npmRepository=npm-repo",
		"--version=nexus2",
		"--url=" + nexusIpAndPort + "/nexus/",
	}

	// Create npm repo for this test because nexus does not create one by default
	payload := strings.NewReader("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<repository><data><name>npm-repo</name><repoPolicy>RELEASE</repoPolicy><repoType>hosted</repoType><id>npm-repo</id><exposed>true</exposed><provider>npm-hosted</provider><providerRole>org.sonatype.nexus.proxy.repository.Repository</providerRole><format>npm</format></data></repository>")
	request, _ := http.NewRequest("POST", url+"service/local/repositories", payload)
	request.Header.Add("Content-Type", "application/xml")
	request.Header.Add("Authorization", "Basic YWRtaW46YWRtaW4xMjM=")
	response, err := http.DefaultClient.Do(request)
	assert.NoError(t, err)
	fmt.Println(response)
	assert.Equal(t, 201, response.StatusCode)

	err = cmd.RunExecutable(getPiperExecutable(), piperOptions...)
	assert.NoError(t, err, "Calling piper with arguments %v failed.", piperOptions)

	resp, err := http.Get(url + "content/repositories/releases/com/mycompany/app/my-app-parent/1.0/my-app-parent-1.0.pom")
	assert.NoError(t, err, "Downloading artifact failed")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Get my-app-parent-1.0.pom: %s", resp.Status)

	resp, err = http.Get(url + "content/repositories/releases/com/mycompany/app/my-app/1.0/my-app-1.0.pom")
	assert.NoError(t, err, "Downloading artifact failed")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Get my-app-1.0.pom: %s", resp.Status)

	resp, err = http.Get(url + "content/repositories/releases/com/mycompany/app/my-app/1.0/my-app-1.0.jar")
	assert.NoError(t, err, "Downloading artifact failed")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Get my-app-1.0.jar: %s", resp.Status)

	resp, err = http.Get(url + "content/repositories/releases/mygroup/mymta/0.3.0/mymta-0.3.0.yaml")
	assert.NoError(t, err, "Downloading artifact failed")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Get mymta-0.3.0.yaml: %s", resp.Status)

	resp, err = http.Get(url + "content/repositories/releases/mygroup/mymta/0.3.0/mymta-0.3.0.mtar")
	assert.NoError(t, err, "Downloading artifact failed")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Get mymta-0.3.0.mtar: %s", resp.Status)

	resp, err = http.Get(url + "content/repositories/npm-repo/npm-nexus-upload-test/-/npm-nexus-upload-test-1.0.0.tgz")
	assert.NoError(t, err, "Downloading artifact failed")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Get npm-nexus-upload-test-1.0.0.tgz: %s", resp.Status)
}
