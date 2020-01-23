package protecode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/sirupsen/logrus"
)

const DELIMITER = "-DeLiMiTeR-"

type ProductData struct {
	Products []Product `json:"products"`
}

type Product struct {
	ProductId int `json:"product_id"`
}

type ResultData struct {
	Result Result `json:"results"`
}

type Result struct {
	ProductId  int         `json:"product_id"`
	ReportUrl  string      `json:"report_url"`
	Status     string      `json:"status"`
	Components []Component `json:"components,omitempty"`
}

type Component struct {
	Vulns []Vulnerability `json:"vulns,omitempty"`
}

type Vulnerability struct {
	Exact  bool     `json:"exact"`
	Vuln   Vuln     `json:"vuln"`
	Triage []Triage `json:"triage,omitempty"`
}

type Vuln struct {
	Cve        string  `json:"cve"`
	Cvss       float64 `json:"cvss"`
	Cvss3Score string  `json:"cvss3_score"`
}

type Triage struct {
	Id          int    `json:"id"`
	VulnId      string `json:"vuln_id"`
	Component   string `json:"component"`
	Vendor      string `json:"vendor"`
	Codetype    string `json:"codetype"`
	Version     string `json:"version"`
	Modified    string `json:"modified"`
	Scope       string `json:"scope"`
	Description string `json:"description"`
	User        User   `json:"user"`
}

type User struct {
	Id        int    `json:"id"`
	Email     string `json:"email"`
	Girstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Username  string `json:"username"`
}

type Protecode struct {
	serverURL string
	client    *piperHttp.Client
	duration  time.Duration
	logger    *logrus.Entry
}

type ProtecodeOptions struct {
	ServerURL string
	Duration  time.Duration
	Username  string
	Password  string
	Logger    *logrus.Entry
}

type Wrapper struct {
	Data string
}

func (pc *Protecode) SetOptions(options ProtecodeOptions) {
	pc.serverURL = options.ServerURL
	pc.client = &piperHttp.Client{}
	pc.duration = (time.Minute * options.Duration)

	if options.Logger != nil {
		pc.logger = options.Logger
	} else {
		pc.logger = log.Entry().WithField("package", "SAP/jenkins-library/pkg/protecode")
	}

	httpOptions := piperHttp.ClientOptions{(time.Minute * options.Duration), options.Username, options.Password, "", options.Logger}
	pc.client.SetOptions(httpOptions)
}

func (pc *Protecode) createUrl(path string, pValue string, fParam string) string {

	protecodeUrl, err := url.Parse(pc.serverURL)
	if err != nil {
		pc.logger.WithError(err).Fatal("Malformed URL")
	}

	if len(path) > 0 {
		protecodeUrl.Path += fmt.Sprintf("%v", path)
	}

	if len(pValue) > 0 {
		protecodeUrl.Path += fmt.Sprintf("%v", pValue)
	}

	// Prepare Query Parameters
	if len(fParam) > 0 {
		encodedFParam := url.QueryEscape(fParam)
		params := url.Values{}
		params.Add("q", fmt.Sprintf("file:%v", encodedFParam))

		// Add Query Parameters to the URL
		protecodeUrl.RawQuery = params.Encode() // Escape Query Parameters
	}

	return protecodeUrl.String()
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
				pc.logger.WithError(err).Fatalf("error during unqote response: %v", newStr)
			}
		} else {
			err = json.Unmarshal([]byte(unquoted), response)
		}

		if err != nil {
			pc.logger.WithError(err).Fatalf("error during decode response: %v", newStr)
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
				pc.logger.WithError(err).Fatalf("error during unqote response: %v", newStr)
			}
		} else {
			err = json.Unmarshal([]byte(unquoted), response)
		}

		if err != nil {
			pc.logger.WithError(err).Fatalf("error during decode response: %v", newStr)
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
				pc.logger.WithError(err).Fatalf("error during unqote response: %v", newStr)
			}
		} else {
			err = json.Unmarshal([]byte(unquoted), response)
		}

		if err != nil {
			pc.logger.WithError(err).Fatalf("error during decode response: %v", newStr)
		}
	}
	return response
}

func (pc *Protecode) uploadFileRequest(url, filePath string, headers map[string][]string) *io.ReadCloser {
	pc.logger.Debugf("Upload %v %v %v", url, filePath, headers)
	r, err := pc.client.UploadRequest(http.MethodPut, url, filePath, "file", headers, nil)
	if err != nil {
		pc.logger.WithError(err).Fatalf("error during %v upload request", url)
	}

	return &r.Body
}

func (pc *Protecode) sendApiRequest(method string, url string, headers map[string][]string) (*io.ReadCloser, error) {

	r, err := pc.client.SendRequest(method, url, nil, headers, nil)

	return &r.Body, err
}

func (pc *Protecode) ResolveSymLink(method string, url string) (*io.ReadCloser, error) {

	link, err := os.Readlink(url)
	if err != nil {
		pc.logger.WithError(err).Fatalf("error during %v resolve symlink", url)
	}
	r, err := pc.sendApiRequest("GET", link, nil)

	return r, err
}

// #####################################
// ParseResultForInflux

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

// #####################################
// DeleteScan

func (pc *Protecode) DeleteScan(cleanupMode string, productId int) {

	switch cleanupMode {
	case "none":
	case "binary":
		return
	case "complete":
		pc.logger.Info("Protecode scan successful. Deleting scan from server.")
		protecodeURL := pc.createUrl("/api/product/", fmt.Sprintf("%v/", productId), "")
		headers := map[string][]string{}

		pc.sendApiRequest("DELETE", protecodeURL, headers)
		break
	default:
		pc.logger.Fatalf("Unknown cleanup mode %v", cleanupMode)
	}

}

// #####################################
// LoadReport

func (pc *Protecode) LoadReport(reportFileName string, productId int) *io.ReadCloser {

	protecodeURL := pc.createUrl("/api/product/", fmt.Sprintf("%v/pdf-report", productId), "")
	headers := map[string][]string{
		"Cache-Control": []string{"no-cache, no-store, must-revalidate"},
		"Pragma":        []string{"no-cache"},
		"Outputfile":    []string{reportFileName},
	}

	readCloser, err := pc.sendApiRequest(http.MethodGet, protecodeURL, headers)
	if err != nil {
		pc.logger.WithError(err).Fatalf("Load Report failed %v", protecodeURL)
	}

	return readCloser
}

// #####################################
// UploadScanFile

func (pc *Protecode) UploadScanFile(cleanupMode, protecodeGroup, filePath string, fileName string) *ResultData {
	deleteBinary := (cleanupMode == "binary" || cleanupMode == "complete")
	headers := map[string][]string{"Group": []string{protecodeGroup}, "Delete-Binary": []string{fmt.Sprintf("%v", deleteBinary)}}

	r := pc.uploadFileRequest(fmt.Sprintf("%v/api/upload/%v", pc.serverURL, fileName), filePath, headers)
	return pc.getResultData(*r)
}

// #####################################
// declareFetchUrl

func (pc *Protecode) DeclareFetchUrl(cleanupMode, protecodeGroup, fetchURL string) *Result {
	deleteBinary := (cleanupMode == "binary" || cleanupMode == "complete")
	headers := map[string][]string{"Group": []string{protecodeGroup}, "Delete-Binary": []string{fmt.Sprintf("%v", deleteBinary)}, "Url": []string{fetchURL}, "Content-Type": []string{"application/json"}}

	protecodeURL := fmt.Sprintf("%v/api/fetch/", pc.serverURL)
	r, err := pc.sendApiRequest(http.MethodPost, protecodeURL, headers)
	if err != nil {
		pc.logger.WithError(err).Fatalf("Protecode scan exception during declare fetch url: %v", protecodeURL)
	}
	return pc.getResult(*r)
}

// #####################################
// Pull result

func (pc *Protecode) PollForResult(productId int, verbose bool) Result {

	var response Result
	var err error

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	ticks := pc.duration / time.Second / 10
	pc.logger.Infof("Poll for result %v times", ticks)

	for i := ticks; i > 0; i-- {

		response, err = pc.pullResult(productId)
		if err != nil {
			ticker.Stop()
			i = 0
			return response
		}
		if len(response.Components) > 0 && response.Status != "B" {
			ticker.Stop()
			i = 0
			break
		}

		select {
		case t := <-ticker.C:
			if verbose {
				pc.logger.Infof("Tick : %v Processing status for productId %v", t, productId)
			}
			response, err = pc.pullResult(productId)
			if err != nil {
				ticker.Stop()
				i = 0
				return response
			}
			if len(response.Components) > 0 && response.Status != "B" {
				ticker.Stop()
				i = 0
				break
			}
		}
	}

	if len(response.Components) == 0 && response.Status == "B" {
		response, err = pc.pullResult(productId)
		if err != nil || len(response.Components) == 0 || response.Status == "B" {
			pc.logger.Fatal("No result for protecode scan")
		}
	}

	return response
}

// #####################################
// Pull result

func (pc *Protecode) pullResult(productId int) (Result, error) {
	protecodeURL, headers := pc.getPullResultRequestData(productId)

	return pc.pullResultData(protecodeURL, headers)

}

func (pc *Protecode) pullResultData(protecodeURL string, headers map[string][]string) (Result, error) {
	r, err := pc.sendApiRequest(http.MethodGet, protecodeURL, headers)
	if err != nil {
		return *new(Result), err
	}
	response := pc.getResultData(*r)

	return response.Result, nil
}

func (pc *Protecode) getPullResultRequestData(productId int) (string, map[string][]string) {
	protecodeURL := pc.createUrl("/api/product/", fmt.Sprintf("%v/", productId), "")
	headers := map[string][]string{
		"acceptType": []string{"APPLICATION_JSON"},
	}

	return protecodeURL, headers
}

// #####################################
// Load existing product
func (pc *Protecode) LoadExistingProduct(protecodeGroup, filePath string, reuseExisting bool) int {
	var productId int = -1

	if reuseExisting {

		response := pc.loadExistingProductByFilename(protecodeGroup, filePath)
		// by definition we will take the first one and trigger rescan
		productId = response.Products[0].ProductId

		pc.logger.Infof("re-use existing Protecode scan - file: %v, group: %v, productId: %v", filePath, protecodeGroup, productId)
	}

	return productId
}

func (pc *Protecode) loadExistingProductByFilename(protecodeGroup, filePath string) *ProductData {

	protecodeURL, headers := pc.getLoadExistiongProductRequestData(protecodeGroup, filePath)

	return pc.loadExisting(protecodeURL, headers)
}

func (pc *Protecode) getLoadExistiongProductRequestData(protecodeGroup, filePath string) (string, map[string][]string) {

	protecodeURL := pc.createUrl("/api/apps/", fmt.Sprintf("%v/", protecodeGroup), filePath)
	headers := map[string][]string{
		//TODO change to mimetype
		"acceptType": []string{"APPLICATION_JSON"},
	}

	return protecodeURL, headers
}

func (pc *Protecode) loadExisting(protecodeURL string, headers map[string][]string) *ProductData {

	r, err := pc.sendApiRequest(http.MethodGet, protecodeURL, headers)
	if err != nil {
		pc.logger.WithError(err).Fatalf("Protecode load existing product failed: %v", protecodeURL)
	}

	return pc.getProductData(*r)
}
