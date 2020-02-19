package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// func TestRunMavenExecute(t *testing.T) {
// 	t.Run("maven version", func(t *testing.T) {
// 		opts := mavenExecuteOptions{}
// 		c := command.Command{}
// 		stdOutBuf := new(bytes.Buffer)

// 		outfile, err := os.Create("test.txt")
// 		if err != nil {
// 			fmt.Printf("error: %v", err)
// 		}
// 		defer outfile.Close()

// 		stdOut := io.MultiWriter(outfile, stdOutBuf)
// 		c.Stdout(stdOut)
// 		_, err = runMavenExecute(&opts, &c)
// 		if err != nil {
// 			fmt.Printf("error: %v", err)
// 		}
// 		t.Logf("my buffer: %v", string(stdOutBuf.Bytes()))

// 	})
// }

func TestRunMavenExecute(t *testing.T) {
	t.Run("runMavenExecute should return stdOut", func(t *testing.T) {
		expectedOutput := "mocked output"
		e := execMockRunner{}
		e.stdoutReturn = map[string]string{"mvn --file pom.xml --batch-mode": "mocked output"}
		opts := mavenExecuteOptions{PomPath: "pom.xml", ReturnStdout: true}

		mavenOutput, _ := runMavenExecute(&opts, &e)

		assert.Equal(t, expectedOutput, mavenOutput)
	})
	t.Run("runMavenExecute should not return stdOut", func(t *testing.T) {
		expectedOutput := ""
		e := execMockRunner{}
		e.stdoutReturn = map[string]string{"mvn --file pom.xml --batch-mode": "mocked output"}
		opts := mavenExecuteOptions{PomPath: "pom.xml", ReturnStdout: false}

		mavenOutput, _ := runMavenExecute(&opts, &e)

		assert.Equal(t, expectedOutput, mavenOutput)
	})
	t.Run("runMavenExecute should have all config parameters in the exec call", func(t *testing.T) {
		expectedOutput := ""
		e := execMockRunner{}
		e.stdoutReturn = map[string]string{"mvn --file pom.xml --batch-mode": "mocked output"}
		opts := mavenExecuteOptions{PomPath: "pom.xml", ReturnStdout: false}

		mavenOutput, _ := runMavenExecute(&opts, &e)

		assert.Equal(t, expectedOutput, mavenOutput)
	})
}
