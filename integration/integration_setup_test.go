package main

import (
	"log"
	"os"
	"os/exec"
	"testing"
)

func TestMain(m *testing.M) {
	err := exec.Command("init.sh").Run()
	if err != nil {
		log.Fatal(err)
	}
	code := m.Run()
	os.Exit(code)
}
