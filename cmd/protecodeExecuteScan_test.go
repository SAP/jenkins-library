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
	if err != nil {
		t.Fatal("Failed to create temporary directory")
	}
	// clean up tmp dir
	defer os.RemoveAll(dir)
	testFile, err := ioutil.TempFile(dir, "t.tar")
	if err != nil {
		t.FailNow()
	}
	fileName := filepath.Base(testFile.Name())
	path := strings.ReplaceAll(testFile.Name(), fileName, "")

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		requestURI = req.RequestURI
		var b bytes.Buffer

		if requestURI == "/api/product/4486/" || requestURI == "/api/product/4711/" {
			violations := filepath.Join("testdata/TestProtecode", "protecode_result_violations.json")
			byteContent, err := ioutil.ReadFile(violations)
			if err != nil {
				t.Fatalf("failed reading %v", violations)
			}
			response := protecode.ResultData{Result: protecode.Result{ProductID: 4711, ReportURL: requestURI}}
			err = json.Unmarshal(byteContent, &response)

			json.NewEncoder(&b).Encode(response)

		} else if requestURI == "/api/fetch/" {
			violations := filepath.Join("testdata/TestProtecode", "protecode_result_violations.json")
			byteContent, err := ioutil.ReadFile(violations)
			if err != nil {
				t.Fatalf("failed reading %v", violations)
			}
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
		config := protecodeExecuteScanOptions{ServerURL: server.URL, TimeoutMinutes: "1", ReuseExisting: false, CleanupMode: "none", Group: "13", FetchURL: "/api/fetch/", ExcludeCVEs: "CVE-2018-1, CVE-2017-1000382", ReportFileName: "./cache/report-file.txt"}
		err = runProtecodeScan(&config, &influx, dClient)
		assert.Nil(t, err, "There should be no Error")
	})

	t.Run("Without tar as scan image", func(t *testing.T) {
		config := protecodeExecuteScanOptions{ServerURL: server.URL, ScanImage: "t", FilePath: path, TimeoutMinutes: "1", ReuseExisting: false, CleanupMode: "none", Group: "13", ExcludeCVEs: "CVE-2018-1, CVE-2017-1000382", ReportFileName: "./cache/report-file.txt"}
		err = runProtecodeScan(&config, &influx, dClient)
		assert.Nil(t, err, "There should be no Error")
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

var fileContent string

func writeToFileMock(f string, d []byte, p os.FileMode) error {
	fileContent = string(d)
	return nil
}

func TestWriteReportDataToJSONFile(t *testing.T) {

	expected := "{\"target\":\"REPORTFILENAME\",\"mandatory\":true,\"productID\":\"4711\",\"serverUrl\":\"DUMMYURL\",\"count\":\"0\",\"cvss2GreaterOrEqualSeven\":\"4\",\"cvss3GreaterOrEqualSeven\":\"3\",\"excludedVulnerabilities\":\"2\",\"triagedVulnerabilities\":\"0\",\"historicalVulnerabilities\":\"1\",\"Vulnerabilities\":[{\"cve\":\"Vulnerability\",\"cvss\":2.5,\"cvss3_score\":\"5.5\"}]}"

	var parsedResult map[string]int = make(map[string]int)
	parsedResult["historical_vulnerabilities"] = 1
	parsedResult["excluded_vulnerabilities"] = 2
	parsedResult["cvss3GreaterOrEqualSeven"] = 3
	parsedResult["cvss2GreaterOrEqualSeven"] = 4
	parsedResult["vulnerabilities"] = 5

	config := protecodeExecuteScanOptions{ServerURL: "DUMMYURL", ReportFileName: "REPORTFILENAME"}

	writeReportDataToJSONFile(&config, parsedResult, 4711, []protecode.Vuln{{"Vulnerability", 2.5, "5.5"}}, writeToFileMock)
	assert.Equal(t, fileContent, expected, "content should be not empty")
}

func TestUploadScanOrDeclareFetch(t *testing.T) {

	testFile, err := ioutil.TempFile("", "testFileUpload")
	if err != nil {
		t.FailNow()
	}
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

		config := protecodeExecuteScanOptions{ReuseExisting: c.reuse, CleanupMode: c.clean, Group: c.group, FetchURL: c.fetchURL, FilePath: c.filePath}
		got := uploadScanOrDeclareFetch(config, 0, pc, fileName)

		assert.Equal(t, c.want, got)
	}
}

func writeReportToFileMock(resp io.ReadCloser, reportFileName string) error {
	return nil
}

func TestExecuteProtecodeScan(t *testing.T) {
	requestURI := ""
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		requestURI = req.RequestURI
		var b bytes.Buffer

		if requestURI == "/api/product/4711/" {
			violations := filepath.Join("testdata/TestProtecode", "protecode_result_violations.json")
			byteContent, err := ioutil.ReadFile(violations)
			if err != nil {
				t.Fatalf("failed reading %v", violations)
			}
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

	for _, c := range cases {

		dir, err := ioutil.TempDir("", "t")
		if err != nil {
			t.Fatal("Failed to create temporary directory")
		}
		// clean up tmp dir
		defer os.RemoveAll(dir)
		reportPath = dir
		config := protecodeExecuteScanOptions{ReuseExisting: c.reuse, CleanupMode: c.clean, Group: c.group, FetchURL: c.fetchURL, TimeoutMinutes: "3", ExcludeCVEs: "CVE-2018-1, CVE-2017-1000382", ReportFileName: "./cache/report-file.txt"}

		got := executeProtecodeScan(pc, &config, "dummy", writeReportToFileMock)

		assert.Equal(t, 1125, got["historical_vulnerabilities"])
		assert.Equal(t, 0, got["triaged_vulnerabilities"])
		assert.Equal(t, 1, got["excluded_vulnerabilities"])
		assert.Equal(t, 129, got["cvss3GreaterOrEqualSeven"])
		assert.Equal(t, 13, got["cvss2GreaterOrEqualSeven"])
		assert.Equal(t, 226, got["vulnerabilities"])
	}
}
