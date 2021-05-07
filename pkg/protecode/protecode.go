package protecode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/sirupsen/logrus"
)

// ProductData holds the product information of the protecode product
type ProductData struct {
	Products []Product `json:"products,omitempty"`
}

// Product holds the id of the protecode product
type Product struct {
	ProductID int `json:"product_id,omitempty"`
}

//ResultData holds the information about the protecode result
type ResultData struct {
	Result Result `json:"results,omitempty"`
}

//Result holds the detail information about the protecode result
type Result struct {
	ProductID  int         `json:"product_id,omitempty"`
	ReportURL  string      `json:"report_url,omitempty"`
	Status     string      `json:"status,omitempty"`
	Components []Component `json:"components,omitempty"`
}

//Component the protecode component information
type Component struct {
	Vulns []Vulnerability `json:"vulns,omitempty"`
}

//Vulnerability the protecode vulnerability information
type Vulnerability struct {
	Exact  bool     `json:"exact,omitempty"`
	Vuln   Vuln     `json:"vuln,omitempty"`
	Triage []Triage `json:"triage,omitempty"`
}

//Vuln holds the inforamtion about the vulnerability
type Vuln struct {
	Cve        string  `json:"cve,omitempty"`
	Cvss       float64 `json:"cvss,omitempty"`
	Cvss3Score string  `json:"cvss3_score,omitempty"`
}

//Triage holds the triaging information
type Triage struct {
	ID          int    `json:"id,omitempty"`
	VulnID      string `json:"vuln_id,omitempty"`
	Component   string `json:"component,omitempty"`
	Vendor      string `json:"vendor,omitempty"`
	Codetype    string `json:"codetype,omitempty"`
	Version     string `json:"version,omitempty"`
	Modified    string `json:"modified,omitempty"`
	Scope       string `json:"scope,omitempty"`
	Description string `json:"description,omitempty"`
	User        User   `json:"user,omitempty"`
}

//User holds the user information
type User struct {
	ID        int    `json:"id,omitempty"`
	Email     string `json:"email,omitempty"`
	Girstname string `json:"firstname,omitempty"`
	Lastname  string `json:"lastname,omitempty"`
	Username  string `json:"username,omitempty"`
}

//Protecode ist the protecode client which is used by the step
type Protecode struct {
	serverURL string
	client    piperHttp.Uploader
	duration  time.Duration
	logger    *logrus.Entry
}

//Options struct which can be used to configure the Protecode struct
type Options struct {
	ServerURL string
	Duration  time.Duration
	Username  string
	Password  string
	Logger    *logrus.Entry
}

//SetOptions setter function to set the internal properties of the protecode
func (pc *Protecode) SetOptions(options Options) {
	pc.serverURL = options.ServerURL
	pc.client = &piperHttp.Client{}
	pc.duration = options.Duration

	if options.Logger != nil {
		pc.logger = options.Logger
	} else {
		pc.logger = log.Entry().WithField("package", "SAP/jenkins-library/pkg/protecode")
	}

	httpOptions := piperHttp.ClientOptions{MaxRequestDuration: options.Duration, Username: options.Username, Password: options.Password, Logger: options.Logger}
	pc.client.SetOptions(httpOptions)
}

func (pc *Protecode) createURL(path string, pValue string, fParam string) string {

	protecodeURL, err := url.Parse(pc.serverURL)
	if err != nil {
		pc.logger.WithError(err).Fatal("Malformed URL")
	}

	if len(path) > 0 {
		protecodeURL.Path += fmt.Sprintf("%v", path)
	}

	if len(pValue) > 0 {
		protecodeURL.Path += fmt.Sprintf("%v", pValue)
	}

	// Prepare Query Parameters
	if len(fParam) > 0 {
		encodedFParam := url.QueryEscape(fParam)
		params := url.Values{}
		params.Add("q", fmt.Sprintf("file:%v", encodedFParam))

		// Add Query Parameters to the URL
		protecodeURL.RawQuery = params.Encode() // Escape Query Parameters
	}

	return protecodeURL.String()
}

func (pc *Protecode) mapResponse(r io.ReadCloser, response interface{}) {
	defer r.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	newStr := buf.String()
	if len(newStr) > 0 {

		unquoted, err := strconv.Unquote(newStr)
		if err != nil {
			err = json.Unmarshal([]byte(newStr), response)
			if err != nil {
				pc.logger.WithError(err).Fatalf("Error during unqote response: %v", newStr)
			}
		} else {
			err = json.Unmarshal([]byte(unquoted), response)
		}

		if err != nil {
			pc.logger.WithError(err).Fatalf("Error during decode response: %v", newStr)
		}
	}
}

func (pc *Protecode) sendAPIRequest(method string, url string, headers map[string][]string) (*io.ReadCloser, error) {

	r, err := pc.client.SendRequest(method, url, nil, headers, nil)
	if err != nil {
		return nil, err
	}

	return &r.Body, nil
}

// ParseResultForInflux parses the result from the scan into the internal format
func (pc *Protecode) ParseResultForInflux(result Result, excludeCVEs string) (map[string]int, []Vuln) {

	var vulns []Vuln

	var m map[string]int = make(map[string]int)
	m["count"] = 0
	m["cvss2GreaterOrEqualSeven"] = 0
	m["cvss3GreaterOrEqualSeven"] = 0
	m["historical_vulnerabilities"] = 0
	m["triaged_vulnerabilities"] = 0
	m["excluded_vulnerabilities"] = 0
	m["minor_vulnerabilities"] = 0
	m["major_vulnerabilities"] = 0
	m["vulnerabilities"] = 0

	for _, components := range result.Components {
		for _, vulnerability := range components.Vulns {

			exact := isExact(vulnerability)
			countVulnerability := isExact(vulnerability) && !isExcluded(vulnerability, excludeCVEs) && !isTriaged(vulnerability)

			if exact && isExcluded(vulnerability, excludeCVEs) {
				m["excluded_vulnerabilities"]++
			}
			if exact && isTriaged(vulnerability) {
				m["triaged_vulnerabilities"]++
			}
			if countVulnerability {
				m["count"]++
				m["vulnerabilities"]++

				//collect all vulns here
				vulns = append(vulns, vulnerability.Vuln)
			}
			if countVulnerability && isSevereCVSS3(vulnerability) {
				m["cvss3GreaterOrEqualSeven"]++
				m["major_vulnerabilities"]++
			}
			if countVulnerability && isSevereCVSS2(vulnerability) {
				m["cvss2GreaterOrEqualSeven"]++
				m["major_vulnerabilities"]++
			}
			if countVulnerability && !isSevereCVSS3(vulnerability) && !isSevereCVSS2(vulnerability) {
				m["minor_vulnerabilities"]++
			}
			if !exact {
				m["historical_vulnerabilities"]++
			}
		}
	}

	return m, vulns
}

func isExact(vulnerability Vulnerability) bool {
	return vulnerability.Exact
}

func isExcluded(vulnerability Vulnerability, excludeCVEs string) bool {
	return strings.Contains(excludeCVEs, vulnerability.Vuln.Cve)
}

func isTriaged(vulnerability Vulnerability) bool {
	return len(vulnerability.Triage) > 0
}

func isSevereCVSS3(vulnerability Vulnerability) bool {
	threshold := 7.0
	cvss3, _ := strconv.ParseFloat(vulnerability.Vuln.Cvss3Score, 64)
	return cvss3 >= threshold
}

func isSevereCVSS2(vulnerability Vulnerability) bool {
	threshold := 7.0
	cvss3, _ := strconv.ParseFloat(vulnerability.Vuln.Cvss3Score, 64)
	return cvss3 == 0 && vulnerability.Vuln.Cvss >= threshold
}

// DeleteScan deletes if configured the scan on the protecode server
func (pc *Protecode) DeleteScan(cleanupMode string, productID int) {
	switch cleanupMode {
	case "none":
	case "binary":
	case "complete":
		pc.logger.Info("Deleting scan from server.")
		protecodeURL := pc.createURL("/api/product/", fmt.Sprintf("%v/", productID), "")
		headers := map[string][]string{}

		pc.sendAPIRequest("DELETE", protecodeURL, headers)
	default:
		pc.logger.Fatalf("Unknown cleanup mode %v", cleanupMode)
	}
}

// LoadReport loads the report of the protecode scan
func (pc *Protecode) LoadReport(reportFileName string, productID int) *io.ReadCloser {

	protecodeURL := pc.createURL("/api/product/", fmt.Sprintf("%v/pdf-report", productID), "")
	headers := map[string][]string{
		"Cache-Control": {"no-cache, no-store, must-revalidate"},
		"Pragma":        {"no-cache"},
		"Outputfile":    {reportFileName},
	}

	readCloser, err := pc.sendAPIRequest(http.MethodGet, protecodeURL, headers)
	if err != nil {
		pc.logger.WithError(err).Fatalf("It is not possible to load report %v", protecodeURL)
	}

	return readCloser
}

// UploadScanFile upload the scan file to the protecode server
func (pc *Protecode) UploadScanFile(cleanupMode, group, filePath, fileName string) *ResultData {
	deleteBinary := (cleanupMode == "binary" || cleanupMode == "complete")
	headers := map[string][]string{"Group": {group}, "Delete-Binary": {fmt.Sprintf("%v", deleteBinary)}}

	uploadURL := fmt.Sprintf("%v/api/upload/%v", pc.serverURL, fileName)

	r, err := pc.client.UploadRequest(http.MethodPut, uploadURL, filePath, "file", headers, nil)
	if err != nil {
		pc.logger.WithError(err).Fatalf("Error during %v upload request", uploadURL)
	} else {
		pc.logger.Info("Upload successful")
	}

	result := new(ResultData)
	pc.mapResponse(r.Body, result)

	return result
}

// DeclareFetchURL configures the fetch url for the protecode scan
func (pc *Protecode) DeclareFetchURL(cleanupMode, group, fetchURL string) *ResultData {
	deleteBinary := (cleanupMode == "binary" || cleanupMode == "complete")
	headers := map[string][]string{"Group": {group}, "Delete-Binary": {fmt.Sprintf("%v", deleteBinary)}, "Url": {fetchURL}, "Content-Type": {"application/json"}}

	protecodeURL := fmt.Sprintf("%v/api/fetch/", pc.serverURL)
	r, err := pc.sendAPIRequest(http.MethodPost, protecodeURL, headers)
	if err != nil {
		pc.logger.WithError(err).Fatalf("Error during declare fetch url: %v", protecodeURL)
	}

	result := new(ResultData)
	pc.mapResponse(*r, result)

	return result
}

//PollForResult polls the protecode scan for the result scan
func (pc *Protecode) PollForResult(productID int, timeOutInMinutes string) ResultData {

	var response ResultData
	var err error

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	var ticks int64 = 6
	if len(timeOutInMinutes) > 0 {
		parsedTimeOutInMinutes, _ := strconv.ParseInt(timeOutInMinutes, 10, 64)
		ticks = parsedTimeOutInMinutes * 6
	}

	pc.logger.Infof("Poll for result %v times", ticks)

	for i := ticks; i > 0; i-- {

		response, err = pc.pullResult(productID)
		if err != nil {
			ticker.Stop()
			i = 0
			return response
		}
		if len(response.Result.Components) > 0 && response.Result.Status != statusBusy {
			ticker.Stop()
			i = 0
			break
		}

		select {
		case t := <-ticker.C:
			pc.logger.Debugf("Tick : %v Processing status for productID %v", t, productID)
		}
	}

	if len(response.Result.Components) == 0 || response.Result.Status == statusBusy {
		response, err = pc.pullResult(productID)
		if err != nil || len(response.Result.Components) == 0 || response.Result.Status == statusBusy {
			pc.logger.Fatal("No result after polling")
		}
	}

	return response
}

func (pc *Protecode) pullResult(productID int) (ResultData, error) {
	protecodeURL := pc.createURL("/api/product/", fmt.Sprintf("%v/", productID), "")
	headers := map[string][]string{
		"acceptType": {"application/json"},
	}
	r, err := pc.sendAPIRequest(http.MethodGet, protecodeURL, headers)
	if err != nil {
		return *new(ResultData), err
	}
	result := new(ResultData)
	pc.mapResponse(*r, result)

	return *result, nil

}

// LoadExistingProduct loads the existing product from protecode service
func (pc *Protecode) LoadExistingProduct(group string, verifyOnly bool) int {
	var productID int = -1

	if verifyOnly {

		protecodeURL := pc.createURL("/api/apps/", fmt.Sprintf("%v/", group), "")
		headers := map[string][]string{
			"acceptType": {"application/json"},
		}

		response := pc.loadExisting(protecodeURL, headers)
		// by definition we will take the first one and trigger rescan
		productID = response.Products[0].ProductID

		pc.logger.Infof("Re-use existing Protecode scan - group: %v, productID: %v", group, productID)
	}

	return productID
}

func (pc *Protecode) loadExisting(protecodeURL string, headers map[string][]string) *ProductData {

	r, err := pc.sendAPIRequest(http.MethodGet, protecodeURL, headers)
	if err != nil {
		pc.logger.WithError(err).Fatalf("Error during load existing product: %v", protecodeURL)
	}

	result := new(ProductData)
	pc.mapResponse(*r, result)

	return result
}
