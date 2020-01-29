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
	Products []Product `json:"products"`
}

// Product holds the id of the protecode product
type Product struct {
	ProductID int `json:"product_id"`
}

//ResultData holds the information about the protecode result
type ResultData struct {
	Result Result `json:"results"`
}

//Result holds the detail information about the protecode result
type Result struct {
	ProductID  int         `json:"product_id"`
	ReportURL  string      `json:"report_url"`
	Status     string      `json:"status"`
	Components []Component `json:"components,omitempty"`
}

//Component the protecode component information
type Component struct {
	Vulns []Vulnerability `json:"vulns,omitempty"`
}

//Vulnerability the protecode vulnerability information
type Vulnerability struct {
	Exact  bool     `json:"exact"`
	Vuln   Vuln     `json:"vuln"`
	Triage []Triage `json:"triage,omitempty"`
}

//Vuln holds the inforamtion about the vulnerability
type Vuln struct {
	Cve        string  `json:"cve"`
	Cvss       float64 `json:"cvss"`
	Cvss3Score string  `json:"cvss3_score"`
}

//Triage holds the triaging information
type Triage struct {
	ID          int    `json:"id"`
	VulnID      string `json:"vuln_id"`
	Component   string `json:"component"`
	Vendor      string `json:"vendor"`
	Codetype    string `json:"codetype"`
	Version     string `json:"version"`
	Modified    string `json:"modified"`
	Scope       string `json:"scope"`
	Description string `json:"description"`
	User        User   `json:"user"`
}

//User holds the user information
type User struct {
	ID        int    `json:"id"`
	Email     string `json:"email"`
	Girstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Username  string `json:"username"`
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

	httpOptions := piperHttp.ClientOptions{options.Duration, options.Username, options.Password, "", options.Logger}
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

func (pc *Protecode) getResultData(r io.ReadCloser) *ResultData {
	defer r.Close()
	response := new(ResultData)

	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	newStr := buf.String()

	if len(newStr) > 0 {

		unquoted, err := strconv.Unquote(newStr)
		if err != nil {
			err = json.Unmarshal([]byte(newStr), response)
			if err != nil {
				pc.logger.WithError(err).Fatalf("Protecode scan failed, error during unqote response: %v", newStr)
			}
		} else {
			err = json.Unmarshal([]byte(unquoted), response)
		}

		if err != nil {
			pc.logger.WithError(err).Fatalf("Protecode scan failed, error during decode response: %v", newStr)
		}
	}

	return response
}

func (pc *Protecode) getResult(r io.ReadCloser) *Result {
	defer r.Close()
	response := new(Result)

	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	newStr := buf.String()

	if len(newStr) > 0 {

		unquoted, err := strconv.Unquote(newStr)
		if err != nil {
			err = json.Unmarshal([]byte(newStr), response)
			if err != nil {
				pc.logger.WithError(err).Fatalf("Protecode scan failed, error during unqote response: %v", newStr)
			}
		} else {
			err = json.Unmarshal([]byte(unquoted), response)
		}

		if err != nil {
			pc.logger.WithError(err).Fatalf("Protecode scan failed, error during decode response: %v", newStr)
		}
	}

	return response
}

func (pc *Protecode) getProductData(r io.ReadCloser) *ProductData {
	defer r.Close()

	response := new(ProductData)

	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	newStr := buf.String()

	if len(newStr) > 0 {

		unquoted, err := strconv.Unquote(newStr)
		if err != nil {
			err = json.Unmarshal([]byte(newStr), response)
			if err != nil {
				pc.logger.WithError(err).Fatalf("Protecode scan failed, error during unqote response: %v", newStr)
			}
		} else {
			err = json.Unmarshal([]byte(unquoted), response)
		}

		if err != nil {
			pc.logger.WithError(err).Fatalf("Protecode scan failed, error during decode response: %v", newStr)
		}
	}
	return response
}

func (pc *Protecode) sendAPIRequest(method string, url string, headers map[string][]string) (*io.ReadCloser, error) {

	r, err := pc.client.SendRequest(method, url, nil, headers, nil)

	return &r.Body, err
}

// ParseResultForInflux parses the result from the scan into the internal format
func (pc *Protecode) ParseResultForInflux(result Result, protecodeExcludeCVEs string) map[string]int {
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
			if vulnerability.Exact {
				if pc.isExcluded(vulnerability, protecodeExcludeCVEs) {
					m["excluded_vulnerabilities"]++
				} else if pc.isTriaged(vulnerability) {
					m["triaged_vulnerabilities"]++
				} else {
					m["count"]++
					if pc.isSevereCVSS3(vulnerability) {
						m["cvss3GreaterOrEqualSeven"]++
						m["major_vulnerabilities"]++
					} else if pc.isSevereCVSS2(vulnerability) {
						m["cvss2GreaterOrEqualSeven"]++
						m["major_vulnerabilities"]++
					} else {
						m["minor_vulnerabilities"]++
					}
					m["vulnerabilities"]++
				}
			} else {
				m["historical_vulnerabilities"]++
			}
		}
	}

	return m
}

func (pc *Protecode) isExcluded(vulnerability Vulnerability, protecodeExcludeCVEs string) bool {
	return strings.Contains(protecodeExcludeCVEs, vulnerability.Vuln.Cve)
}

func (pc *Protecode) isTriaged(vulnerability Vulnerability) bool {
	return len(vulnerability.Triage) > 0
}

func (pc *Protecode) isSevereCVSS3(vulnerability Vulnerability) bool {
	threshold := 7.0
	cvss3, _ := strconv.ParseFloat(vulnerability.Vuln.Cvss3Score, 64)
	return cvss3 >= threshold
}

func (pc *Protecode) isSevereCVSS2(vulnerability Vulnerability) bool {
	threshold := 7.0
	cvss3, _ := strconv.ParseFloat(vulnerability.Vuln.Cvss3Score, 64)
	return cvss3 == 0 && vulnerability.Vuln.Cvss >= threshold
}

// DeleteScan deletes if configured the scan on the protecode server
func (pc *Protecode) DeleteScan(cleanupMode string, productID int) {

	switch cleanupMode {
	case "none":
	case "binary":
		return
	case "complete":
		pc.logger.Info("Protecode scan successful. Deleting scan from server.")
		protecodeURL := pc.createURL("/api/product/", fmt.Sprintf("%v/", productID), "")
		headers := map[string][]string{}

		pc.sendAPIRequest("DELETE", protecodeURL, headers)
		break
	default:
		pc.logger.Fatalf("Protecode scan failed, unknown cleanup mode %v", cleanupMode)
	}

}

// LoadReport loads the report of the protecode scan
func (pc *Protecode) LoadReport(reportFileName string, productID int) *io.ReadCloser {

	protecodeURL := pc.createURL("/api/product/", fmt.Sprintf("%v/pdf-report", productID), "")
	headers := map[string][]string{
		"Cache-Control": []string{"no-cache, no-store, must-revalidate"},
		"Pragma":        []string{"no-cache"},
		"Outputfile":    []string{reportFileName},
	}

	readCloser, err := pc.sendAPIRequest(http.MethodGet, protecodeURL, headers)
	if err != nil {
		pc.logger.WithError(err).Fatalf("Protecode scan failed, not possible to load report %v", protecodeURL)
	}

	return readCloser
}

// UploadScanFile upload the scan file to the protecode server
func (pc *Protecode) UploadScanFile(cleanupMode, protecodeGroup, filePath string, fileName string) *ResultData {
	deleteBinary := (cleanupMode == "binary" || cleanupMode == "complete")
	headers := map[string][]string{"Group": []string{protecodeGroup}, "Delete-Binary": []string{fmt.Sprintf("%v", deleteBinary)}}

	url := fmt.Sprintf("%v/api/upload/%v", pc.serverURL, fileName)
	r, err := pc.client.UploadRequest(http.MethodPut, url, filePath, "file", headers, nil)
	if err != nil {
		pc.logger.WithError(err).Fatalf("Protecode scan failed, error during %v upload request", url)
	}

	return pc.getResultData(r.Body)
}

// DeclareFetchURL configures the fetch url for the protecode scan
func (pc *Protecode) DeclareFetchURL(cleanupMode, protecodeGroup, fetchURL string) *Result {
	deleteBinary := (cleanupMode == "binary" || cleanupMode == "complete")
	headers := map[string][]string{"Group": []string{protecodeGroup}, "Delete-Binary": []string{fmt.Sprintf("%v", deleteBinary)}, "Url": []string{fetchURL}, "Content-Type": []string{"application/json"}}

	protecodeURL := fmt.Sprintf("%v/api/fetch/", pc.serverURL)
	r, err := pc.sendAPIRequest(http.MethodPost, protecodeURL, headers)
	if err != nil {
		pc.logger.WithError(err).Fatalf("Protecode scan failed, exception during declare fetch url: %v", protecodeURL)
	}
	return pc.getResult(*r)
}

//PollForResult polls the protecode scan for the result scan
func (pc *Protecode) PollForResult(productID int, timeOutInMinutes string, verbose bool) ResultData {

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
		if len(response.Result.Components) > 0 && response.Result.Status != "B" {
			ticker.Stop()
			i = 0
			break
		}

		select {
		case t := <-ticker.C:
			if verbose {
				pc.logger.Infof("Tick : %v Processing status for productID %v", t, productID)
			}
			response, err = pc.pullResult(productID)
			if err != nil {
				ticker.Stop()
				i = 0
				return response
			}
			if len(response.Result.Components) > 0 && response.Result.Status != "B" {
				ticker.Stop()
				i = 0
				break
			}
		}
	}

	if len(response.Result.Components) == 0 || response.Result.Status == "B" {
		response, err = pc.pullResult(productID)
		if err != nil || len(response.Result.Components) == 0 || response.Result.Status == "B" {
			pc.logger.Fatal("Protecode scan failed, no result after polling")
		}
	}

	return response
}

func (pc *Protecode) pullResult(productID int) (ResultData, error) {
	protecodeURL, headers := pc.getPullResultRequestData(productID)

	r, err := pc.sendAPIRequest(http.MethodGet, protecodeURL, headers)
	if err != nil {
		return *new(ResultData), err
	}
	response := pc.getResultData(*r)

	return *response, nil
}

func (pc *Protecode) getPullResultRequestData(productID int) (string, map[string][]string) {
	protecodeURL := pc.createURL("/api/product/", fmt.Sprintf("%v/", productID), "")
	headers := map[string][]string{
		"acceptType": []string{"application/json"},
	}

	return protecodeURL, headers
}

// LoadExistingProduct loads the existing product from protecode service
func (pc *Protecode) LoadExistingProduct(protecodeGroup string, reuseExisting bool) int {
	var productID int = -1

	if reuseExisting {

		protecodeURL := pc.createURL("/api/apps/", fmt.Sprintf("%v/", protecodeGroup), "")
		headers := map[string][]string{
			"acceptType": []string{"application/json"},
		}

		response := pc.loadExisting(protecodeURL, headers)
		// by definition we will take the first one and trigger rescan
		productID = response.Products[0].ProductID

		pc.logger.Infof("re-use existing Protecode scan - group: %v, productID: %v", protecodeGroup, productID)
	}

	return productID
}

func (pc *Protecode) loadExisting(protecodeURL string, headers map[string][]string) *ProductData {

	r, err := pc.sendAPIRequest(http.MethodGet, protecodeURL, headers)
	if err != nil {
		pc.logger.WithError(err).Fatalf("Protecode scan failed, during load existing product: %v", protecodeURL)
	}

	return pc.getProductData(*r)
}
