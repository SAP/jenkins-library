package cmd

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

type execRunnerMock struct {
	dirValue   string
	envValue   []string
	outWriter  io.Writer
	errWriter  io.Writer
	executable string
	parameters []string
}

func (er *execRunnerMock) Dir(d string) {
	er.dirValue = d
}

func (er *execRunnerMock) SetEnv(e []string) {
	er.envValue = e
}

func (er *execRunnerMock) Stdout(out io.Writer) {
	er.outWriter = out
}

func (er *execRunnerMock) Stderr(err io.Writer) {
	er.errWriter = err
}
func (er *execRunnerMock) RunExecutable(e string, p ...string) error {
	er.executable = e
	er.parameters = p
	return nil
}

func TestDeterminePullRequestMerge(t *testing.T) {
	config := fortifyExecuteScanOptions{CommitMessage: "Merge pull request #2462 from branch f-test", PullRequestMessageRegex: `(?m).*Merge pull request #(\d+) from.*`, PullRequestMessageRegexGroup: 1}

	t.Run("success", func(t *testing.T) {
		match := determinePullRequestMerge(config)
		assert.Equal(t, "2462", match, "Expected different result")
	})

	t.Run("no match", func(t *testing.T) {
		config.CommitMessage = "Some test commit"
		match := determinePullRequestMerge(config)
		assert.Equal(t, "", match, "Expected different result")
	})
}

func TestTranslateProject(t *testing.T) {
	execRunner := execRunnerMock{}

	t.Run("python", func(t *testing.T) {
		config := fortifyExecuteScanOptions{ScanType: "pip", Memory: "-Xmx4G", Translate: `[{"pythonPath":"./some/path","pythonIncludes":"./**/*","pythonExcludes":"./tests/**/*"}]`}
		translateProject(config, &execRunner, "/commit/7267658798797")
		assert.Equal(t, "sourceanalyzer", execRunner.executable, "Expected different executable")
		assert.Equal(t, []string{"-verbose", "-64", "-b", "/commit/7267658798797", "-Xmx4G", "-python-path", "./some/path", "-exclude", "./tests/**/*", "./**/*"}, execRunner.parameters, "Expected different executable")
	})

	t.Run("asp", func(t *testing.T) {
		config := fortifyExecuteScanOptions{ScanType: "windows", Memory: "-Xmx6G", Translate: `[{"aspnetcore":"true","dotNetCoreVersion":"3.5","exclude":"./tests/**/*","libDirs":"tmp/","src":"./**/*"}]`}
		translateProject(config, &execRunner, "/commit/7267658798797")
		assert.Equal(t, "sourceanalyzer", execRunner.executable, "Expected different executable")
		assert.Equal(t, []string{"-verbose", "-64", "-b", "/commit/7267658798797", "-Xmx6G", "-aspnetcore", "-dotnet-core-version", "3.5", "-exclude", "./tests/**/*", "-libdirs", "tmp/", "./**/*"}, execRunner.parameters, "Expected different executable")
	})

	t.Run("java", func(t *testing.T) {
		config := fortifyExecuteScanOptions{ScanType: "java", Memory: "-Xmx2G", Translate: `[{"classpath":"./classes/*.jar","extdirs":"tmp/","jdk":"1.8.0-21","source":"1.8","sourcepath":"src/ext/","src":"./**/*"}]`}
		translateProject(config, &execRunner, "/commit/7267658798797")
		assert.Equal(t, "sourceanalyzer", execRunner.executable, "Expected different executable")
		assert.Equal(t, []string{"-verbose", "-64", "-b", "/commit/7267658798797", "-Xmx2G", "-cp", "./classes/*.jar", "-extdirs", "tmp/", "-source", "1.8", "-jdk", "1.8.0-21", "-sourcepath", "src/ext/", "./**/*"}, execRunner.parameters, "Expected different executable")
	})
}

func TestScanProject(t *testing.T) {
	config := fortifyExecuteScanOptions{Memory: "-Xmx4G"}
	execRunner := execRunnerMock{}

	t.Run("normal", func(t *testing.T) {
		scanProject(config, &execRunner, "/commit/7267658798797", "label")
		assert.Equal(t, "sourceanalyzer", execRunner.executable, "Expected different executable")
		assert.Equal(t, []string{"-verbose", "-64", "-b", "/commit/7267658798797", "-scan", "-Xmx4G", "-build-label", "label", "-logfile", "target/fortify-scan.log", "-f", "target/result.fpr"}, execRunner.parameters, "Expected different executable")
	})

	t.Run("quick", func(t *testing.T) {
		config.QuickScan = true
		scanProject(config, &execRunner, "/commit/7267658798797", "")
		assert.Equal(t, "sourceanalyzer", execRunner.executable, "Expected different executable")
		assert.Equal(t, []string{"-verbose", "-64", "-b", "/commit/7267658798797", "-scan", "-Xmx4G", "-quick", "-logfile", "target/fortify-scan.log", "-f", "target/result.fpr"}, execRunner.parameters, "Expected different executable")
	})
}
