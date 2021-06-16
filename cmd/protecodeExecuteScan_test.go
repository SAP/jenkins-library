package cmd

import (
	"testing"

	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	pkgutil "github.com/GoogleContainerTools/container-diff/pkg/util"
	"github.com/SAP/jenkins-library/pkg/protecode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type DockerClientMock struct {
	imageName     string
	registryURL   string
	localPath     string
	includeLayers bool
}

//Download interface for download an image to a local path
type Download interface {
	GetImageSource() (string, error)
	DownloadImageToPath(imageSource, filePath string) (pkgutil.Image, error)
	TarImage(writer io.Writer, image pkgutil.Image) error
}

const (
	daemonPrefix = "daemon://"
	remotePrefix = "remote://"
)

func (c *DockerClientMock) GetImageSource() (string, error) {

	imageSource := c.imageName

	if len(c.registryURL) > 0 && len(c.localPath) <= 0 {
		registry := c.registryURL

		url, _ := url.Parse(c.registryURL)
		//remove protocoll from registryURL to get registry
		if len(url.Scheme) > 0 {
			registry = strings.Replace(c.registryURL, fmt.Sprintf("%v://", url.Scheme), "", 1)
		}

		if strings.HasSuffix(registry, "/") {
			imageSource = fmt.Sprintf("%v%v%v", remotePrefix, registry, c.imageName)
		} else {
			imageSource = fmt.Sprintf("%v%v/%v", remotePrefix, registry, c.imageName)
		}
	} else if len(c.localPath) > 0 {
		imageSource = c.localPath
		if !pkgutil.IsTar(c.localPath) {
			imageSource = fmt.Sprintf("%v%v", daemonPrefix, c.localPath)
		}
	}

	if len(imageSource) <= 0 {
		return imageSource, fmt.Errorf("There is no image source for the parameters: (Name: %v, Registry: %v, local Path: %v)", c.imageName, c.registryURL, c.localPath)
	}

	return imageSource, nil
}

//DownloadImageToPath download the image to the specified path
func (c *DockerClientMock) DownloadImageToPath(imageSource, filePath string) (pkgutil.Image, error) {

	return pkgutil.Image{}, nil
}

//TarImage write a tar from the given image
func (c *DockerClientMock) TarImage(writer io.Writer, image pkgutil.Image) error {

	return nil
}

func TestRunProtecodeScan(t *testing.T) {
	requestURI := ""
	dir, err := ioutil.TempDir("", "t")
	require.NoError(t, err, "Failed to create temporary directory")
	// clean up tmp dir
	defer func() { _ = os.RemoveAll(dir) }()
	testFile, err := ioutil.TempFile(dir, "t.tar")
	require.NoError(t, err)
	fileName := filepath.Base(testFile.Name())
	path := strings.ReplaceAll(testFile.Name(), fileName, "")

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		requestURI = req.RequestURI
		var b bytes.Buffer

		if requestURI == "/api/product/4486/" || requestURI == "/api/product/4711/" {
			violations := filepath.Join("testdata/TestProtecode", "protecode_result_violations.json")
			byteContent, err := ioutil.ReadFile(violations)
			require.NoErrorf(t, err, "failed reading %v", violations)
			response := protecode.ResultData{Result: protecode.Result{ProductID: 4711, ReportURL: requestURI}}
			err = json.Unmarshal(byteContent, &response)

			json.NewEncoder(&b).Encode(response)

		} else if requestURI == "/api/fetch/" {
			violations := filepath.Join("testdata/TestProtecode", "protecode_result_violations.json")
			byteContent, err := ioutil.ReadFile(violations)
			require.NoErrorf(t, err, "failed reading %v", violations)
			response := protecode.ResultData{Result: protecode.Result{ProductID: 4486, ReportURL: requestURI}}
			err = json.Unmarshal(byteContent, &response)

			json.NewEncoder(&b).Encode(response)

		} else if requestURI == "/api/product/4486/pdf-report" {

		} else if requestURI == "/api/upload/t.tar" {
			response := protecode.ResultData{Result: protecode.Result{ProductID: 4486, ReportURL: requestURI}}

			var b bytes.Buffer
			json.NewEncoder(&b).Encode(&response)
			rw.Write([]byte(b.Bytes()))
		} else {
			response := protecode.Result{ProductID: 4486, ReportURL: requestURI}
			json.NewEncoder(&b).Encode(&response)
		}

		rw.Write([]byte(b.Bytes()))
	}))

	// Close the server when test finishes
	defer server.Close()

	po := protecode.Options{ServerURL: server.URL}
	pc := protecode.Protecode{}
	pc.SetOptions(po)

	dClient := &DockerClientMock{imageName: "t", registryURL: "", localPath: path, includeLayers: false}

	influx := protecodeExecuteScanInflux{}
	reportPath = dir
	cachePath = dir

	t.Run("With tar as scan image", func(t *testing.T) {
		config := protecodeExecuteScanOptions{ServerURL: server.URL, TimeoutMinutes: "1", VerifyOnly: false, CleanupMode: "none", Group: "13", FetchURL: "/api/fetch/", ExcludeCVEs: "CVE-2018-1, CVE-2017-1000382", ReportFileName: "./cache/report-file.txt"}
		err = runProtecodeScan(&config, &influx, dClient)
		assert.NoError(t, err)
	})

	t.Run("Without tar as scan image", func(t *testing.T) {
		config := protecodeExecuteScanOptions{ServerURL: server.URL, ScanImage: "t", TimeoutMinutes: "1", VerifyOnly: false, CleanupMode: "none", Group: "13", ExcludeCVEs: "CVE-2018-1, CVE-2017-1000382", ReportFileName: "./cache/report-file.txt"}
		err = runProtecodeScan(&config, &influx, dClient)
		assert.NoError(t, err)
	})

}

func TestHandleArtifactVersion(t *testing.T) {
	cases := []struct {
		version string
		want    string
	}{
		{"1.0.0-20200131085038+eeb7c1033339bfd404d21ec5e7dc05c80e9e985e", "1"},
		{"2.20.20-20200131085038+eeb7c1033339bfd404d21ec5e7dc05c80e9e985e", "2"},
		{"3.20.20-20200131085038+eeb7c1033339bfd404d21ec5e7dc05c80e9e985e", "3"},
		{"4.20.20-20200131085038", "4"},
		{"5.20.20-20200131085038+", "5"},
		{"6.00", "6.00"},
		{"7.20.20", "7.20.20"},
	}

	for _, c := range cases {
		got := handleArtifactVersion(c.version)
		assert.Equal(t, c.want, got)
	}
}

func TestCreateClient(t *testing.T) {
	cases := []struct {
		timeout string
	}{
		{""},
		{"1"},
	}

	for _, c := range cases {
		config := protecodeExecuteScanOptions{TimeoutMinutes: c.timeout}

		client := createClient(&config)
		assert.NotNil(t, client, "client should not be empty")
	}
}

func TestCreateDockerClient(t *testing.T) {
	cases := []struct {
		scanImage         string
		dockerRegistryURL string
		filePath          string
		includeLayers     bool
	}{
		{"test", "url", "path", false},
		{"", "", "", true},
	}

	for _, c := range cases {
		config := protecodeExecuteScanOptions{ScanImage: c.scanImage, DockerRegistryURL: c.dockerRegistryURL, FilePath: c.filePath, IncludeLayers: c.includeLayers}
		client := createDockerClient(&config)
		assert.NotNil(t, client, "client should not be empty")
	}
}

func TestUploadScanOrDeclareFetch(t *testing.T) {
	// init
	testFile, err := ioutil.TempFile("", "testFileUpload")
	require.NoError(t, err)
	defer os.RemoveAll(testFile.Name()) // clean up
	fileName := filepath.Base(testFile.Name())
	path := strings.ReplaceAll(testFile.Name(), fileName, "")

	requestURI := ""
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		requestURI = req.RequestURI

		if requestURI == "/api/fetch/" {
			response := protecode.ResultData{Result: protecode.Result{ProductID: 4711, ReportURL: requestURI}}
			var b bytes.Buffer
			json.NewEncoder(&b).Encode(&response)
			rw.Write([]byte(b.Bytes()))
		}
		if requestURI == fmt.Sprintf("/api/upload/%v", fileName) || requestURI == fmt.Sprintf("/api/upload/PR_4711_%v", fileName) {
			response := protecode.ResultData{Result: protecode.Result{ProductID: 4711, ReportURL: requestURI}}

			var b bytes.Buffer
			json.NewEncoder(&b).Encode(&response)
			rw.Write([]byte(b.Bytes()))
		}
	}))

	// Close the server when test finishes
	defer server.Close()

	po := protecode.Options{ServerURL: server.URL}
	pc := protecode.Protecode{}
	pc.SetOptions(po)

	cases := []struct {
		reuse    bool
		clean    string
		group    string
		fetchURL string
		filePath string
		prName   string
		want     int
	}{
		{false, "test", "group1", "/api/fetch/", "", "", 4711},
		{false, "test", "group1", "", path, "", 4711},
		{false, "test", "group1", "", path, "PR_4711", 4711},
	}

	for _, c := range cases {
		// test
		config := protecodeExecuteScanOptions{VerifyOnly: c.reuse, CleanupMode: c.clean, Group: c.group, FetchURL: c.fetchURL, FilePath: c.filePath}
		got := uploadScanOrDeclareFetch(config, pc, fileName)
		// assert
		assert.Equal(t, c.want, got)
	}
}

func writeReportToFileMock(resp io.ReadCloser, reportFileName string) error {
	return nil
}

func TestExecuteProtecodeScan(t *testing.T) {
	testDataFile := filepath.Join("testdata", "TestProtecode", "protecode_result_violations.json")
	violationsAbsPath, err := filepath.Abs(testDataFile)
	require.NoErrorf(t, err, "failed to obtain absolute path to test data with violations: %v", err)

	requestURI := ""
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		requestURI = req.RequestURI
		var b bytes.Buffer

		if requestURI == "/api/product/4711/" {
			byteContent, err := ioutil.ReadFile(violationsAbsPath)
			require.NoErrorf(t, err, "failed reading %v", violationsAbsPath)
			response := protecode.ResultData{}
			err = json.Unmarshal(byteContent, &response)

			json.NewEncoder(&b).Encode(response)

		} else if requestURI == "/api/product/4711/pdf-report" {

		} else {
			response := protecode.ResultData{Result: protecode.Result{ProductID: 4711, ReportURL: requestURI}}
			json.NewEncoder(&b).Encode(&response)
		}

		rw.Write([]byte(b.Bytes()))
	}))

	// Close the server when test finishes
	defer server.Close()

	po := protecode.Options{ServerURL: server.URL, Duration: time.Minute * 3}
	pc := protecode.Protecode{}
	pc.SetOptions(po)

	cases := []struct {
		reuse    bool
		clean    string
		group    string
		fetchURL string
		want     int
	}{
		{false, "binary", "group1", "/api/fetch/", 4711},
	}

	resetDir, err := os.Getwd()
	require.NoErrorf(t, err, "Failed to get current directory: %v", err)
	defer func() { _ = os.Chdir(resetDir) }()

	for _, c := range cases {
		// init
		dir, err := ioutil.TempDir("", "t")
		require.NoErrorf(t, err, "Failed to create temporary directory: %v", err)
		// clean up tmp dir
		defer func() { _ = os.RemoveAll(dir) }()
		// change into tmp dir and write test data
		err = os.Chdir(dir)
		require.NoErrorf(t, err, "Failed to change into temporary directory: %v", err)
		reportPath = dir
		config := protecodeExecuteScanOptions{VerifyOnly: c.reuse, CleanupMode: c.clean, Group: c.group, FetchURL: c.fetchURL, TimeoutMinutes: "3", ExcludeCVEs: "CVE-2018-1, CVE-2017-1000382", ReportFileName: "./cache/report-file.txt"}
		influxData := &protecodeExecuteScanInflux{}
		// test
		executeProtecodeScan(influxData, pc, &config, "dummy", writeReportToFileMock)
		// assert
		assert.Equal(t, 1125, influxData.protecode_data.fields.historical_vulnerabilities)
		assert.Equal(t, 0, influxData.protecode_data.fields.triaged_vulnerabilities)
		assert.Equal(t, 1, influxData.protecode_data.fields.excluded_vulnerabilities)
		assert.Equal(t, 142, influxData.protecode_data.fields.major_vulnerabilities)
		assert.Equal(t, 226, influxData.protecode_data.fields.vulnerabilities)
	}
}

func TestCorrectDockerConfigEnvVar(t *testing.T) {
	t.Run("with credentials", func(t *testing.T) {
		// init
		testDirectory, _ := ioutil.TempDir(".", "")
		require.DirExists(t, testDirectory)
		defer os.RemoveAll(testDirectory)

		dockerConfigDir := filepath.Join(testDirectory, "myConfig")
		os.Mkdir(dockerConfigDir, 0755)
		require.DirExists(t, dockerConfigDir)

		dockerConfigFile := filepath.Join(dockerConfigDir, "docker.json")
		file, _ := os.Create(dockerConfigFile)
		defer file.Close()
		require.FileExists(t, dockerConfigFile)

		resetValue := os.Getenv("DOCKER_CONFIG")
		defer os.Setenv("DOCKER_CONFIG", resetValue)
		// test
		correctDockerConfigEnvVar(&protecodeExecuteScanOptions{DockerConfigJSON: dockerConfigFile})
		// assert
		absolutePath, _ := filepath.Abs(dockerConfigDir)
		assert.Equal(t, absolutePath, os.Getenv("DOCKER_CONFIG"))
	})
	t.Run("without credentials", func(t *testing.T) {
		// init
		resetValue := os.Getenv("DOCKER_CONFIG")
		defer os.Setenv("DOCKER_CONFIG", resetValue)
		// test
		correctDockerConfigEnvVar(&protecodeExecuteScanOptions{})
		// assert
		assert.Equal(t, resetValue, os.Getenv("DOCKER_CONFIG"))
	})
}

func TestGetTarName(t *testing.T) {
	cases := map[string]struct {
		image   string
		version string
		expect  string
	}{
		"with version suffix": {
			"com.sap.piper/sample-k8s-app-multistage:1.11-20200902040158_97a5cc34f1796ad735159f020dd55c0f3670a4cb",
			"1.11-20200902040158_97a5cc34f1796ad735159f020dd55c0f3670a4cb",
			"com.sap.piper_sample-k8s-app-multistage_1.tar",
		},
		"without version suffix": {
			"abc",
			"3.20.20-20200131085038+eeb7c1033339bfd404d21ec5e7dc05c80e9e985e",
			"abc_3.tar",
		},
		"without version": {
			"abc",
			"",
			"abc.tar",
		},
		"ScanImage without sha as artifactVersion": {
			"abc@sha256:12345",
			"",
			"abc.tar",
		},
		"ScanImage with sha as artifactVersion": {
			"ppiper/cf-cli@sha256:c25dbacb9ab6e912afe0fe926d8f9d949c60adfe55d16778bde5941e6c37be11",
			"c25dbacb9ab6e912afe0fe926d8f9d949c60adfe55d16778bde5941e6c37be11",
			"ppiper_cf-cli_c25dbacb9ab6e912afe0fe926d8f9d949c60adfe55d16778bde5941e6c37be11.tar",
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, c.expect, getTarName(&protecodeExecuteScanOptions{ScanImage: c.image, Version: c.version}))
		})
	}
}
