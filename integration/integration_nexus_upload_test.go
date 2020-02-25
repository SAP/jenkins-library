// +build integration
// can be execute with go test -tags=integration ./integration/...

package main

import (
	"context"
	"fmt"
	"net/http"
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
		Image:        "sonatype/nexus3:3.21.1",
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
	cmd.Dir("testdata/TestNexusIntegration/mta")

	piperOptions := []string{
		"nexusUpload",
		"--groupId=mygroup",
		"--artifactId=mymta",
		"--user=admin",
		"--password=admin123",
		"--repository=maven-releases",
		"--url=" + nexusIpAndPort,
	}

	err = cmd.RunExecutable(getPiperExecutable(), piperOptions...)
	assert.NoError(t, err, "Calling piper with arguments %v failed.", piperOptions)

	cmd = command.Command{}
	cmd.Dir("testdata/TestNexusIntegration/maven")

	piperOptions = []string{
		"nexusUpload",
		"--groupId=mygroup",
		"--artifactId=mymaven",
		"--user=admin",
		"--password=admin123",
		"--repository=maven-releases",
		"--url=" + nexusIpAndPort,
	}

	err = cmd.RunExecutable(getPiperExecutable(), piperOptions...)
	assert.NoError(t, err, "Calling piper with arguments %v failed.", piperOptions)


	//resp, err = http.Get(url + "/repository/maven-releases/mygroup/mymaven/1.0/mymaven-1.0.pom")
	////  'http://localhost:32859/repository/maven-releases/mygroup/mymaven/1.0/mymaven-1.0.pom'
	//assert.NoError(t, err, "Downloading artifact failed")
	//assert.Equal(t, http.StatusOK, resp.StatusCode)

	//resp, err = http.Get(url + "/repository/maven-releases/mygroup/myapp-jar/1.0/myapp-jar-1.0.jar")
	//assert.NoError(t, err, "Downloading artifact failed")
	//assert.Equal(t, resp.StatusCode, http.StatusOK)

	resp, err = http.Get(url + "/repository/maven-releases/mygroup/mymta/0.3.0/mymta-0.3.0.yaml")
	assert.NoError(t, err, "Downloading artifact failed")
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = http.Get(url + "/repository/maven-releases/mygroup/mymta/0.3.0/mymta-0.3.0.mtar")
	assert.NoError(t, err, "Downloading artifact failed")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
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
	cmd.Dir("testdata/TestNexusIntegration/")

	piperOptions := []string{
		"nexusUpload",
		`--artifacts=[{"artifactId":"myapp-pom","classifier":"","type":"pom","file":"pom.xml"},{"artifactId":"myapp-jar","classifier":"","type":"jar","file":"test.jar"},{"artifactId":"myapp-yaml","classifier":"","type":"yaml","file":"mta.yaml"},{"artifactId":"myapp-mtar","classifier":"","type":"mtar","file":"test.mtar"}]`,
		"--groupId=mygroup",
		"--user=admin",
		"--password=admin123",
		"--repository=releases",
		"--version=1.0",
		"--nexusVersion=nexus2",
		"--url=" + nexusIpAndPort + "/nexus/",
	}

	err = cmd.RunExecutable(getPiperExecutable(), piperOptions...)
	assert.NoError(t, err, "Calling piper with arguments %v failed.", piperOptions)

	resp, err := http.Get(url + "content/repositories/releases/mygroup/myapp-pom/1.0/myapp-pom-1.0.pom")
	assert.NoError(t, err, "Downloading artifact failed")
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	resp, err = http.Get(url + "content/repositories/releases/mygroup/myapp-jar/1.0/myapp-jar-1.0.jar")
	assert.NoError(t, err, "Downloading artifact failed")
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	resp, err = http.Get(url + "content/repositories/releases/mygroup/myapp-yaml/1.0/myapp-yaml-1.0.yaml")
	assert.NoError(t, err, "Downloading artifact failed")
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	resp, err = http.Get(url + "content/repositories/releases/mygroup/myapp-mtar/1.0/myapp-mtar-1.0.mtar")
	assert.NoError(t, err, "Downloading artifact failed")
	assert.Equal(t, resp.StatusCode, http.StatusOK)
}
