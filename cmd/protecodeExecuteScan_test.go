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

	"github.com/SAP/jenkins-library/pkg/protecode"
	"github.com/stretchr/testify/assert"
)

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
		config := protecodeExecuteScanOptions{ReuseExisting: c.reuse, CleanupMode: c.clean, ProtecodeGroup: c.group, FetchURL: c.fetchURL, FilePath: testFile.Name()}

		got, _ := uploadScanOrDeclareFetch(config, 0, pc, testFile.Name())

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
		config := protecodeExecuteScanOptions{ReuseExisting: c.reuse, CleanupMode: c.clean, ProtecodeGroup: c.group, FetchURL: c.fetchURL, ProtecodeTimeoutMinutes: "3", ProtecodeExcludeCVEs: "CVE-2018-1, CVE-2017-1000382", ReportFileName: "./cache/report-file.txt"}

		got, productID, _ := executeProtecodeScan(pc, &config, "dummy", writeReportToFileMock)

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
		protocol    string
		want        string
	}{
		{"scanImage", "", "", "scanImage"},
		{"scanImage", "registryURL", "protocol", "remote://registryURL/scanImage"},
		{"containerScanImage", "containerRegistryUrl", "protocol", "remote://containerRegistryUrl/containerScanImage"},
		{"containerScanImage", "registryURL", "protocol", "remote://registryURL/containerScanImage"},
	}

	for _, c := range cases {
		config := protecodeExecuteScanOptions{ScanImage: c.scanImage, DockerRegistryURL: c.registryURL}

		got, _ := getURLAndFileNameFromDockerImage(&config)

		assert.Equal(t, c.want, got)
	}

}
