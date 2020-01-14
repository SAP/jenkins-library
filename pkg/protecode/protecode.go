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
	ReportUrl  string      `json:"report_url`
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
	client    piperHttp.Client
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

func (pc *Protecode) SetOptions(options ProtecodeOptions) {
	pc.serverURL = options.ServerURL
	pc.client = piperHttp.Client{}
	pc.duration = options.Duration

	if options.Logger != nil {
		pc.logger = options.Logger
	} else {
		pc.logger = log.Entry().WithField("package", "SAP/jenkins-library/pkg/protecode")
	}

	httpOptions := piperHttp.ClientOptions{options.Duration, options.Username, options.Password, "", options.Logger}
	pc.client.SetOptions(httpOptions)
}

func (pc *Protecode) createUrl(path string, pValue string, fParam string) (string, error) {

	protecodeUrl, err := url.Parse(pc.serverURL)
	if err != nil {
		pc.logger.WithError(err).Fatal("Malformed URL")
		return "", err
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

	return protecodeUrl.String(), nil
}

func (pc *Protecode) getResultData(r io.ReadCloser) (*ResultData, error) {
	defer r.Close()

	response := new(ResultData)

	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	newStr := buf.String()

	if len(newStr) > 0 {
		err := json.Unmarshal([]byte(newStr), response)

		if err != nil {
			pc.logger.WithError(err).Fatalf("error during decode response: %v", r)
			return response, err
		}
	}

	return response, nil
}

func (pc *Protecode) getResult(r io.ReadCloser) (*Result, error) {
	defer r.Close()

	response := new(Result)

	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	newStr := buf.String()

	if len(newStr) > 0 {
		err := json.Unmarshal([]byte(newStr), response)

		if err != nil {
			pc.logger.WithError(err).Fatalf("error during decode response: %v", r)
			return response, err
		}
	}

	return response, nil
}

func (pc *Protecode) getProductData(r io.ReadCloser) (*ProductData, error) {
	defer r.Close()

	response := new(ProductData)

	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	newStr := buf.String()
	if len(newStr) > 0 {
		err := json.Unmarshal([]byte(newStr), response)

		if err != nil {
			pc.logger.WithError(err).Fatalf("error during decode response: %v", r)
			return response, err
		}
	}

	return response, nil
}

func (pc *Protecode) uploadFileRequest(url, filePath string, headers map[string][]string) (*io.ReadCloser, error) {
	r, err := pc.client.UploadFile(url, filePath, "file", headers, nil)
	if err != nil {
		pc.logger.WithError(err).Fatalf("error during %v upload reuqest", url)
		return &r.Body, err
	}

	return &r.Body, nil
}

func (pc *Protecode) sendApiRequest(method string, url string, headers map[string][]string) (*io.ReadCloser, error) {

	r, err := pc.client.SendRequest(method, url, nil, headers, nil)
	if err != nil {
		pc.logger.WithError(err).Fatalf("error during %v: %v request", method, url)
		return nil, err
	}

	return &r.Body, nil
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

func (pc *Protecode) DeleteScan(cleanupMode string, productId int) error {

	switch cleanupMode {
	case "none":
	case "binary":
		return nil
	case "complete":
		pc.logger.Info("Protecode scan successful. Deleting scan from server.")
		protecodeURL, err := pc.createUrl("/api/product/", fmt.Sprintf("%v/", productId), "")
		if err != nil {
			return err
		}
		headers := map[string][]string{}

		_, err = pc.sendApiRequest("DELETE", protecodeURL, headers)
		if err != nil {
			return err
		}
		break
	default:
		pc.logger.Fatalf("Unknown cleanup mode %v", cleanupMode)
	}

	return nil
}

// #####################################
// LoadReport

func (pc *Protecode) LoadReport(reportFileName string, productId int) (*io.ReadCloser, error) {

	protecodeURL, err := pc.createUrl("/api/product/", fmt.Sprintf("%v/pdf-report", productId), "")
	if err != nil {
		return nil, err
	}
	headers := map[string][]string{
		"Cache-Control": []string{"no-cache, no-store, must-revalidate"},
		"Pragma":        []string{"no-cache"},
		"Outputfile":    []string{reportFileName},
	}

	return pc.sendApiRequest(http.MethodGet, protecodeURL, headers)
}

// #####################################
// UploadScanFile

func (pc *Protecode) UploadScanFile(cleanupMode, protecodeGroup, filePath string) (*Result, error) {
	deleteBinary := (cleanupMode == "binary" || cleanupMode == "complete")
	headers := map[string][]string{"Group": []string{protecodeGroup}, "Delete-Binary": []string{fmt.Sprintf("%v", deleteBinary)}}

	r, err := pc.uploadFileRequest(fmt.Sprintf("%v/api/upload/", pc.serverURL), filePath, headers)
	if err != nil {
		pc.logger.WithError(err).Fatalf("error during %v upload request", fmt.Sprintf("%v/api/fetch/", pc.serverURL))
		return new(Result), err
	}
	return pc.getResult(*r)
}

// #####################################
// declareFetchUrl

func (pc *Protecode) DeclareFetchUrl(cleanupMode, protecodeGroup, fetchURL string) (*Result, error) {
	deleteBinary := (cleanupMode == "binary" || cleanupMode == "complete")
	headers := map[string][]string{"Group": []string{protecodeGroup}, "Delete-Binary": []string{fmt.Sprintf("%v", deleteBinary)}, "Url": []string{fetchURL}, "Content-Type": []string{"application/json"}}

	r, err := pc.sendApiRequest(http.MethodPost, fmt.Sprintf("%v/api/fetch/", pc.serverURL), headers)
	if err != nil {
		pc.logger.WithError(err).Fatalf("error during POST: %v request", fmt.Sprintf("%v/api/fetch/", pc.serverURL))
		return new(Result), err
	}
	return pc.getResult(*r)
}

// #####################################
// Pull result

func (pc *Protecode) PollForResult(productId int, verbose bool) (Result, error) {

	var response Result
	var err error

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	ticks := pc.duration / 10

	for i := ticks; i > 0; i-- {

		response, err = pc.pullResult(productId)
		if err != nil {
			ticker.Stop()
			i = 0
			return response, err
		}
		if len(response.Components) > 0 && response.Status != "B" {
			ticker.Stop()
			i = 0
			break
		}

		select {
		case t := <-ticker.C:
			fmt.Printf("Ticker %v", t)
			if verbose {
				fmt.Printf("Processing status for productId %v", productId)
			}
			response, err = pc.pullResult(productId)
			if err != nil {
				ticker.Stop()
				i = 0
				return response, err
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
			return response, err
		}
	}

	return response, nil
}

// #####################################
// Pull result

func (pc *Protecode) pullResult(productId int) (Result, error) {
	protecodeURL, headers, err := pc.getPullResultRequestData(productId)
	if err != nil {
		return *new(Result), err
	}

	return pc.pullResultData(protecodeURL, headers)

}

func (pc *Protecode) pullResultData(protecodeURL string, headers map[string][]string) (Result, error) {
	r, err := pc.sendApiRequest(http.MethodGet, protecodeURL, headers)

	response, err := pc.getResultData(*r)

	return response.Result, err
}

func (pc *Protecode) getPullResultRequestData(productId int) (string, map[string][]string, error) {
	protecodeURL, err := pc.createUrl("/api/product/", fmt.Sprintf("%v/", productId), "")
	headers := map[string][]string{
		"acceptType": []string{"APPLICATION_JSON"},
	}

	return protecodeURL, headers, err
}

// #####################################
// Load existing product
func (pc *Protecode) LoadExistingProduct(protecodeGroup, filePath string, reuseExisting bool) (int, error) {
	var productId int = -1

	if reuseExisting {

		response, err := pc.loadExistingProductByFilename(protecodeGroup, filePath)
		if err != nil {
			return 0, err
		}
		// by definition we will take the first one and trigger rescan
		productId = response.Products[0].ProductId

		fmt.Printf("re-use existing Protecode scan - file: %v, group: %v, productId: %v", filePath, protecodeGroup, productId)
	}

	return productId, nil
}

func (pc *Protecode) loadExistingProductByFilename(protecodeGroup, filePath string) (*ProductData, error) {

	protecodeURL, headers, err := pc.getLoadExistiongProductRequestData(protecodeGroup, filePath)

	if err != nil {
		return new(ProductData), err
	}

	return pc.loadExisting(protecodeURL, headers)
}

func (pc *Protecode) getLoadExistiongProductRequestData(protecodeGroup, filePath string) (string, map[string][]string, error) {

	protecodeURL, err := pc.createUrl("/api/apps/", fmt.Sprintf("%v/", protecodeGroup), filePath)
	headers := map[string][]string{
		//TODO change to mimetype
		"acceptType": []string{"APPLICATION_JSON"},
	}

	return protecodeURL, headers, err
}

func (pc *Protecode) loadExisting(protecodeURL string, headers map[string][]string) (*ProductData, error) {

	r, err := pc.sendApiRequest(http.MethodGet, protecodeURL, headers)
	if err != nil {
		return new(ProductData), err
	}

	return pc.getProductData(*r)
}
