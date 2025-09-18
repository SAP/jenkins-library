package protecode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMapResponse(t *testing.T) {

	cases := []struct {
		give  string
		input interface{}
		want  interface{}
	}{
		{`"{}"`, new(Result), &Result{ProductID: 0}},
		{`{"product_id": 1}`, new(Result), &Result{ProductID: 1}},
		{`"{\"product_id\": 4711}"`, new(Result), &Result{ProductID: 4711}},
		{"{\"results\": {\"product_id\": 1}}", new(ResultData), &ResultData{Result: Result{ProductID: 1}}},
		{`{"results": {"status": "B", "id": 209396, "product_id": 209396, "report_url": "https://protecode.c.eu-de-2.cloud.sap/products/209396/"}}`, new(ResultData), &ResultData{Result: Result{ProductID: 209396, Status: statusBusy, ReportURL: "https://protecode.c.eu-de-2.cloud.sap/products/209396/"}}},
		{`{"products": [{"product_id": 1}]}`, new(ProductData), &ProductData{Products: []Product{{ProductID: 1}}}},
	}
	pc := makeProtecode(Options{})
	for _, c := range cases {

		r := io.NopCloser(bytes.NewReader([]byte(c.give)))
		pc.mapResponse(r, c.input)
		assert.Equal(t, c.want, c.input)
	}
}

func TestParseResultSuccess(t *testing.T) {

	var result Result = Result{
		ProductID: 4712,
		ReportURL: "ReportUrl",
		Status:    statusBusy,
		Components: []Component{
			{Vulns: []Vulnerability{
				{Exact: true, Triage: []Triage{}, Vuln: Vuln{Cve: "Cve1", Cvss: "7.2", Cvss3Score: "0.0"}},
				{Exact: true, Triage: []Triage{{ID: 1}}, Vuln: Vuln{Cve: "Cve2", Cvss: "2.2", Cvss3Score: "2.3"}},
				{Exact: true, Triage: []Triage{}, Vuln: Vuln{Cve: "Cve2b", Cvss: "0.0", Cvss3Score: "0.0"}},
			},
			},
			{Vulns: []Vulnerability{
				{Exact: true, Triage: []Triage{}, Vuln: Vuln{Cve: "Cve3", Cvss: "3.2", Cvss3Score: "7.3"}},
				{Exact: true, Triage: []Triage{}, Vuln: Vuln{Cve: "Cve4", Cvss: "8.0", Cvss3Score: "8.0"}},
				{Exact: false, Triage: []Triage{}, Vuln: Vuln{Cve: "Cve4b", Cvss: "8.0", Cvss3Score: "8.0"}},
			},
			},
		},
	}
	pc := makeProtecode(Options{})
	m, vulns := pc.ParseResultForInflux(result, "Excluded CVES: Cve4,")
	t.Run("Parse Protecode Results", func(t *testing.T) {
		assert.Equal(t, 1, m["historical_vulnerabilities"])
		assert.Equal(t, 1, m["triaged_vulnerabilities"])
		assert.Equal(t, 1, m["excluded_vulnerabilities"])
		assert.Equal(t, 1, m["minor_vulnerabilities"])
		assert.Equal(t, 2, m["major_vulnerabilities"])
		assert.Equal(t, 3, m["vulnerabilities"])

		assert.Equal(t, 3, len(vulns))
	})
}

func TestParseResultViolations(t *testing.T) {

	violations := filepath.Join("testdata", "protecode_result_violations.json")
	byteContent, err := os.ReadFile(violations)
	if err != nil {
		t.Fatalf("failed reading %v", violations)
	}
	pc := makeProtecode(Options{})

	resultData := new(ResultData)
	pc.mapResponse(io.NopCloser(strings.NewReader(string(byteContent))), resultData)

	m, vulns := pc.ParseResultForInflux(resultData.Result, "CVE-2018-1, CVE-2017-1000382")
	t.Run("Parse Protecode Results", func(t *testing.T) {
		assert.Equal(t, 1125, m["historical_vulnerabilities"])
		assert.Equal(t, 0, m["triaged_vulnerabilities"])
		assert.Equal(t, 1, m["excluded_vulnerabilities"])
		assert.Equal(t, 129, m["cvss3GreaterOrEqualSeven"])
		assert.Equal(t, 13, m["cvss2GreaterOrEqualSeven"])
		assert.Equal(t, 226, m["vulnerabilities"])

		assert.Equal(t, 226, len(vulns))
	})
}

func TestParseResultNoViolations(t *testing.T) {

	noViolations := filepath.Join("testdata", "protecode_result_no_violations.json")
	byteContent, err := os.ReadFile(noViolations)
	if err != nil {
		t.Fatalf("failed reading %v", noViolations)
	}

	pc := makeProtecode(Options{})
	resultData := new(ResultData)
	pc.mapResponse(io.NopCloser(strings.NewReader(string(byteContent))), resultData)

	m, vulns := pc.ParseResultForInflux(resultData.Result, "CVE-2018-1, CVE-2017-1000382")
	t.Run("Parse Protecode Results", func(t *testing.T) {
		assert.Equal(t, 27, m["historical_vulnerabilities"])
		assert.Equal(t, 0, m["triaged_vulnerabilities"])
		assert.Equal(t, 0, m["excluded_vulnerabilities"])
		assert.Equal(t, 0, m["cvss3GreaterOrEqualSeven"])
		assert.Equal(t, 0, m["cvss2GreaterOrEqualSeven"])
		assert.Equal(t, 0, m["vulnerabilities"])

		assert.Equal(t, 0, len(vulns))
	})
}

func TestParseResultTriaged(t *testing.T) {

	triaged := filepath.Join("testdata", "protecode_result_triaging.json")
	byteContent, err := os.ReadFile(triaged)
	if err != nil {
		t.Fatalf("failed reading %v", triaged)
	}

	pc := makeProtecode(Options{})
	resultData := new(ResultData)
	pc.mapResponse(io.NopCloser(strings.NewReader(string(byteContent))), resultData)

	m, vulns := pc.ParseResultForInflux(resultData.Result, "")
	t.Run("Parse Protecode Results", func(t *testing.T) {
		assert.Equal(t, 1132, m["historical_vulnerabilities"])
		assert.Equal(t, 187, m["triaged_vulnerabilities"])
		assert.Equal(t, 0, m["excluded_vulnerabilities"])
		assert.Equal(t, 15, m["cvss3GreaterOrEqualSeven"])
		assert.Equal(t, 0, m["cvss2GreaterOrEqualSeven"])
		assert.Equal(t, 36, m["vulnerabilities"])

		assert.Equal(t, 36, len(vulns))
	})
}

func TestLoadExistingProductSuccess(t *testing.T) {

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		response := ProductData{
			Products: []Product{
				{ProductID: 1, FileName: "file_1.zip"},
			},
		}

		var b bytes.Buffer
		json.NewEncoder(&b).Encode(&response)
		rw.Write([]byte(b.Bytes()))
	}))
	// Close the server when test finishes
	defer server.Close()

	cases := []struct {
		pc             Protecode
		protecodeGroup string
		fileName       string
		want           int
	}{
		{makeProtecode(Options{ServerURL: server.URL}), "group", "file_1.zip", 1},
	}
	for _, c := range cases {

		got := c.pc.LoadExistingProduct(c.protecodeGroup, c.fileName)
		assert.Equal(t, c.want, got)
	}
}

func TestPollForResultSuccess(t *testing.T) {
	requestURI := ""
	var response ResultData = ResultData{}

	protecodePollInterval = time.Nanosecond

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		requestURI = req.RequestURI
		productID := 111

		// 2021-04-20 d :
		// Added case '333' to test proper handling of case where Component list is empty.
		if strings.Contains(requestURI, "222") {
			productID = 222
		} else if strings.Contains(requestURI, "333") {
			productID = 333
		}

		var cmpnts []Component
		if productID != 333 {
			cmpnts = []Component{
				{Vulns: []Vulnerability{
					{Triage: []Triage{{ID: 1}}}},
				}}
		}

		response = ResultData{Result: Result{ProductID: productID, ReportURL: requestURI, Status: "D", Components: cmpnts}}

		var b bytes.Buffer
		json.NewEncoder(&b).Encode(&response)
		rw.Write([]byte(b.Bytes()))

	}))

	cases := []struct {
		productID int
		want      ResultData
	}{
		{111, ResultData{Result: Result{ProductID: 111, ReportURL: "/api/product/111/", Status: "D", Components: []Component{
			{Vulns: []Vulnerability{
				{Triage: []Triage{{ID: 1}}}},
			}},
		}}},
		{222, ResultData{Result: Result{ProductID: 222, ReportURL: "/api/product/222/", Status: "D", Components: []Component{
			{Vulns: []Vulnerability{
				{Triage: []Triage{{ID: 1}}}},
			}},
		}}},
		{333, ResultData{Result: Result{ProductID: 333, ReportURL: "/api/product/333/", Status: "D"}}},
	}
	// Close the server when test finishes
	defer server.Close()

	pc := makeProtecode(Options{ServerURL: server.URL, Duration: (time.Minute * 1)})

	for _, c := range cases {
		got := pc.PollForResult(c.productID, "1")
		assert.Equal(t, c.want, got)
		assert.Equal(t, fmt.Sprintf("/api/product/%v/", c.productID), requestURI)
	}
}

func TestPullResultSuccess(t *testing.T) {

	requestURI := ""

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		requestURI = req.RequestURI

		var response ResultData = ResultData{}

		if strings.Contains(requestURI, "111") {
			response = ResultData{
				Result: Result{ProductID: 111, ReportURL: requestURI}}
		} else {
			response = ResultData{
				Result: Result{ProductID: 222, ReportURL: requestURI}}
		}

		var b bytes.Buffer
		json.NewEncoder(&b).Encode(&response)
		rw.Write([]byte(b.Bytes()))
	}))
	// Close the server when test finishes
	defer server.Close()

	cases := []struct {
		pc        Protecode
		productID int
		want      ResultData
	}{
		{makeProtecode(Options{ServerURL: server.URL}), 111, ResultData{Result: Result{ProductID: 111, ReportURL: "/api/product/111/"}}},
		{makeProtecode(Options{ServerURL: server.URL}), 222, ResultData{Result: Result{ProductID: 222, ReportURL: "/api/product/222/"}}},
	}
	for _, c := range cases {

		got, _ := c.pc.pullResult(c.productID)
		assert.Equal(t, c.want, got)
		assert.Equal(t, fmt.Sprintf("/api/product/%v/", c.productID), requestURI)
	}
}

func TestDeclareFetchURLSuccess(t *testing.T) {

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

		response := ResultData{Result: Result{ProductID: 111, ReportURL: requestURI}}

		var b bytes.Buffer
		json.NewEncoder(&b).Encode(&response)
		rw.Write([]byte(b.Bytes()))
	}))
	// Close the server when test finishes
	defer server.Close()
	pc := makeProtecode(Options{ServerURL: server.URL})

	cases := []struct {
		cleanupMode       string
		protecodeGroup    string
		customDataJSONMap string
		fetchURL          string
		version           string
		productID         int
		replaceBinary     bool
		want              int
	}{
		{"binary", "group1", `{"custom-header": "custom-value"}`, "/api/fetch/", "", 1, true, 111},
		{"binary", "group1", "", "/api/fetch/", "custom-test-version", -1, true, 111},
		{"binary", "group1", "", "/api/fetch/", "1.2.3", 0, true, 111},

		{"binary", "group1", "", "/api/fetch/", "", 1, false, 111},
		{"binary", "group1", "", "/api/fetch/", "custom-test-version", -1, false, 111},
		{"binary", "group1", "", "/api/fetch/", "1.2.3", 0, false, 111},
	}
	for _, c := range cases {

		// pc.DeclareFetchURL(c.cleanupMode, c.protecodeGroup, c.fetchURL)
		got := pc.DeclareFetchURL(c.cleanupMode, c.protecodeGroup, c.customDataJSONMap, c.fetchURL, c.version, c.productID, c.replaceBinary)

		assert.Equal(t, requestURI, "/api/fetch/")
		assert.Equal(t, got.Result.ProductID, c.want)
		assert.Equal(t, got.Result.Status, "")

		assert.Contains(t, passedHeaders, "Group")
		assert.Contains(t, passedHeaders, "Delete-Binary")
		assert.Contains(t, passedHeaders, "Url")

		if c.replaceBinary {
			assert.Contains(t, passedHeaders, "Replace")
		}

	}
}

func TestUploadScanFileSuccess(t *testing.T) {

	requestURI := ""
	var passedHeaders = map[string][]string{}
	var reader io.Reader
	var passedFileContents []byte
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		requestURI = req.RequestURI

		passedHeaders = map[string][]string{}
		if req.Header != nil {
			for name, headers := range req.Header {
				passedHeaders[name] = headers
			}
		}

		response := new(ResultData)

		err := req.ParseMultipartForm(4096)
		if err == nil {
			reader, _, err = req.FormFile("file")
			if err != nil {
				t.FailNow()
			}
		} else {
			reader = req.Body
		}

		defer req.Body.Close()
		passedFileContents, err = io.ReadAll(reader)
		if err != nil {
			t.FailNow()
		}

		// When replace binary option is true then mock server should return same product id with http status code 201 (not 200)
		if strReplaceID, isExist := passedHeaders["Replace"]; isExist {
			// convert string to int
			intReplaceID, _ := strconv.Atoi(strReplaceID[0])

			response.Result.ProductID = intReplaceID
			response.Result.ReportURL = requestURI

			rw.WriteHeader(201)

		} else {
			response.Result.ProductID = 112
			response.Result.ReportURL = requestURI

			rw.WriteHeader(200)
		}

		var b bytes.Buffer
		json.NewEncoder(&b).Encode(&response)
		rw.Write([]byte(b.Bytes()))

	}))
	// Close the server when test finishes
	defer server.Close()
	pc := makeProtecode(Options{ServerURL: server.URL})

	testFile, err := os.CreateTemp("", "testFileUpload")
	if err != nil {
		t.FailNow()
	}
	defer os.RemoveAll(testFile.Name()) // clean up

	fileContents, err := os.ReadFile(testFile.Name())
	if err != nil {
		t.FailNow()
	}

	cases := []struct {
		cleanupMode       string
		protecodeGroup    string
		customDataJSONMap string
		filePath          string
		version           string
		productID         int
		replaceBinary     bool
		want              int
	}{
		{"binary", "group1", `{"custom-header": "custom-value"}`, testFile.Name(), "", 1, true, 1},
		{"binary", "group1", "", testFile.Name(), "custom-test-version", 0, true, 0},
		{"binary", "group1", "", testFile.Name(), "1.2.3", -1, true, -1},

		{"binary", "group1", "", testFile.Name(), "", 1, false, 112},
		{"binary", "group1", "", testFile.Name(), "custom-test-version", 0, false, 112},
		{"binary", "group1", "", testFile.Name(), "1.2.3", -1, false, 112},

		// {"binary", "group1", testFile.Name(), "/api/upload/dummy"},
		// {"Test", "group2", testFile.Name(), "/api/upload/dummy"},
	}
	for _, c := range cases {

		got := pc.UploadScanFile(c.cleanupMode, c.protecodeGroup, c.customDataJSONMap, c.filePath, "dummy.tar", c.version, c.productID, c.replaceBinary)

		assert.Equal(t, requestURI, "/api/upload/dummy.tar")
		assert.Contains(t, passedHeaders, "Group")
		assert.Contains(t, passedHeaders, "Delete-Binary")
		assert.Equal(t, fileContents, passedFileContents, "Uploaded file incorrect")
		assert.Equal(t, c.want, got.Result.ProductID)
		assert.Equal(t, "", got.Result.Status)

		if c.replaceBinary {
			assert.Contains(t, passedHeaders, "Replace")
		}
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

	pc := makeProtecode(Options{ServerURL: server.URL})

	cases := []struct {
		productID      int
		reportFileName string
		want           string
	}{
		{1, "fileName", "/api/product/1/pdf-report"},
		{2, "fileName", "/api/product/2/pdf-report"},
	}
	for _, c := range cases {

		pc.LoadReport(c.reportFileName, c.productID)
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

	pc := makeProtecode(Options{})
	po := Options{ServerURL: server.URL}
	pc.SetOptions(po)

	cases := []struct {
		cleanupMode string
		productID   int
		want        string
	}{
		{"binary", 1, ""},
		{"complete", 2, "/api/product/2/"},
	}
	for _, c := range cases {

		pc.DeleteScan(c.cleanupMode, c.productID)
		assert.Equal(t, requestURI, c.want)
		if c.cleanupMode == "complete" {
			assert.Contains(t, requestURI, fmt.Sprintf("%v", c.productID))
		}
	}
}
