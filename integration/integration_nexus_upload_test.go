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
	cmd.Dir("testdata/TestNexusIntegration/")

	piperOptions := []string{
		"nexusUpload",
		`--artifacts=[{"artifactId":"myapp-pom","classifier":"myapp-1.0","type":"pom","file":"pom.xml"},{"artifactId":"myapp-jar","classifier":"myapp-1.0","type":"jar","file":"Test.jar"}]`,
		"--groupId=mygroup",
		"--user=admin",
		"--password=admin123",
		"--repository=maven-releases",
		"--version=1.0",
		"--url=" + nexusIpAndPort,
	}

	err = cmd.RunExecutable(getPiperExecutable(), piperOptions...)
	assert.NoError(t, err, "Calling piper with arguments %v failed.", piperOptions)

	resp, err = http.Get(url + "/repository/maven-releases/mygroup/myapp-pom/1.0/myapp-pom-1.0.pom")
	assert.NoError(t, err, "Downloading artifact failed")
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	resp, err = http.Get(url + "/repository/maven-releases/mygroup/myapp-jar/1.0/myapp-jar-1.0.jar")
	assert.NoError(t, err, "Downloading artifact failed")
	assert.Equal(t, resp.StatusCode, http.StatusOK)
}
