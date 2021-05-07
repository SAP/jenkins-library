package protecode

import (
	"fmt"
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

// LoadExistingProduct loads the existing product from protecode service
func (pc *Protecode) LoadExistingProduct(group string, verifyOnly bool) int {
	var productID int = -1

	if verifyOnly {

		response := pc.loadProductData(group)
		// by definition we will take the first one and trigger rescan
		productID = response.Products[0].ProductID

		pc.logger.Infof("Re-use existing Protecode scan - group: %v, productID: %v", group, productID)
	}

	return productID
}
