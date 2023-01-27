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

	"github.com/sirupsen/logrus"

	piperHttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
)

// ReportsDirectory defines the subfolder for the Protecode reports which are generated
const ReportsDirectory = "protecode"

// ProductData holds the product information of the protecode product
type ProductData struct {
	Products []Product `json:"products,omitempty"`
}

// Product holds the id of the protecode product
type Product struct {
	ProductID int    `json:"product_id,omitempty"`
	FileName  string `json:"name,omitempty"`
}

// ResultData holds the information about the protecode result
type ResultData struct {
	Result Result `json:"results,omitempty"`
}

// Result holds the detail information about the protecode result
type Result struct {
	ProductID  int         `json:"product_id,omitempty"`
	ReportURL  string      `json:"report_url,omitempty"`
	Status     string      `json:"status,omitempty"`
	Components []Component `json:"components,omitempty"`
}

// Component the protecode component information
type Component struct {
	Vulns []Vulnerability `json:"vulns,omitempty"`
}

// Vulnerability the protecode vulnerability information
type Vulnerability struct {
	Exact  bool     `json:"exact,omitempty"`
	Vuln   Vuln     `json:"vuln,omitempty"`
	Triage []Triage `json:"triage,omitempty"`
}

// Vuln holds the information about the vulnerability
type Vuln struct {
	Cve        string `json:"cve,omitempty"`
	Cvss       string `json:"cvss,omitempty"`
	Cvss3Score string `json:"cvss3_score,omitempty"`
}

// Triage holds the triaging information
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

// User holds the user information
type User struct {
	ID        int    `json:"id,omitempty"`
	Email     string `json:"email,omitempty"`
	Firstname string `json:"firstname,omitempty"`
	Lastname  string `json:"lastname,omitempty"`
	Username  string `json:"username,omitempty"`
}

// Protecode ist the protecode client which is used by the step
type Protecode struct {
	serverURL string
	client    piperHttp.Uploader
	duration  time.Duration
	logger    *logrus.Entry
}

// Used to reduce wait time during tests
var protecodePollInterval = 10 * time.Second

// Just calls SetOptions which makes sure logger is set.
// Added to make test code more resilient
func makeProtecode(opts Options) Protecode {
	ret := Protecode{}
	ret.SetOptions(opts)
	return ret
}

// Options struct which can be used to configure the Protecode struct
type Options struct {
	ServerURL  string
	Duration   time.Duration
	Username   string
	Password   string
	UserAPIKey string
	Logger     *logrus.Entry
}

// SetOptions setter function to set the internal properties of the protecode
func (pc *Protecode) SetOptions(options Options) {
	pc.serverURL = options.ServerURL
	pc.client = &piperHttp.Client{}
	pc.duration = options.Duration

	if options.Logger != nil {
		pc.logger = options.Logger
	} else {
		pc.logger = log.Entry().WithField("package", "SAP/jenkins-library/pkg/protecode")
	}

	httpOptions := piperHttp.ClientOptions{MaxRequestDuration: options.Duration, Logger: options.Logger}

	// If userAPIKey is not empty then we will use it for user authentication, instead of username & password
	if options.UserAPIKey != "" {
		httpOptions.Token = "Bearer " + options.UserAPIKey
	} else {
		httpOptions.Username = options.Username
		httpOptions.Password = options.Password
	}
	pc.client.SetOptions(httpOptions)
}

// SetHttpClient setter function to set the http client
func (pc *Protecode) SetHttpClient(client piperHttp.Uploader) {
	pc.client = client
}

func (pc *Protecode) createURL(path string, pValue string, fParam string) string {

	protecodeURL, err := url.Parse(pc.serverURL)
	if err != nil {
		//TODO: bubble up error
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
		// encodedFParam := url.QueryEscape(fParam)
		params := url.Values{}
		params.Add("q", fmt.Sprintf("file:%v", fParam))

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
				//TODO: bubble up error
				pc.logger.WithError(err).Fatalf("Error during unqote response: %v", newStr)
			}
		} else {
			err = json.Unmarshal([]byte(unquoted), response)
		}

		if err != nil {
			//TODO: bubble up error
			pc.logger.WithError(err).Fatalf("Error during decode response: %v", newStr)
		}
	}
}

func (pc *Protecode) sendAPIRequest(method string, url string, headers map[string][]string) (*io.ReadCloser, int, error) {

	r, err := pc.client.SendRequest(method, url, nil, headers, nil)
	if err != nil {
		if r != nil {
			return nil, r.StatusCode, err
		}
		return nil, 400, err
	}

	//return &r.Body, nil
	return &r.Body, r.StatusCode, nil
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
	parsedCvss, _ := strconv.ParseFloat(vulnerability.Vuln.Cvss, 64)
	return cvss3 == 0 && parsedCvss >= threshold
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
		//TODO: bubble up error
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

	readCloser, _, err := pc.sendAPIRequest(http.MethodGet, protecodeURL, headers)
	if err != nil {
		//TODO: bubble up error
		pc.logger.WithError(err).Fatalf("It is not possible to load report %v", protecodeURL)
	}

	return readCloser
}

// UploadScanFile upload the scan file to the protecode server
func (pc *Protecode) UploadScanFile(cleanupMode, group, customDataJSONMap, filePath, fileName, version string, productID int, replaceBinary bool) *ResultData {
	log.Entry().Debugf("[DEBUG] ===> UploadScanFile started.....")

	deleteBinary := (cleanupMode == "binary" || cleanupMode == "complete")

	var headers = make(map[string][]string)
	if len(customDataJSONMap) > 0 {
		customDataHeaders := map[string]string{}
		if err := json.Unmarshal([]byte(customDataJSONMap), &customDataHeaders); err != nil {
			log.Entry().Warn("[WARN] ===> customDataJSONMap flag must be a valid JSON map. Check the value of --customDataJSONMap and try again.")
		} else {
			for k, v := range customDataHeaders {
				headers["META-"+strings.ToUpper(k)] = []string{v}
			}
		}
	}

	headers["Group"] = []string{group}
	headers["Delete-Binary"] = []string{fmt.Sprintf("%v", deleteBinary)}

	if (replaceBinary) && (version != "") {
		log.Entry().Debugf("[DEBUG] ===> replaceBinary && version != empty ")
		headers["Replace"] = []string{fmt.Sprintf("%v", productID)}
		headers["Version"] = []string{version}
	} else if replaceBinary {
		headers["Replace"] = []string{fmt.Sprintf("%v", productID)}
		log.Entry().Debugf("[DEBUG] ===> replaceBinary")
	} else if version != "" {
		log.Entry().Debugf("[DEBUG] ===> version != empty ")
		headers["Version"] = []string{version}
	}

	uploadURL := fmt.Sprintf("%v/api/upload/%v", pc.serverURL, fileName)

	r, err := pc.client.UploadRequest(http.MethodPut, uploadURL, filePath, "file", headers, nil, "binary")
	if err != nil {
		//TODO: bubble up error
		pc.logger.WithError(err).Fatalf("Error during upload request %v", uploadURL)
	} else {
		pc.logger.Info("Upload successful")
	}

	// For replaceBinary option response doesn't contain any result but just a message saying that product successfully replaced.
	if replaceBinary && r.StatusCode == 201 {
		result := new(ResultData)
		result.Result.ProductID = productID
		return result

	} else {
		result := new(ResultData)
		pc.mapResponse(r.Body, result)
		return result

	}

	//return result
}

// DeclareFetchURL configures the fetch url for the protecode scan
func (pc *Protecode) DeclareFetchURL(cleanupMode, group, customDataJSONMap, fetchURL, version string, productID int, replaceBinary bool) *ResultData {
	deleteBinary := (cleanupMode == "binary" || cleanupMode == "complete")

	var headers = make(map[string][]string)
	if len(customDataJSONMap) > 0 {
		customDataHeaders := map[string]string{}
		if err := json.Unmarshal([]byte(customDataJSONMap), &customDataHeaders); err != nil {
			log.Entry().Warn("[WARN] ===> customDataJSONMap flag must be a valid JSON map. Check the value of --customDataJSONMap and try again.")
		} else {
			for k, v := range customDataHeaders {
				headers["META-"+strings.ToUpper(k)] = []string{v}
			}
		}
	}

	headers["Group"] = []string{group}
	headers["Delete-Binary"] = []string{fmt.Sprintf("%v", deleteBinary)}
	headers["Url"] = []string{fetchURL}
	headers["Content-Type"] = []string{"application/json"}
	if (replaceBinary) && (version != "") {
		log.Entry().Debugf("[DEBUG][FETCH_URL] ===> replaceBinary && version != empty ")
		headers["Replace"] = []string{fmt.Sprintf("%v", productID)}
		headers["Version"] = []string{version}
	} else if replaceBinary {
		log.Entry().Debugf("[DEBUG][FETCH_URL] ===> replaceBinary")
		headers["Replace"] = []string{fmt.Sprintf("%v", productID)}
	} else if version != "" {
		log.Entry().Debugf("[DEBUG][FETCH_URL] ===> version != empty ")
		headers["Version"] = []string{version}
	}

	protecodeURL := fmt.Sprintf("%v/api/fetch/", pc.serverURL)
	r, statusCode, err := pc.sendAPIRequest(http.MethodPost, protecodeURL, headers)
	if err != nil {
		//TODO: bubble up error
		pc.logger.WithError(err).Fatalf("Error during declare fetch url: %v", protecodeURL)
	}

	// For replaceBinary option response doesn't contain any result but just a message saying that product successfully replaced.
	if replaceBinary && statusCode == 201 {
		result := new(ResultData)
		result.Result.ProductID = productID
		return result

	} else {
		result := new(ResultData)
		pc.mapResponse(*r, result)
		return result
	}

	// return result
}

// 2021-04-20 d :
// Found, via web search, an announcement that the set of status codes is expanding from
// B, R, F
// to
// B, R, F, S, D, P.
// Only R and F indicate work has completed.
func scanInProgress(status string) bool {
	return status != statusReady && status != statusFailed
}

// PollForResult polls the protecode scan for the result scan
func (pc *Protecode) PollForResult(productID int, timeOutInMinutes string) ResultData {

	var response ResultData
	var err error

	ticker := time.NewTicker(protecodePollInterval)
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
		if !scanInProgress(response.Result.Status) {
			ticker.Stop()
			i = 0
			break
		}

		select {
		case t := <-ticker.C:
			pc.logger.Debugf("Tick : %v Processing status for productID %v", t, productID)
		}
	}

	if scanInProgress(response.Result.Status) {
		response, err = pc.pullResult(productID)

		if len(response.Result.Components) < 1 {
			// 2020-04-20 d :
			// We are required to scan all images including 3rd party ones.
			// We have found that Crossplane makes use docker images that contain no
			// executable code.
			// So we can no longer treat an empty Components list as an error.
			pc.logger.Warn("Protecode scan did not identify any components.")
		}

		if err != nil || response.Result.Status == statusBusy {
			//TODO: bubble up error
			pc.logger.Fatalf("No result after polling err: %v protecode status: %v", err, response.Result.Status)
		}
	}

	return response
}

func (pc *Protecode) pullResult(productID int) (ResultData, error) {
	protecodeURL := pc.createURL("/api/product/", fmt.Sprintf("%v/", productID), "")
	headers := map[string][]string{
		"acceptType": {"application/json"},
	}
	r, _, err := pc.sendAPIRequest(http.MethodGet, protecodeURL, headers)

	if err != nil {
		return *new(ResultData), err
	}
	result := new(ResultData)
	pc.mapResponse(*r, result)

	return *result, nil

}

// verify provided product id
func (pc *Protecode) VerifyProductID(ProductID int) bool {
	pc.logger.Infof("Verification of product id (%v) started ... ", ProductID)

	// TODO: Optimise product id verification
	_, err := pc.pullResult(ProductID)

	// If response has an error then we assume this product id doesn't exist or user has no access
	if err != nil {
		return false
	}

	// Otherwise product exists
	return true

}

// LoadExistingProduct loads the existing product from protecode service
func (pc *Protecode) LoadExistingProduct(group string, fileName string) int {
	var productID int = -1

	protecodeURL := pc.createURL("/api/apps/", fmt.Sprintf("%v/", group), fileName)
	headers := map[string][]string{
		"acceptType": {"application/json"},
	}

	response := pc.loadExisting(protecodeURL, headers)

	if len(response.Products) > 0 {
		// Highest product id means the latest scan for this particular product, therefore we take a product id with the highest number
		for i := 0; i < len(response.Products); i++ {
			// Check filename, it should be the same as we searched
			if response.Products[i].FileName == fileName {
				if productID < response.Products[i].ProductID {
					productID = response.Products[i].ProductID
				}
			}
		}
	}

	return productID
}

//

func (pc *Protecode) loadExisting(protecodeURL string, headers map[string][]string) *ProductData {

	r, _, err := pc.sendAPIRequest(http.MethodGet, protecodeURL, headers)
	if err != nil {
		//TODO: bubble up error
		pc.logger.WithError(err).Fatalf("Error during load existing product: %v", protecodeURL)
	}

	result := new(ProductData)
	pc.mapResponse(*r, result)

	return result
}
