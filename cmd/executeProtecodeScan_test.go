package cmd

import (
	"testing"

	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
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

		response := protecode.ProductData{
			Products: []protecode.Product{
				{ProductId: 1}},
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
		want               int
	}{
		{server.URL, "filePath", "group", true, 1},
		{server.URL, "filePath", "group32", false, 0},
	}
	for _, c := range cases {

		var config executeProtecodeScanOptions = executeProtecodeScanOptions{
			ProtecodeServerURL: c.protecodeServerURL,
			FilePath:           c.filePath,
			ProtecodeGroup:     c.protecodeGroup,
			ReuseExisting:      c.reuseExisting,
		}

		got, _ := loadExistingProduct(config, client)
		assert.Equal(t, c.want, got)
	}
}
func TestLoadExistingProductByFilenameSuccess(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		response := protecode.ProductData{
			Products: []protecode.Product{
				{ProductId: 1}},
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
		want               *protecode.ProductData
	}{
		{server.URL, "filePath", "group", &protecode.ProductData{
			Products: []protecode.Product{{ProductId: 1}}}},
		{server.URL, "filePäth!", "group32", &protecode.ProductData{
			Products: []protecode.Product{{ProductId: 1}}}},
	}
	for _, c := range cases {

		var config executeProtecodeScanOptions = executeProtecodeScanOptions{
			ProtecodeServerURL: c.protecodeServerURL,
			FilePath:           c.filePath,
			ProtecodeGroup:     c.protecodeGroup,
		}

		got, _ := loadExistingProductByFilename(config, client)
		assert.Equal(t, c.want, got)
	}
}

func TestPullResultSuccess(t *testing.T) {

	requestURI := ""

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		requestURI = req.RequestURI

		var response protecode.ResultData = protecode.ResultData{}

		if strings.Contains(requestURI, "111") {
			response = protecode.ResultData{
				Result: protecode.Result{ProductId: 111, ReportUrl: requestURI}}
		} else {
			response = protecode.ResultData{
				Result: protecode.Result{ProductId: 222, ReportUrl: requestURI}}
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
		productID          int
		want               protecode.Result
	}{
		{server.URL, 111, protecode.Result{ProductId: 111, ReportUrl: "/api/product/111/"}},
		{server.URL, 222, protecode.Result{ProductId: 222, ReportUrl: "/api/product/222/"}},
	}
	for _, c := range cases {

		var config executeProtecodeScanOptions = executeProtecodeScanOptions{
			ProtecodeServerURL: c.protecodeServerURL,
		}

		got, _ := pullResult(config, c.productID, client)
		assert.Equal(t, c.want, got)
		assert.Equal(t, fmt.Sprintf("/api/product/%v/", c.productID), requestURI)
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
		productID          int
		reportFileName     string
		want               string
	}{
		{server.URL, 1, "fileName", "/api/product/1/pdf-report"},
		{server.URL, 2, "fileName", "/api/product/2/pdf-report"},
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
		productID          int
		want               string
	}{
		{"binary", server.URL, 1, ""},
		{"complete", server.URL, 2, "/api/product/2/"},
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
			assert.Contains(t, requestURI, fmt.Sprintf("%v", c.productID))
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

	count := 1
	requestURI := ""
	var response protecode.ResultData = protecode.ResultData{}

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		requestURI = req.RequestURI

		response = protecode.ResultData{Result: protecode.Result{ProductId: 1, ReportUrl: requestURI, Status: "D", Components: []protecode.Component{
			{Vulns: []protecode.Vulnerability{
				{Triage: []protecode.Triage{{Id: 1}}}},
			}},
		}}

		var b bytes.Buffer
		json.NewEncoder(&b).Encode(&response)

		if count == 0 {
			rw.Write([]byte(b.Bytes()))
		} else {
			count--
			rw.Write([]byte(""))
		}
	}))
	// Close the server when test finishes
	defer server.Close()

	cases := []struct {
		protecodeServerURL string
		productID          int
		want               protecode.Result
	}{
		{server.URL, 1, protecode.Result{ProductId: 1, ReportUrl: "/api/product/1/", Status: "D", Components: []protecode.Component{
			{Vulns: []protecode.Vulnerability{
				{Triage: []protecode.Triage{{Id: 1}}}},
			}},
		}},
	}
	client := piperHttp.Client{}
	for _, c := range cases {

		var config executeProtecodeScanOptions = executeProtecodeScanOptions{
			ProtecodeServerURL: c.protecodeServerURL,
		}

		got, _ := pollForResult(config, c.productID, client, 30)
		assert.Equal(t, c.want, got)
		assert.Equal(t, fmt.Sprintf("/api/product/%v/", c.productID), requestURI)
	}
}
