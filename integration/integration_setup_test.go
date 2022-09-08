package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	pwd = filepath.Dir(pwd)
	f, err := os.OpenFile(filepath.Join(pwd, "integration", "images.txt"), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatal(err)
	}
	raw, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal(err)
	}
	lines := strings.Split(string(raw), "\n")
	for _, line := range lines {
		if err := exec.Command("docker", "pull", line).Run(); err != nil {
			log.Fatal(err)
		}
	}

	if err != nil {
		log.Fatal(err)
	}
	code := m.Run()
	os.Exit(code)
}
