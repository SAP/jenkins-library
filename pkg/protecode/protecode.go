package protecode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/SAP/jenkins-library/pkg/command"
	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
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

func CreateUrl(pURL string, path string, pValue string, fParam string) (string, error) {

	protecodeUrl, err := url.Parse(pURL)
	if err != nil {
		log.Entry().WithError(err).Fatal("Malformed URL")
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

func GetResultData(r io.ReadCloser) (*ResultData, error) {
	defer r.Close()

	response := new(ResultData)

	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	newStr := buf.String()

	if len(newStr) > 0 {
		err := json.Unmarshal([]byte(newStr), response)

		if err != nil {
			log.Entry().WithError(err).Fatalf("error during decode response: %v", r)
			return response, err
		}
	}

	return response, nil
}

func GetResult(r io.ReadCloser) (*Result, error) {
	defer r.Close()

	response := new(Result)

	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	newStr := buf.String()

	if len(newStr) > 0 {
		err := json.Unmarshal([]byte(newStr), response)

		if err != nil {
			log.Entry().WithError(err).Fatalf("error during decode response: %v", r)
			return response, err
		}
	}

	return response, nil
}

func GetProductData(r io.ReadCloser) (*ProductData, error) {
	defer r.Close()

	response := new(ProductData)

	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	newStr := buf.String()
	if len(newStr) > 0 {
		err := json.Unmarshal([]byte(newStr), response)

		if err != nil {
			log.Entry().WithError(err).Fatalf("error during decode response: %v", r)
			return response, err
		}
	}

	return response, nil
}

func UploadScanFile(url, filePath string, headers map[string][]string, client piperHttp.Client) (*io.ReadCloser, error) {
	r, err := client.UploadFile(url, filePath, "", headers, nil)
	if err != nil {
		log.Entry().WithError(err).Fatalf("error during %v: %v reuqest", method, url)
		return nil, err
	}

	return &r.Body, nil
}

func SendApiRequest(method string, url string, headers map[string][]string, client piperHttp.Client) (*io.ReadCloser, error) {

	r, err := client.SendRequest(method, url, nil, headers, nil)
	if err != nil {
		log.Entry().WithError(err).Fatalf("error during %v: %v reuqest", method, url)
		return nil, err
	}

	return &r.Body, nil
}

func ParseResultForInflux(result Result, protecodeExcludeCVEs string) map[string]int {
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
				if isExcluded(vulnerability, protecodeExcludeCVEs) {
					m["excluded_vulnerabilities"]++
				} else if isTriaged(vulnerability) {
					m["triaged_vulnerabilities"]++
				} else {
					m["count"]++
					if isSevereCVSS3(vulnerability) {
						m["cvss3GreaterOrEqualSeven"]++
						m["major_vulnerabilities"]++
					} else if isSevereCVSS2(vulnerability) {
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

func isExcluded(vulnerability Vulnerability, protecodeExcludeCVEs string) bool {
	return strings.Contains(protecodeExcludeCVEs, vulnerability.Vuln.Cve)
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

func WriteResultAsJSONToFile(m map[string]int, filename string, writeFunc func(f string, b []byte, p os.FileMode) error) error {
	b, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return writeFunc(filename, b, 644)
}
