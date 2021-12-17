package gcs

import (
	"os"
	"testing"
)

func Test_gcsClient_prepareEnv(t *testing.T) {
	os.Setenv("TESTVAR1", "test1")

	gcsClient := gcsClient{envVars: []EnvVar{
		{Name: "TESTVAR1", Value: "test1_new"},
		{Name: "TESTVAR2", Value: "test2_new"},
	}}

	gcsClient.prepareEnv()

	if gcsClient.envVars[0].Modified {
		t.Errorf("%v - expected '%v' was '%v'", gcsClient.envVars[0].Name, false, gcsClient.envVars[0].Modified)
	}
	if !gcsClient.envVars[1].Modified {
		t.Errorf("%v - expected '%v' was '%v'", gcsClient.envVars[1].Name, true, gcsClient.envVars[1].Modified)
	}

	os.Setenv("TESTVAR1", "")
	os.Setenv("TESTVAR2", "")
}

func TestCleanupEnv(t *testing.T) {
	os.Setenv("TESTVAR1", "test1")
	os.Setenv("TESTVAR2", "test2")

	gcsClient := gcsClient{envVars: []EnvVar{
		{Name: "TESTVAR1", Modified: false},
		{Name: "TESTVAR2", Modified: true},
	}}

	gcsClient.cleanupEnv()

	if os.Getenv("TESTVAR1") != "test1" {
		t.Errorf("%v - expected '%v' was '%v'", gcsClient.envVars[0].Name, "test1", os.Getenv("TESTVAR1"))
	}
	if len(os.Getenv("TESTVAR2")) > 0 {
		t.Errorf("%v - expected '%v' was '%v'", gcsClient.envVars[1].Name, "", os.Getenv("TESTVAR2"))
	}

	os.Setenv("TESTVAR1", "")
	os.Setenv("TESTVAR2", "")
}
