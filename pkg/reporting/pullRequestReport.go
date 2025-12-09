package reporting

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"github.com/SAP/jenkins-library/pkg/log"
)

// Components - for parsing from file
type Components []Component

type Component struct {
	ComponentName                  string                         `json:"componentName"`
	ComponentVersion               string                         `json:"versionName"`
	ComponentIdentifier            string                         `json:"componentIdentifier"`
	ViolatingPolicyNames           []string                       `json:"violatingPolicyNames"`
	PolicyViolationVulnerabilities []PolicyViolationVulnerability `json:"policyViolationVulnerabilities"`
	PolicyViolationLicenses        []PolicyViolationLicense       `json:"policyViolationLicenses"`
	WarningMessage                 string                         `json:"warningMessage"`
	ErrorMessage                   string                         `json:"errorMessage"`
}

type PolicyViolationVulnerability struct {
	Name                 string   `json:"name"`
	ViolatingPolicyNames []string `json:"ViolatingPolicyNames"`
	WarningMessage       string   `json:"warningMessage"`
	ErrorMessage         string   `json:"errorMessage"`
	Meta                 Meta     `json:"_meta"`
}

type PolicyViolationLicense struct {
	LicenseName          string   `json:"licenseName"`
	ViolatingPolicyNames []string `json:"violatingPolicyNames"`
	Meta                 Meta     `json:"_meta"`
}

type Meta struct {
	Href string `json:"href"`
}

// RapidScanReport - for commenting to pull requests
type RapidScanReport struct {
	Success bool

	ExecutedTime string

	MainTableHeaders []string
	MainTableValues  [][]string

	VulnerabilitiesTable []Vulnerabilities
	LicensesTable        []Licenses
	OtherViolationsTable []OtherViolations
}

type Vulnerabilities struct {
	PolicyViolationName string
	Values              []Vulnerability
}

type Vulnerability struct {
	VulnerabilityID    string
	VulnerabilityScore string
	ComponentName      string
	VulnerabilityHref  string
}

type Licenses struct {
	PolicyViolationName string
	Values              []License
}

type License struct {
	LicenseName   string
	ComponentName string
	LicenseHref   string
}

type OtherViolations struct {
	PolicyViolationName string
	Values              []OtherViolation
}

type OtherViolation struct {
	ComponentName string
}

const rapidReportMdTemplate = `
 {{if .Success}}:heavy_check_mark: **OSS related checks passed successfully**
  :clipboard: OSS related checks executed by Black Duck - rapid scan passed successfully.
 <h4><a href="https://documentation.blackduck.com/bundle/detect/page/runningdetect/rapidscan.html">RAPID SCAN</a></h4>

{{else}} :x: **OSS related checks failed**
  :clipboard: Policies violated by added OSS components
 <table>
 <tr>{{range $s := .MainTableHeaders -}}<td><b>{{$s}}</b></td>{{- end}}</tr>
 {{range $s := .MainTableValues -}}<tr>{{range $s1 := $s }}<td>{{$s1}}</td>{{- end}}</tr>
 {{- end}}
 </table>

{{range $index := .VulnerabilitiesTable -}}
<details><summary>
{{$len := len $index.Values}}
{{if le $len 1}} <h4> {{$len}} Policy Violation of {{$index.PolicyViolationName}}</h4>
{{else}}<h4> {{$len}} Policy Violations of {{$index.PolicyViolationName}} </h4> {{end}}
</summary>
	<table>
		<tr><td><b>Vulnerability ID</b></td><td><b>Vulnerability Score</b></td><td><b>Component Name</b></td></tr>
		{{range $value := $index.Values -}}
			<tr>
			<td> <a href="{{$value.VulnerabilityHref}}"> {{$value.VulnerabilityID}} </a> </td><td>{{$value.VulnerabilityScore}}</td><td>{{$value.ComponentName}}</td>
			</tr>
		{{end -}}
	</table>
</details>
{{end -}}
{{range $index := .LicensesTable -}}
<details><summary>
{{$len := len $index.Values}}
{{if le $len 1}} <h4> {{$len}} Policy Violation of {{$index.PolicyViolationName}}</h4>
{{else}}<h4> {{$len}} Policy Violations of {{$index.PolicyViolationName}} </h4> {{end}}
</summary>
	<table>
		<tr><td><b>License Name</b></td><td><b>Component Name</b></td></tr>
		{{range $value := $index.Values -}}
			<tr><td> <a href="{{$value.LicenseHref}}"> {{$value.LicenseName}} </a> </td><td>{{$value.ComponentName}}</td></tr>
		{{end -}}
	</table>
</details>
{{end -}}
{{range $index := .OtherViolationsTable -}}
<details><summary>
{{$len := len $index.Values}}
{{if le $len 1}} <h4> {{$len}} Policy Violation of {{$index.PolicyViolationName}}</h4>
{{else}}<h4> {{$len}} Policy Violations of {{$index.PolicyViolationName}} </h4> {{end}}
</summary>
	<table>
		<tr><td><b>Component Name</b></td></tr>
		{{range $value := $index.Values -}}
			<tr><td>{{$value.ComponentName}}</td></tr>
		{{end -}}
	</table>
</details>
{{end -}}
{{end}}
`

// RapidScanResult reads result of Rapid scan from generated file
func RapidScanResult(dir string) (string, error) {
	components, removeDir, err := findAndReadJsonFile(dir)
	if err != nil {
		return "", err
	}
	if components == nil {
		return "", errors.New("couldn't parse info from file")
	}

	buf, err := createMarkdownReport(components)
	if err != nil {
		return "", err
	}

	err = os.RemoveAll(removeDir)
	if err != nil {
		log.Entry().Error("Couldn't remove report file", err)
	}

	return buf.String(), nil
}

type Files []os.DirEntry

// findLastCreatedDir finds last created directory
func findLastCreatedDir(directories []os.DirEntry) os.DirEntry {
	lastCreatedDir := directories[0]
	for _, dir := range directories {
		if dir.Name() > lastCreatedDir.Name() {
			lastCreatedDir = dir
		}
	}
	return lastCreatedDir
}

// findAndReadJsonFile find file BlackDuck_DeveloperMode_Result.json generated by detectExecuteStep and read it
func findAndReadJsonFile(dir string) (*Components, string, error) {
	var err error
	filePath := dir + "/runs"
	allFiles, err := os.ReadDir(filePath)
	if err != nil {
		return nil, "", err
	}
	if allFiles == nil {
		return nil, "", errors.New("no report files")
	}
	lastDir := findLastCreatedDir(allFiles)
	removeDir := filePath + "/" + lastDir.Name()
	filePath = filePath + "/" + lastDir.Name() + "/scan"
	files, err := os.ReadDir(filePath)
	if err != nil {
		return nil, "", err
	}
	if files == nil {
		return nil, "", errors.New("no report files")
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), "BlackDuck_DeveloperMode_Result.json") {
			var result Components
			jsonFile, err := os.Open(filePath + "/" + file.Name())
			if err != nil {
				return nil, "", err
			}
			fileBody, err := io.ReadAll(jsonFile)
			if err != nil {
				return nil, "", err
			}
			err = json.Unmarshal(fileBody, &result)
			if err != nil {
				return nil, "", err
			}
			err = jsonFile.Close()
			if err != nil {
				log.Entry().Error(fmt.Sprintf("Couldn't close %s", jsonFile.Name()), err)
			}
			return &result, removeDir, nil
		}
	}

	return nil, "", nil
}

// createMarkdownReport creates markdown report to upload it as GitHub PR comment
func createMarkdownReport(components *Components) (*bytes.Buffer, error) {
	// preparing report
	var scanReport RapidScanReport
	scanReport.Success = true

	// getting reports to maps
	allPolicyViolationsMapUsed := make(map[string]bool)
	countPolicyViolationComponent := make(map[string]map[string]int)
	vulnerabilities := make(map[string][]Vulnerability)
	licenses := make(map[string][]License)
	otherViolations := make(map[string][]OtherViolation)
	componentNames := make([]string, len(*components))

	for idx, component := range *components {
		componentName := component.ComponentName + " " + component.ComponentVersion + " (" + component.ComponentIdentifier + ")"
		componentNames[idx] = componentName

		// for others
		for _, policyViolationName := range component.ViolatingPolicyNames {
			if !allPolicyViolationsMapUsed[policyViolationName] {
				allPolicyViolationsMapUsed[policyViolationName] = true
				scanReport.MainTableHeaders = append(scanReport.MainTableHeaders, policyViolationName)
			}
			if countPolicyViolationComponent[policyViolationName] == nil {
				countPolicyViolationComponent[policyViolationName] = make(map[string]int)
			}
			msg := component.ErrorMessage + " " + component.WarningMessage
			if strings.Contains(msg, policyViolationName) {
				countPolicyViolationComponent[policyViolationName][componentName]++
				otherViolations[policyViolationName] = append(otherViolations[policyViolationName], OtherViolation{ComponentName: componentName})
			}
		}

		// for Vulnerabilities
		for _, policyVulnerability := range component.PolicyViolationVulnerabilities {
			for _, policyViolationName := range policyVulnerability.ViolatingPolicyNames {
				if countPolicyViolationComponent[policyViolationName] == nil {
					countPolicyViolationComponent[policyViolationName] = make(map[string]int)
				}
				countPolicyViolationComponent[policyViolationName][componentName]++
				vulnerabilities[policyViolationName] = append(vulnerabilities[policyViolationName],
					Vulnerability{
						VulnerabilityID:    policyVulnerability.Name,
						VulnerabilityHref:  policyVulnerability.Meta.Href,
						VulnerabilityScore: getScore(policyVulnerability.ErrorMessage, "score") + " " + getScore(policyVulnerability.ErrorMessage, "severity"),
						ComponentName:      componentName,
					})
			}
		}

		// for Licenses
		for _, policyViolationLicense := range component.PolicyViolationLicenses {
			for _, policyViolationName := range policyViolationLicense.ViolatingPolicyNames {
				if countPolicyViolationComponent[policyViolationName] == nil {
					countPolicyViolationComponent[policyViolationName] = make(map[string]int)
				}
				countPolicyViolationComponent[policyViolationName][componentName]++
				licenses[policyViolationName] = append(licenses[policyViolationName],
					License{
						LicenseName:   policyViolationLicense.LicenseName,
						LicenseHref:   policyViolationLicense.Meta.Href + "/license-terms",
						ComponentName: componentName,
					})
			}
		}
	}

	if scanReport.MainTableHeaders != nil && componentNames != nil {
		scanReport.Success = false

		// MainTable sort & copy
		sort.Strings(scanReport.MainTableHeaders)
		sort.Strings(componentNames)
		scanReport.MainTableHeaders = append([]string{"Component name"}, scanReport.MainTableHeaders...)
		for i := range componentNames {
			scanReport.MainTableValues = append(scanReport.MainTableValues, []string{})
			scanReport.MainTableValues[i] = append(scanReport.MainTableValues[i], componentNames[i])
			for j := 1; j < len(scanReport.MainTableHeaders); j++ {
				policyV := scanReport.MainTableHeaders[j]
				comp := componentNames[i]
				count := strconv.Itoa(countPolicyViolationComponent[policyV][comp])
				scanReport.MainTableValues[i] = append(scanReport.MainTableValues[i], count)
			}
		}

		// VulnerabilitiesTable sort & copy
		for key := range vulnerabilities {
			item := vulnerabilities[key]
			sort.Slice(item, func(i, j int) bool {
				return scoreLogicSort(item[i].VulnerabilityScore, item[j].VulnerabilityScore)
			})
			scanReport.VulnerabilitiesTable = append(scanReport.VulnerabilitiesTable, Vulnerabilities{
				PolicyViolationName: key,
				Values:              item,
			})
		}
		sort.Slice(scanReport.VulnerabilitiesTable, func(i, j int) bool {
			return scanReport.VulnerabilitiesTable[i].PolicyViolationName < scanReport.VulnerabilitiesTable[j].PolicyViolationName
		})

		// LicensesTable sort & copy
		for key := range licenses {
			item := licenses[key]
			sort.Slice(item, func(i, j int) bool {
				if item[i].LicenseName < item[j].LicenseName {
					return true
				}
				if item[i].LicenseName > item[j].LicenseName {
					return false
				}
				return item[i].ComponentName < item[j].ComponentName
			})
			scanReport.LicensesTable = append(scanReport.LicensesTable, Licenses{
				PolicyViolationName: key,
				Values:              item,
			})
		}
		sort.Slice(scanReport.LicensesTable, func(i, j int) bool {
			return scanReport.LicensesTable[i].PolicyViolationName < scanReport.LicensesTable[j].PolicyViolationName
		})

		// OtherViolationsTable sort & copy
		for key := range otherViolations {
			item := otherViolations[key]
			sort.Slice(item, func(i, j int) bool {
				return item[i].ComponentName < item[j].ComponentName
			})
			scanReport.OtherViolationsTable = append(scanReport.OtherViolationsTable, OtherViolations{
				PolicyViolationName: key,
				Values:              item,
			})
		}
		sort.Slice(scanReport.OtherViolationsTable, func(i, j int) bool {
			return scanReport.OtherViolationsTable[i].PolicyViolationName < scanReport.OtherViolationsTable[j].PolicyViolationName
		})
	}

	tmpl, err := template.New("report").Parse(rapidReportMdTemplate)
	if err != nil {
		return nil, errors.New("failed to create Markdown report template err:" + err.Error())
	}
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, scanReport)
	if err != nil {
		return nil, errors.New("failed to create Markdown report template err:" + err.Error())
	}

	return buf, nil
}

// getScore extracts score or severity from error message
func getScore(message, key string) string {
	indx := strings.Index(message, key)
	if indx == -1 {
		return ""
	}
	var result string
	var notFirstSpace bool
	for _, s := range message[indx+len(key):] {
		if s == ' ' && notFirstSpace {
			break
		}
		notFirstSpace = true
		result = result + string(s)
	}
	return strings.Trim(result, " ")
}

// scoreLogicSort sorts two scores
func scoreLogicSort(iStr, jStr string) bool {
	if strings.Contains(iStr, "10.0") {
		return true
	} else if strings.Contains(jStr, "10.0") {
		return false
	}
	if iStr >= jStr {
		return true
	}
	return false
}
