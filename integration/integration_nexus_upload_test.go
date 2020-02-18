package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"
	"cmd"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestNexusUpload(t *testing.T) {
	// Nexus might take a long time to boot
	ctx, cancel := context.WithDeadline(context.Background(), time.Now().Add(5*time.Minute))
	defer cancel()

	req := testcontainers.ContainerRequest{
		Image:        "sonatype/nexus3:3.14.0", //FIXME in 3.14.0 nexus still has a hardcoded admin pw by default. In later versions the password is written to a file in a volueme -> harder to create the testcase
		ExposedPorts: []string{"8081/tcp"},
		WaitingFor:   wait.ForLog("Started Sonatype Nexus"),
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
	resp, err := http.Get(fmt.Sprintf("http://%s:%s", ip, port.Port()))
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d. Got %d.", http.StatusOK, resp.StatusCode)
	}

	nexusContainer.Exec(ctx, []string{"cat", "/nexus-data/admin.password"})
	cmd := exec.Command("go", "run", ".", "nexusUpload")
	err := cmd.Run()


	//todo the actual test
}
