package cmd

import (
	"testing"

	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"time"

	pkgutil "github.com/GoogleContainerTools/container-diff/pkg/util"
	"github.com/SAP/jenkins-library/pkg/protecode"
	"github.com/stretchr/testify/assert"
)

type mockGetDocker struct {
}

func (p mockGetDocker) GetDockerImage(scanImage string, registryURL string, includeLayers bool, cacheImagePath string) pkgutil.Image {

	return pkgutil.Image{}
}

func (p mockGetDocker) writeReportToFile(resp io.ReadCloser, reportFileName string) error {

	return nil
}

func TestRunProtecodeScan(t *testing.T) {

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
			response := protecode.Result{ProductId: 4711, ReportUrl: requestURI}
			json.NewEncoder(&b).Encode(&response)
		}

		rw.Write([]byte(b.Bytes()))
	}))

	// Close the server when test finishes
	defer server.Close()
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

	po := protecode.ProtecodeOptions{ServerURL: server.URL}
	pc := protecode.Protecode{}
	pc.SetOptions(po)

	mockGetDocker := mockGetDocker{}
	getImage = mockGetDocker.GetDockerImage
	writeReportToFile = mockGetDocker.writeReportToFile

	config := protecodeExecuteScanOptions{ProtecodeServerURL: server.URL, ScanImage: "t.tar", FilePath: testFile.Name(), ProtecodeTimeoutMinutes: "1", ReuseExisting: false, CleanupMode: "none", ProtecodeGroup: "13", FetchURL: "/api/fetch/", ProtecodeExcludeCVEs: "CVE-2018-1, CVE-2017-1000382", ReportFileName: "./cache/report-file.txt", Verbose: true}
	influx := protecodeExecuteScanInflux{}

	err = runProtecodeScan(&config, &influx)
	assert.Nil(t, err, "client should not be empty")
}

func TestCreateClient(t *testing.T) {
	cases := []struct {
		timeout string
	}{
		{""},
		{"1"},
	}

	for _, c := range cases {
		config := protecodeExecuteScanOptions{ProtecodeTimeoutMinutes: c.timeout, Verbose: true}

		client := createClient(&config)
		assert.NotNil(t, client, "client should not be empty")
	}
}

var fileContent string

func writeToFileMock(f string, d []byte, p os.FileMode) error {
	fileContent = string(d)
	return nil
}

func TestWriteReportDataToJSONFile(t *testing.T) {

	expected := "{\"target\":\"REPORTFILENAME\",\"mandatory\":true,\"productID\":\"4711\",\"protecodeServerUrl\":\"DUMMYURL\",\"count\":\"0\",\"cvss2GreaterOrEqualSeven\":\"4\",\"cvss3GreaterOrEqualSeven\":\"3\",\"excludedVulnerabilities\":\"2\",\"triagedVulnerabilities\":\"0\",\"historicalVulnerabilities\":\"1\"}"

	var parsedResult map[string]int = make(map[string]int)
	parsedResult["historical_vulnerabilities"] = 1
	parsedResult["excluded_vulnerabilities"] = 2
	parsedResult["cvss3GreaterOrEqualSeven"] = 3
	parsedResult["cvss2GreaterOrEqualSeven"] = 4
	parsedResult["vulnerabilities"] = 5

	config := protecodeExecuteScanOptions{ProtecodeServerURL: "DUMMYURL", ReportFileName: "REPORTFILENAME", Verbose: true}

	writeReportDataToJSONFile(&config, parsedResult, 4711, writeToFileMock)
	assert.Equal(t, fileContent, expected, "content should be not empty")
}

func TestUploadScanOrDeclareFetch(t *testing.T) {
	requestURI := ""
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		requestURI = req.RequestURI

		if requestURI == "/api/fetch/" {
			response := protecode.Result{ProductId: 4711, ReportUrl: requestURI}

			var b bytes.Buffer
			json.NewEncoder(&b).Encode(&response)
			rw.Write([]byte(b.Bytes()))
		} else {
			response := protecode.ResultData{Result: protecode.Result{ProductId: 4711, ReportUrl: requestURI}}

			var b bytes.Buffer
			json.NewEncoder(&b).Encode(&response)
			rw.Write([]byte(b.Bytes()))
		}
	}))

	// Close the server when test finishes
	defer server.Close()

	po := protecode.ProtecodeOptions{ServerURL: server.URL}
	pc := protecode.Protecode{}
	pc.SetOptions(po)
	testFile, err := ioutil.TempFile("", "testFileUpload")
	if err != nil {
		t.FailNow()
	}
	defer os.RemoveAll(testFile.Name()) // clean up

	cases := []struct {
		reuse    bool
		clean    string
		group    string
		fetchURL string
		filePath string
		want     int
	}{
		{false, "test", "group1", "/api/fetch/", "", 4711},
		{false, "test", "group1", "", testFile.Name(), 4711},
	}

	for _, c := range cases {
		config := protecodeExecuteScanOptions{ReuseExisting: c.reuse, CleanupMode: c.clean, ProtecodeGroup: c.group, FetchURL: c.fetchURL, FilePath: c.filePath, Verbose: true}

		got := uploadScanOrDeclareFetch(config, 0, pc, testFile.Name())

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
			response := protecode.Result{ProductId: 4711, ReportUrl: requestURI}
			json.NewEncoder(&b).Encode(&response)
		}

		rw.Write([]byte(b.Bytes()))
	}))

	// Close the server when test finishes
	defer server.Close()

	po := protecode.ProtecodeOptions{ServerURL: server.URL, Duration: time.Minute * 3}
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
		config := protecodeExecuteScanOptions{ReuseExisting: c.reuse, CleanupMode: c.clean, ProtecodeGroup: c.group, FetchURL: c.fetchURL, ProtecodeTimeoutMinutes: "3", ProtecodeExcludeCVEs: "CVE-2018-1, CVE-2017-1000382", ReportFileName: "./cache/report-file.txt", Verbose: true}

		got, productID := executeProtecodeScan(pc, &config, "dummy", writeReportToFileMock)

		assert.Equal(t, 4711, productID)
		assert.Equal(t, 1125, got["historical_vulnerabilities"])
		assert.Equal(t, 0, got["triaged_vulnerabilities"])
		assert.Equal(t, 1, got["excluded_vulnerabilities"])
		assert.Equal(t, 129, got["cvss3GreaterOrEqualSeven"])
		assert.Equal(t, 13, got["cvss2GreaterOrEqualSeven"])
		assert.Equal(t, 226, got["vulnerabilities"])
	}
}

func TestGetURLAndFileNameFromDockerImage(t *testing.T) {

	cases := []struct {
		scanImage   string
		registryURL string
		want        string
	}{
		{"scanImage", "", "scanImage"},
		{"scanImage", "registryURL", "remote://registryURL/scanImage"},
		{"containerScanImage", "containerRegistryUrl", "remote://containerRegistryUrl/containerScanImage"},
		{"containerScanImage", "registryURL", "remote://registryURL/containerScanImage"},
	}

	for _, c := range cases {

		got := getURLAndFileNameFromDockerImage(c.scanImage, c.registryURL)

		assert.Equal(t, c.want, got)
	}

}
