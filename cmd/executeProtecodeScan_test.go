package cmd

import (
	"testing"

	"bytes"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/protecode"
	"github.com/stretchr/testify/assert"
)

func TestLoadExistingProductSuccess(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		response := protecode.ProteCodeProductData{
			Products: []protecode.ProteCodeProduct{
				{ProductId: "test"}},
		}

		var b bytes.Buffer
		json.NewEncoder(&b).Encode(&response)
		rw.Write([]byte(b.Bytes()))
	}))
	// Close the server when test finishes
	defer server.Close()

	client := piperHttp.Client{}

	cases := []struct {
		protecodeServerURL string
		filePath           string
		protecodeGroup     string
		reuseExisting      bool
		want               string
	}{
		{server.URL, "filePath", "group", true, "test"},
		{server.URL, "filePath", "group32", false, ""},
	}
	for _, c := range cases {

		var config executeProtecodeScanOptions = executeProtecodeScanOptions{
			ProtecodeServerURL: c.protecodeServerURL,
			FilePath:           c.filePath,
			ProtecodeGroup:     c.protecodeGroup,
			ReuseExisting:      c.reuseExisting,
		}

		got := loadExistingProduct(config, client)
		assert.Equal(t, c.want, got)
	}
}
func TestLoadExistingProductByFilenameSuccess(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		response := protecode.ProteCodeProductData{
			Products: []protecode.ProteCodeProduct{
				{ProductId: "test"}},
		}

		var b bytes.Buffer
		json.NewEncoder(&b).Encode(&response)
		rw.Write([]byte(b.Bytes()))
	}))
	// Close the server when test finishes
	defer server.Close()

	client := piperHttp.Client{}

	cases := []struct {
		protecodeServerURL string
		filePath           string
		protecodeGroup     string
		want               *protecode.ProteCodeProductData
	}{
		{server.URL, "filePath", "group", &protecode.ProteCodeProductData{
			Products: []protecode.ProteCodeProduct{{ProductId: "test"}}}},
		{server.URL, "filePäth!", "group32", &protecode.ProteCodeProductData{
			Products: []protecode.ProteCodeProduct{{ProductId: "test"}}}},
	}
	for _, c := range cases {

		var config executeProtecodeScanOptions = executeProtecodeScanOptions{
			ProtecodeServerURL: c.protecodeServerURL,
			FilePath:           c.filePath,
			ProtecodeGroup:     c.protecodeGroup,
		}

		got := loadExistingProductByFilename(config, client)
		assert.Equal(t, c.want, got)
	}
}

func TestPullResultSuccess(t *testing.T) {

	requestURI := ""

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		requestURI = req.RequestURI

		var response protecode.ProteCodeResultData = protecode.ProteCodeResultData{}

		if strings.Contains(requestURI, "productID1") {
			response = protecode.ProteCodeResultData{
				Result: protecode.ProteCodeResult{ProductId: "productID1", ReportUrl: requestURI}}
		} else {
			response = protecode.ProteCodeResultData{
				Result: protecode.ProteCodeResult{ProductId: "productID2", ReportUrl: requestURI}}
		}

		var b bytes.Buffer
		json.NewEncoder(&b).Encode(&response)
		rw.Write([]byte(b.Bytes()))
	}))
	// Close the server when test finishes
	defer server.Close()

	client := piperHttp.Client{}

	cases := []struct {
		protecodeServerURL string
		productID          string
		want               protecode.ProteCodeResult
	}{
		{server.URL, "productID1", protecode.ProteCodeResult{ProductId: "productID1", ReportUrl: "/api/product/productID1/"}},
		{server.URL, "productID2", protecode.ProteCodeResult{ProductId: "productID2", ReportUrl: "/api/product/productID2/"}},
	}
	for _, c := range cases {

		var config executeProtecodeScanOptions = executeProtecodeScanOptions{
			ProtecodeServerURL: c.protecodeServerURL,
		}

		got := pullResult(config, c.productID, client)
		assert.Equal(t, c.want, got)
		assert.Equal(t, "/api/product/"+c.productID+"/", requestURI)
	}
}

func TestLoadReportSuccess(t *testing.T) {

	requestURI := ""
	var passedHeaders = map[string][]string{}
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		requestURI = req.RequestURI

		passedHeaders = map[string][]string{}
		if req.Header != nil {
			for name, headers := range req.Header {
				passedHeaders[name] = headers
			}
		}

		rw.Write([]byte("OK"))
	}))
	// Close the server when test finishes
	defer server.Close()

	client := piperHttp.Client{}

	cases := []struct {
		protecodeServerURL string
		productID          string
		reportFileName     string
		want               string
	}{
		{server.URL, "productID1", "fileName", "/api/product/productID1/pdf-report"},
		{server.URL, "productID2", "fileName", "/api/product/productID2/pdf-report"},
	}
	for _, c := range cases {

		var config executeProtecodeScanOptions = executeProtecodeScanOptions{
			ProtecodeServerURL: c.protecodeServerURL,
			Verbose:            false,
			ReportFileName:     c.reportFileName,
		}

		loadReport(config, c.productID, client)
		assert.Equal(t, requestURI, c.want)
		assert.Contains(t, passedHeaders, "Outputfile")
		assert.Contains(t, passedHeaders, "Pragma")
		assert.Contains(t, passedHeaders, "Cache-Control")
	}
}

func TestDeleteScanSuccess(t *testing.T) {

	requestURI := ""
	var passedHeaders = map[string][]string{}
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		requestURI = req.RequestURI

		passedHeaders = map[string][]string{}
		if req.Header != nil {
			for name, headers := range req.Header {
				passedHeaders[name] = headers
			}
		}

		rw.Write([]byte("OK"))
	}))
	// Close the server when test finishes
	defer server.Close()

	client := piperHttp.Client{}

	cases := []struct {
		cleanupMode        string
		protecodeServerURL string
		productID          string
		want               string
	}{
		{"binary", server.URL, "productID1", ""},
		{"complete", server.URL, "productID2", "/api/product/productID2/"},
	}
	for _, c := range cases {

		var config executeProtecodeScanOptions = executeProtecodeScanOptions{
			CleanupMode:        c.cleanupMode,
			ProtecodeServerURL: c.protecodeServerURL,
			Verbose:            false,
		}

		deleteScan(config, c.productID, client)
		assert.Equal(t, requestURI, c.want)
		if c.cleanupMode == "complete" {
			assert.Contains(t, requestURI, c.productID)
		}
	}
}

func TestCmdStringUploadScanFileSuccess(t *testing.T) {

	os.Setenv("PIPER_user", "usr")
	os.Setenv("PIPER_password", "pwd")
	sEnc := base64.StdEncoding.EncodeToString([]byte("usr:pwd"))

	cases := []struct {
		auth         string
		group        string
		deleteBinary string
		filePath     string
		serverURL    string
		Delimiter    string
		httpCode     string
		cmd          string
	}{
		{sEnc, "group", "true", "path", "URL", protecode.DELIMITER, "%{http_code}", "curl --insecure -H"},
	}
	for _, c := range cases {

		var config executeProtecodeScanOptions = executeProtecodeScanOptions{
			FilePath:           c.filePath,
			ProtecodeServerURL: c.serverURL,
			ProtecodeGroup:     c.group,
			CleanupMode:        "binary",
		}

		got := cmdStringUploadScanFile(config)
		assert.Contains(t, got, c.cmd)
		assert.Contains(t, got, c.auth)
		assert.Contains(t, got, c.group)
		assert.Contains(t, got, c.deleteBinary)
		assert.Contains(t, got, c.filePath)
		assert.Contains(t, got, c.serverURL)
		assert.Contains(t, got, c.Delimiter)
		assert.Contains(t, got, c.httpCode)
	}
}

func TestCmdStringDeclareFetchUrlSuccess(t *testing.T) {

	os.Setenv("PIPER_user", "usr")
	os.Setenv("PIPER_password", "pwd")
	sEnc := base64.StdEncoding.EncodeToString([]byte("usr:pwd"))

	cases := []struct {
		auth         string
		group        string
		deleteBinary string
		fetchURL     string
		serverURL    string
		Delimiter    string
		httpCode     string
		cmd          string
	}{
		{sEnc, "group", "true", "FETCH", "URL", protecode.DELIMITER, "%{http_code}", "curl -X POST -H"},
	}
	for _, c := range cases {

		var config executeProtecodeScanOptions = executeProtecodeScanOptions{
			FetchURL:           c.fetchURL,
			ProtecodeServerURL: c.serverURL,
			ProtecodeGroup:     c.group,
			CleanupMode:        "binary",
		}

		got := cmdStringDeclareFetchUrl(config)
		assert.Contains(t, got, c.cmd)
		assert.Contains(t, got, c.auth)
		assert.Contains(t, got, c.group)
		assert.Contains(t, got, c.deleteBinary)
		assert.Contains(t, got, c.fetchURL)
		assert.Contains(t, got, c.serverURL)
		assert.Contains(t, got, c.Delimiter)
		assert.Contains(t, got, c.httpCode)

	}
}

func TestPollForResultSuccess(t *testing.T) {

	requestURI := ""
	var response protecode.ProteCodeResultData = protecode.ProteCodeResultData{}

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		requestURI = req.RequestURI

		response = protecode.ProteCodeResultData{Result: protecode.ProteCodeResult{ProductId: "productID1", ReportUrl: requestURI, Status: "D", Components: []protecode.ProteCodeComponent{
			{Vulns: []protecode.ProteCodeVulnerability{
				{Triage: "triage"}},
			}},
		}}

		var b bytes.Buffer
		json.NewEncoder(&b).Encode(&response)
		rw.Write([]byte(b.Bytes()))
	}))
	// Close the server when test finishes
	defer server.Close()

	cases := []struct {
		protecodeServerURL string
		productID          string
		want               protecode.ProteCodeResult
	}{
		{server.URL, "productID1", protecode.ProteCodeResult{ProductId: "productID1", ReportUrl: "/api/product/productID1/", Status: "D", Components: []protecode.ProteCodeComponent{
			{Vulns: []protecode.ProteCodeVulnerability{
				{Triage: "triage"}},
			}},
		}},
	}
	client := piperHttp.Client{}
	for _, c := range cases {

		var config executeProtecodeScanOptions = executeProtecodeScanOptions{
			ProtecodeServerURL: c.protecodeServerURL,
		}

		got := pollForResult(config, c.productID, client, 30)
		assert.Equal(t, c.want, got)
		assert.Equal(t, "/api/product/"+c.productID+"/", requestURI)
	}
}
