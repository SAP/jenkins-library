// +build integration
// can be execute with go test -tags=integration ./integration/...

package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestNexusUpload(t *testing.T) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "sonatype/nexus3:3.22.0",
		ExposedPorts: []string{"8081/tcp"},
		Env:          map[string]string{"NEXUS_SECURITY_RANDOMPASSWORD": "false"},
		WaitingFor:   wait.ForLog("Started Sonatype Nexus").WithStartupTimeout(5 * time.Minute), // Nexus takes more than one minute to boot
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
	url := "http://" + nexusIpAndPort
	resp, err := http.Get(url)
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	cmd := command.Command{}
	cmd.SetDir("testdata/TestNexusIntegration/mta")

	piperOptions := []string{
		"nexusUpload",
		"--groupId=mygroup",
		"--artifactId=mymta",
		"--user=admin",
		"--password=admin123",
		"--mavenRepository=maven-releases",
		"--url=" + nexusIpAndPort,
	}

	err = cmd.RunExecutable(getPiperExecutable(), piperOptions...)
	assert.NoError(t, err, "Calling piper with arguments %v failed.", piperOptions)

	cmd = command.Command{}
	cmd.SetDir("testdata/TestNexusIntegration/maven")

	piperOptions = []string{
		"nexusUpload",
		"--user=admin",
		"--password=admin123",
		"--mavenRepository=maven-releases",
		"--url=" + nexusIpAndPort,
	}

	err = cmd.RunExecutable(getPiperExecutable(), piperOptions...)
	assert.NoError(t, err, "Calling piper with arguments %v failed.", piperOptions)

	cmd = command.Command{}
	cmd.SetDir("testdata/TestNexusIntegration/npm")

	piperOptions = []string{
		"nexusUpload",
		"--user=admin",
		"--password=admin123",
		"--npmRepository=npm-repo",
		"--url=" + nexusIpAndPort,
	}

	// Create npm repo for this test because nexus does not create one by default
	request, err := http.NewRequest("POST", url+"/service/rest/beta/repositories/npm/hosted", strings.NewReader(`{"name": "npm-repo", "online": true, "storage": {"blobStoreName": "default", "strictContentTypeValidation": true, "writePolicy": "ALLOW_ONCE"}}`))
	assert.NoError(t, err)
	request.Header.Set("Content-Type", "application/json")
	request.SetBasicAuth("admin", "admin123")
	response, err := http.DefaultClient.Do(request)
	assert.NoError(t, err)
	assert.Equal(t, 201, response.StatusCode)

	err = cmd.RunExecutable(getPiperExecutable(), piperOptions...)
	assert.NoError(t, err, "Calling piper with arguments %v failed.", piperOptions)

	resp, err = http.Get(url + "/repository/maven-releases/com/mycompany/app/my-app-parent/1.0/my-app-parent-1.0.pom")
	assert.NoError(t, err, "Downloading artifact failed")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Get my-app-parent-1.0.pom: %s", resp.Status)

	resp, err = http.Get(url + "/repository/maven-releases/com/mycompany/app/my-app/1.0/my-app-1.0.pom")
	assert.NoError(t, err, "Downloading artifact failed")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Get my-app-1.0.pom: %s", resp.Status)

	resp, err = http.Get(url + "/repository/maven-releases/com/mycompany/app/my-app/1.0/my-app-1.0.jar")
	assert.NoError(t, err, "Downloading artifact failed")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Get my-app-1.0.jar: %s", resp.Status)

	resp, err = http.Get(url + "/repository/maven-releases/com/mycompany/app/my-app/1.0/my-app-1.0-classes.jar")
	assert.NoError(t, err, "Downloading artifact failed")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Get my-app-1.0-classes.jar: %s", resp.Status)

	resp, err = http.Get(url + "/repository/maven-releases/mygroup/mymta/0.3.0/mymta-0.3.0.yaml")
	assert.NoError(t, err, "Downloading artifact failed")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Get mymta-0.3.0.yaml: %s", resp.Status)

	resp, err = http.Get(url + "/repository/maven-releases/mygroup/mymta/0.3.0/mymta-0.3.0.mtar")
	assert.NoError(t, err, "Downloading artifact failed")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Get mymta-0.3.0.mtar: %s", resp.Status)

	resp, err = http.Get(url + "/repository/npm-repo/npm-nexus-upload-test/-/npm-nexus-upload-test-1.0.0.tgz")
	assert.NoError(t, err, "Downloading artifact failed")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Get npm-nexus-upload-test-1.0.0.tgz: %s", resp.Status)
}

func TestNexus2Upload(t *testing.T) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "sonatype/nexus:2.14.16-01",
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
		"--user=admin",
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
		"--user=admin",
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
		"--user=admin",
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
