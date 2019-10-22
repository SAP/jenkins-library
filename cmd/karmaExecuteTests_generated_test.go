package cmd

import (
	"testing"
)

func TestKarmaExecuteTestsCommand(t *testing.T) {

	testCmd := KarmaExecuteTestsCommand()

	// only high level testing performed - details are tested in step generation procudure
	if testCmd.Use != "karmaExecuteTests" {
		t.Errorf("Expected command name to be 'karmaExecuteTests' but was '%v'", testCmd.Use)
	}

}
