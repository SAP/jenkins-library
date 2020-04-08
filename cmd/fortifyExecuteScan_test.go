package cmd

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
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
		translateProject(config, &execRunner, "/commit/7267658798797", "")
		assert.Equal(t, "sourceanalyzer", execRunner.executable, "Expected different executable")
		assert.Equal(t, []string{"-verbose", "-64", "-b", "/commit/7267658798797", "-Xmx4G", "-python-path", "./some/path", "-exclude", "./tests/**/*", "./**/*"}, execRunner.parameters, "Expected different parameters")
	})

	t.Run("asp", func(t *testing.T) {
		config := fortifyExecuteScanOptions{ScanType: "windows", Memory: "-Xmx6G", Translate: `[{"aspnetcore":"true","dotNetCoreVersion":"3.5","exclude":"./tests/**/*","libDirs":"tmp/","src":"./**/*"}]`}
		translateProject(config, &execRunner, "/commit/7267658798797", "")
		assert.Equal(t, "sourceanalyzer", execRunner.executable, "Expected different executable")
		assert.Equal(t, []string{"-verbose", "-64", "-b", "/commit/7267658798797", "-Xmx6G", "-aspnetcore", "-dotnet-core-version", "3.5", "-exclude", "./tests/**/*", "-libdirs", "tmp/", "./**/*"}, execRunner.parameters, "Expected different parameters")
	})

	t.Run("java", func(t *testing.T) {
		config := fortifyExecuteScanOptions{ScanType: "maven", Memory: "-Xmx2G", Translate: `[{"classpath":"./classes/*.jar","extdirs":"tmp/","jdk":"1.8.0-21","source":"1.8","sourcepath":"src/ext/","src":"./**/*"}]`}
		translateProject(config, &execRunner, "/commit/7267658798797", "")
		assert.Equal(t, "sourceanalyzer", execRunner.executable, "Expected different executable")
		assert.Equal(t, []string{"-verbose", "-64", "-b", "/commit/7267658798797", "-Xmx2G", "-cp", "./classes/*.jar", "-extdirs", "tmp/", "-source", "1.8", "-jdk", "1.8.0-21", "-sourcepath", "src/ext/", "./**/*"}, execRunner.parameters, "Expected different parameters")
	})

	t.Run("auto classpath", func(t *testing.T) {
		config := fortifyExecuteScanOptions{ScanType: "maven", Memory: "-Xmx2G", Translate: `[{"classpath":"./classes/*.jar", "extdirs":"tmp/","jdk":"1.8.0-21","source":"1.8","sourcepath":"src/ext/","src":"./**/*"}]`}
		translateProject(config, &execRunner, "/commit/7267658798797", "./WEB-INF/lib/*.jar")
		assert.Equal(t, "sourceanalyzer", execRunner.executable, "Expected different executable")
		assert.Equal(t, []string{"-verbose", "-64", "-b", "/commit/7267658798797", "-Xmx2G", "-cp", "./WEB-INF/lib/*.jar", "-extdirs", "tmp/", "-source", "1.8", "-jdk", "1.8.0-21", "-sourcepath", "src/ext/", "./**/*"}, execRunner.parameters, "Expected different parameters")
	})
}

func TestScanProject(t *testing.T) {
	config := fortifyExecuteScanOptions{Memory: "-Xmx4G"}
	execRunner := execRunnerMock{}

	t.Run("normal", func(t *testing.T) {
		scanProject(config, &execRunner, "/commit/7267658798797", "label")
		assert.Equal(t, "sourceanalyzer", execRunner.executable, "Expected different executable")
		assert.Equal(t, []string{"-verbose", "-64", "-b", "/commit/7267658798797", "-scan", "-Xmx4G", "-build-label", "label", "-logfile", "target/fortify-scan.log", "-f", "target/result.fpr"}, execRunner.parameters, "Expected different parameters")
	})

	t.Run("quick", func(t *testing.T) {
		config.QuickScan = true
		scanProject(config, &execRunner, "/commit/7267658798797", "")
		assert.Equal(t, "sourceanalyzer", execRunner.executable, "Expected different executable")
		assert.Equal(t, []string{"-verbose", "-64", "-b", "/commit/7267658798797", "-scan", "-Xmx4G", "-quick", "-logfile", "target/fortify-scan.log", "-f", "target/result.fpr"}, execRunner.parameters, "Expected different parameters")
	})
}

func TestAutoresolveClasspath(t *testing.T) {
	config := fortifyExecuteScanOptions{AutodetectClasspath: false, AutodetectClasspathCommand: "{{.PythonVersion}} -c 'import sys;p=sys.path;p.remove('');print(';'.join(p))' > {{.File}}"}
	execRunner := execRunnerMock{}

	t.Run("turned off", func(t *testing.T) {
		result := autoresolveClasspath(config, &execRunner, nil)
		assert.Equal(t, "", result, "Expected different executable")
	})

	t.Run("turned on", func(t *testing.T) {
		dir, err := ioutil.TempDir("", "classpath")
		assert.NoError(t, err, "Unexpected error detected")
		defer os.RemoveAll(dir)
		file := filepath.Join(dir, "cp.txt")
		classpath := "/usr/lib/python35.zip;/usr/lib/python3.5;/usr/lib/python3.5/plat-x86_64-linux-gnu;/usr/lib/python3.5/lib-dynload;/home/piper/.local/lib/python3.5/site-packages;/usr/local/lib/python3.5/dist-packages;/usr/lib/python3/dist-packages;./lib"
		ioutil.WriteFile(file, []byte(classpath), 0700)

		config.AutodetectClasspath = true
		config.ScanType = "pip"
		context := map[string]string{"PythonVersion": "python2", "File": file}
		result := autoresolveClasspath(config, &execRunner, context)
		assert.Equal(t, "python2", execRunner.executable, "Expected different executable")
		assert.Equal(t, []string{"-c", "'import", "sys;p=sys.path;p.remove('');print(';'.join(p))'", ">", file}, execRunner.parameters, "Expected different parameters")
		assert.Equal(t, classpath, result, "Expected different result")
	})
}
