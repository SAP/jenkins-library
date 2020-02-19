package main

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestNexusUpload(t *testing.T) {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "sonatype/nexus3:3.14.0", //FIXME in 3.14.0 nexus still has a hardcoded admin pw by default. In later versions the password is written to a file in a volueme -> harder to create the testcase
		ExposedPorts: []string{"8081/tcp"},
		WaitingFor:   wait.ForLog("Started Sonatype Nexus").WithStartupTimeout(time.Minute * 5), // Nexus takes more then one minute to boot
	}
	nexusContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Error(err)
	}
	defer nexusContainer.Terminate(ctx)
	ip, err := nexusContainer.Host(ctx)
	if err != nil {
		t.Error(err)
	}
	port, err := nexusContainer.MappedPort(ctx, "8081")
	if err != nil {
		t.Error(err)
	}
	url := fmt.Sprintf("http://%s:%s", ip, port.Port())
	resp, err := http.Get(url)
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}

	// cmd := exec.Command("go", "build", "..", "-o", "pipertestbin")
	// err = cmd.Run()
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// when
	params := []string{"nexusUpload", `--artifacts=[{"id":"blob","classifier":"blob-1.0","type":"pom","file":"testdata/TestNexusIntegration/pom.xml"}]`, "--groupId=foo", "--user=admin", "--password=admin123", "--repository=maven-releases", "--version=1.0", "--url=" + fmt.Sprintf("%s:%s", ip, port.Port())}
	c := command.Command{}
	err = c.RunExecutable("../pipertestbin", params...)
	if err != nil {
		t.Log(err)
	}

	// then
	resp, err = http.Get(fmt.Sprintf("http://%s:%s", ip, port.Port()) + "/repository/maven-releases/foo/blob/1.0/blob-1.0.pom")
	if resp.StatusCode != http.StatusOK {
		t.Log("Test failed")
		t.Log(err)
	}
	if resp != nil {
		t.Log(resp)
	}

}
