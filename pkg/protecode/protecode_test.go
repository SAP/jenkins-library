package protecode

import (
	"testing"

	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"time"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
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
		{`{"results": {"status": "B", "id": 209396, "product_id": 209396, "report_url": "https://protecode.c.eu-de-2.cloud.sap/products/209396/"}}`, new(ResultData), &ResultData{Result: Result{ProductID: 209396, Status: "B", ReportURL: "https://protecode.c.eu-de-2.cloud.sap/products/209396/"}}},
		{`{"products": [{"product_id": 1}]}`, new(ProductData), &ProductData{Products: []Product{{ProductID: 1}}}},
	}
	pc := Protecode{}
	for _, c := range cases {

		r := ioutil.NopCloser(bytes.NewReader([]byte(c.give)))
		pc.mapResponse(r, c.input)
		assert.Equal(t, c.want, c.input)
	}
}
func TestParseResultSuccess(t *testing.T) {

	var result Result = Result{
		ProductID: 4712,
		ReportURL: "ReportUrl",
		Status:    "B",
		Components: []Component{
			{Vulns: []Vulnerability{
				{Exact: true, Triage: []Triage{}, Vuln: Vuln{Cve: "Cve1", Cvss: 7.2, Cvss3Score: "0.0"}},
				{Exact: true, Triage: []Triage{{ID: 1}}, Vuln: Vuln{Cve: "Cve2", Cvss: 2.2, Cvss3Score: "2.3"}},
				{Exact: true, Triage: []Triage{}, Vuln: Vuln{Cve: "Cve2b", Cvss: 0.0, Cvss3Score: "0.0"}},
			},
			},
			{Vulns: []Vulnerability{
				{Exact: true, Triage: []Triage{}, Vuln: Vuln{Cve: "Cve3", Cvss: 3.2, Cvss3Score: "7.3"}},
				{Exact: true, Triage: []Triage{}, Vuln: Vuln{Cve: "Cve4", Cvss: 8.0, Cvss3Score: "8.0"}},
				{Exact: false, Triage: []Triage{}, Vuln: Vuln{Cve: "Cve4b", Cvss: 8.0, Cvss3Score: "8.0"}},
			},
			},
		},
	}
	pc := Protecode{}
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
	byteContent, err := ioutil.ReadFile(violations)
	if err != nil {
		t.Fatalf("failed reading %v", violations)
	}
	pc := Protecode{}

	resultData := new(ResultData)
	pc.mapResponse(ioutil.NopCloser(strings.NewReader(string(byteContent))), resultData)

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
	byteContent, err := ioutil.ReadFile(noViolations)
	if err != nil {
		t.Fatalf("failed reading %v", noViolations)
	}

	pc := Protecode{}
	resultData := new(ResultData)
	pc.mapResponse(ioutil.NopCloser(strings.NewReader(string(byteContent))), resultData)

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
	byteContent, err := ioutil.ReadFile(triaged)
	if err != nil {
		t.Fatalf("failed reading %v", triaged)
	}

	pc := Protecode{}
	resultData := new(ResultData)
	pc.mapResponse(ioutil.NopCloser(strings.NewReader(string(byteContent))), resultData)

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
				{ProductID: 1}},
		}

		var b bytes.Buffer
		json.NewEncoder(&b).Encode(&response)
		rw.Write([]byte(b.Bytes()))
	}))
	// Close the server when test finishes
	defer server.Close()

	client := &piperHttp.Client{}
	client.SetOptions(piperHttp.ClientOptions{})

	cases := []struct {
		pc             Protecode
		protecodeGroup string
		reuseExisting  bool
		want           int
	}{
		{Protecode{serverURL: server.URL, client: client, logger: log.Entry().WithField("package", "SAP/jenkins-library/pkg/protecode")}, "group", true, 1},
		{Protecode{serverURL: server.URL, client: client}, "group32", false, -1},
	}
	for _, c := range cases {

		got := c.pc.LoadExistingProduct(c.protecodeGroup, c.reuseExisting)
		assert.Equal(t, c.want, got)
	}
}

func TestPollForResultSuccess(t *testing.T) {
	requestURI := ""
	var response ResultData = ResultData{}

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		requestURI = req.RequestURI
		productID := 111
		if strings.Contains(requestURI, "222") {
			productID = 222
		}

		response = ResultData{Result: Result{ProductID: productID, ReportURL: requestURI, Status: "D", Components: []Component{
			{Vulns: []Vulnerability{
				{Triage: []Triage{{ID: 1}}}},
			}},
		}}

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
	}
	// Close the server when test finishes
	defer server.Close()

	client := &piperHttp.Client{}
	client.SetOptions(piperHttp.ClientOptions{})
	pc := Protecode{serverURL: server.URL, client: client, duration: (time.Minute * 1), logger: log.Entry().WithField("package", "SAP/jenkins-library/pkg/protecode")}

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

	client := &piperHttp.Client{}
	client.SetOptions(piperHttp.ClientOptions{})

	cases := []struct {
		pc        Protecode
		productID int
		want      ResultData
	}{
		{Protecode{serverURL: server.URL, client: client}, 111, ResultData{Result: Result{ProductID: 111, ReportURL: "/api/product/111/"}}},
		{Protecode{serverURL: server.URL, client: client}, 222, ResultData{Result: Result{ProductID: 222, ReportURL: "/api/product/222/"}}},
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

	pc := Protecode{}
	po := Options{ServerURL: server.URL}
	pc.SetOptions(po)

	cases := []struct {
		cleanupMode    string
		protecodeGroup string
		fetchURL       string
		want           string
	}{
		{"binary", "group1", "dummy", "/api/fetch/"},
		{"Test", "group2", "dummy", "/api/fetch/"},
	}
	for _, c := range cases {

		pc.DeclareFetchURL(c.cleanupMode, c.protecodeGroup, c.fetchURL)
		assert.Equal(t, requestURI, c.want)
		assert.Contains(t, passedHeaders, "Group")
		assert.Contains(t, passedHeaders, "Delete-Binary")
		assert.Contains(t, passedHeaders, "Url")
	}
}

func TestUploadScanFileSuccess(t *testing.T) {

	requestURI := ""
	var passedHeaders = map[string][]string{}
	var multipartFile multipart.File
	var passedFileContents []byte
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {

		requestURI = req.RequestURI

		passedHeaders = map[string][]string{}
		if req.Header != nil {
			for name, headers := range req.Header {
				passedHeaders[name] = headers
			}
		}

		response := Result{ProductID: 111, ReportURL: requestURI}

		err := req.ParseMultipartForm(4096)
		if err != nil {
			t.FailNow()
		}
		multipartFile, _, err = req.FormFile("file")
		if err != nil {
			t.FailNow()
		}
		defer req.Body.Close()
		passedFileContents, err = ioutil.ReadAll(multipartFile)
		if err != nil {
			t.FailNow()
		}

		var b bytes.Buffer
		json.NewEncoder(&b).Encode(&response)
		rw.Write([]byte(b.Bytes()))
	}))
	// Close the server when test finishes
	defer server.Close()

	pc := Protecode{}
	po := Options{ServerURL: server.URL}
	pc.SetOptions(po)

	testFile, err := ioutil.TempFile("", "testFileUpload")
	if err != nil {
		t.FailNow()
	}
	defer os.RemoveAll(testFile.Name()) // clean up

	fileContents, err := ioutil.ReadFile(testFile.Name())
	if err != nil {
		t.FailNow()
	}

	cases := []struct {
		cleanupMode    string
		protecodeGroup string
		fileName       string
		want           string
	}{
		{"binary", "group1", testFile.Name(), "/api/upload/dummy"},
		{"Test", "group2", testFile.Name(), "/api/upload/dummy"},
	}
	for _, c := range cases {

		pc.UploadScanFile(c.cleanupMode, c.protecodeGroup, c.fileName, "dummy")
		assert.Equal(t, requestURI, c.want)
		assert.Contains(t, passedHeaders, "Group")
		assert.Contains(t, passedHeaders, "Delete-Binary")
		assert.Equal(t, fileContents, passedFileContents, "Uploaded file incorrect")
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

	client := &piperHttp.Client{}
	client.SetOptions(piperHttp.ClientOptions{})

	pc := Protecode{serverURL: server.URL, client: client}

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

	pc := Protecode{}
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
