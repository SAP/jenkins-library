package cmd

import (
	"testing"

	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	//"github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/protecode"
	"github.com/stretchr/testify/assert"
)

func TestLoadExistingProductByFilenameSuccess(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		response := protecode.ProteCodeResultData{
			Result: protecode.ProteCodeResult{ProductId: "test", ReportUrl: "ReportUrl_Test"}}

		var b bytes.Buffer
		json.NewEncoder(&b).Encode(&response)
		rw.Write([]byte(b.Bytes()))
	}))
	// Close the server when test finishes
	defer server.Close()

	client := protecode.Client{}

	cases := []struct {
		protecodeServerURL string
		filePath           string
		protecodeGroup     string
		want               *protecode.ProteCodeResultData
	}{
		{server.URL, "filePath", "group", &protecode.ProteCodeResultData{
			Result: protecode.ProteCodeResult{ProductId: "test", ReportUrl: "ReportUrl_Test"}}},
		{server.URL, "filePäth!", "group32", &protecode.ProteCodeResultData{
			Result: protecode.ProteCodeResult{ProductId: "test", ReportUrl: "ReportUrl_Test"}}},
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

	client := protecode.Client{}

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

	client := protecode.Client{}

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
		assert.Equal(t, c.want, requestURI)
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

	client := protecode.Client{}

	cases := []struct {
		cleanupMode            string
		protecodeServerURL     string
		productID              string
		protecodeCredentialsID string
		want                   string
	}{
		{"binary", server.URL, "productID1", "credentialsID1", ""},
		{"complete", server.URL, "productID2", "credentialsID1", "/api/product/productID2/"},
	}
	for _, c := range cases {

		var config executeProtecodeScanOptions = executeProtecodeScanOptions{
			CleanupMode:            c.cleanupMode,
			ProtecodeServerURL:     c.protecodeServerURL,
			Verbose:                false,
			ProtecodeCredentialsID: c.protecodeCredentialsID,
		}

		deleteScan(config, c.productID, client)
		assert.Equal(t, c.want, requestURI)
		if c.cleanupMode == "complete" {
			assert.Contains(t, passedHeaders, "Httpmode")
		}
	}
}

func TestCmdStringUploadScanFileSuccess(t *testing.T) {

	cases := []struct {
		auth         string
		callback     string
		group        string
		deleteBinary string
		filePath     string
		serverURL    string
		Delimiter    string
		httpCode     string
		want         string
	}{
		{"auth", "" /* Callback */, "group", "true", "path", "URL", protecode.DELIMITER, "%{http_code}", "curl --insecure -H 'Authorization: Basic auth'  -H 'Group: group' -H 'Delete-Binary: true' -T path URL/api/upload/ --write-out '-DeLiMiTeR-status=%{http_code}'"},
	}
	for _, c := range cases {

		var config executeProtecodeScanOptions = executeProtecodeScanOptions{
			FilePath:           c.filePath,
			ProtecodeServerURL: c.serverURL,
			ProtecodeGroup:     c.group,
			CleanupMode:        "binary",
		}

		got := cmdStringUploadScanFile(config)
		assert.Equal(t, c.want, got)
	}
}

func TestCmdStringDeclareFetchUrlSuccess(t *testing.T) {

	cases := []struct {
		auth         string
		callback     string
		group        string
		deleteBinary string
		fetchURL     string
		serverURL    string
		Delimiter    string
		httpCode     string
		want         string
	}{
		{"auth", "" /* Callback */, "group", "true", "FETCH", "URL", protecode.DELIMITER, "%{http_code}", "curl -X POST -H 'Authorization: Basic auth'  -H 'Group: group' -H 'Delete-Binary: true' -H 'Url:FETCH'  URL/api/fetch/ --write-out '-DeLiMiTeR-status=%{http_code}'"},
	}
	for _, c := range cases {

		var config executeProtecodeScanOptions = executeProtecodeScanOptions{
			FetchURL:           c.fetchURL,
			ProtecodeServerURL: c.serverURL,
			ProtecodeGroup:     c.group,
			CleanupMode:        "binary",
		}

		got := cmdStringDeclareFetchUrl(config)
		assert.Equal(t, c.want, got)
	}
}
