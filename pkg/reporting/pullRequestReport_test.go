//go:build unit

package reporting

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Testing createMarkdownReport function
func TestCreateMarkdownReport(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName       string
		components     *Components
		expectedErr    error
		expectedReport string
	}{

		{
			testName: "Vulnerabilities were found",
			components: &Components{
				{
					ComponentName:       "qs -  QS Querystring",
					ComponentVersion:    "5.2.1",
					ComponentIdentifier: "npmjs:qs/5.2.1",
					ViolatingPolicyNames: []string{
						"High Vulnerability Security Issue",
					},
					PolicyViolationVulnerabilities: []PolicyViolationVulnerability{
						{
							Name:                 "CVE-2017-1000048",
							ViolatingPolicyNames: []string{"High Vulnerability Security Issue"},
							WarningMessage:       "",
							ErrorMessage: "Component qs -  QS Querystring version 5.2.1 with ID npmjs:qs/5.2.1 violates policy" +
								" High Vulnerability Security Issue: found vulnerability CVE-2017-1000048 with severity HIGH and CVSS score 7.5",
							Meta: Meta{
								Href: "https://sap-staging.app.blackduck.com/api/vulnerabilities/CVE-2017-1000048",
							},
						},
					},
					PolicyViolationLicenses: nil,
					WarningMessage:          "",
					ErrorMessage:            "",
				},
				{
					ComponentName:       "Lodash",
					ComponentVersion:    "4.17.10",
					ComponentIdentifier: "npmjs:lodash/4.17.10",
					ViolatingPolicyNames: []string{
						"High Vulnerability Security Issue",
						"Test High Severity Vuln Filter",
						"OutdatedFOSSLibraries",
					},
					PolicyViolationVulnerabilities: []PolicyViolationVulnerability{
						{
							Name: "CVE-2019-10744",
							ViolatingPolicyNames: []string{
								"High Vulnerability Security Issue",
								"Test High Severity Vuln Filter",
							},
							WarningMessage: "Component Lodash version 4.17.10 with ID npmjs:lodash/4.17.10 violates policy Test High Severity Vuln " +
								"Filter: found vulnerability CVE-2019-10744 with severity CRITICAL and CVSS score 9.1",
							ErrorMessage: "Component Lodash version 4.17.10 with ID npmjs:lodash/4.17.10 violates policy High Vulnerability " +
								"Security Issue: found vulnerability CVE-2019-10744 with severity CRITICAL and CVSS score 9.1",
							Meta: Meta{
								Href: "https://sap-staging.app.blackduck.com/api/vulnerabilities/CVE-2019-10744"},
						},
						{
							Name: "CVE-2020-8203",
							ViolatingPolicyNames: []string{
								"High Vulnerability Security Issue",
								"Test High Severity Vuln Filter",
							},
							WarningMessage: "Component Lodash version 4.17.10 with ID npmjs:lodash/4.17.10 violates policy Test " +
								"High Severity Vuln Filter: found vulnerability CVE-2020-8203 with severity HIGH and CVSS score 7.4",
							ErrorMessage: "Component Lodash version 4.17.10 with ID npmjs:lodash/4.17.10 violates policy Test High Severity Vuln Filter: " +
								"found vulnerability CVE-2020-8203 with severity HIGH and CVSS score 7.4",
							Meta: Meta{
								Href: "https://sap-staging.app.blackduck.com/api/vulnerabilities/CVE-2020-8203",
							},
						},
						{
							Name: "BDSA-2019-3842",
							ViolatingPolicyNames: []string{
								"High Vulnerability Security Issue",
								"Test High Severity Vuln Filter",
							},
							WarningMessage: "Component Lodash version 4.17.10 with ID npmjs:lodash/4.17.10 violates policy Test High Severity Vuln Filter: found vulnerability BDSA-2019-3842 with severity HIGH and CVSS score 7.1",
							ErrorMessage:   "Component Lodash version 4.17.10 with ID npmjs:lodash/4.17.10 violates policy High Vulnerability Security Issue: found vulnerability BDSA-2019-3842 with severity HIGH and CVSS score 7.1",
							Meta: Meta{
								Href: "https://sap-staging.app.blackduck.com/api/vulnerabilities/BDSA-2019-3842",
							},
						},
					},
					PolicyViolationLicenses: nil,
					WarningMessage:          "Component Lodash version 4.17.10 with ID npmjs:lodash/4.17.10 violates policy OutdatedFOSSLibraries",
					ErrorMessage:            "",
				},
				{
					ComponentName:       "Chalk",
					ComponentVersion:    "1.1.3",
					ComponentIdentifier: "npmjs:chalk/1.1.3",
					ViolatingPolicyNames: []string{
						"OutdatedFOSSLibraries",
					},
					PolicyViolationVulnerabilities: nil,
					PolicyViolationLicenses:        nil,
					WarningMessage:                 "Component Chalk version 1.1.3 with ID npmjs:chalk/1.1.3 violates policy OutdatedFOSSLibraries",
					ErrorMessage:                   "",
				},
			},
			expectedReport: "\n  :x: **OSS related checks failed**\n  :clipboard: Policies violated by added OSS components\n " +
				"<table>\n <tr><td><b>Component name</b></td><td><b>High Vulnerability Security Issue</b></td><td><b>OutdatedFOSSLibraries</b></td><td><b>" +
				"Test High Severity Vuln Filter</b></td></tr>\n <tr><td>Chalk 1.1.3 (npmjs:chalk/1.1.3)</td><td>0</td><td>1</td><td>0</td></tr><tr><td>Lodash " +
				"4.17.10 (npmjs:lodash/4.17.10)</td><td>3</td><td>1</td><td>3</td></tr><tr><td>qs -  QS Querystring 5.2.1 " +
				"(npmjs:qs/5.2.1)</td><td>1</td><td>0</td><td>0</td></tr>\n </table>\n\n<details><summary>\n\n<h4> 4 Policy " +
				"Violations of High Vulnerability Security Issue </h4> \n</summary>\n\t<table>\n\t\t<tr><td><b>Vulnerability ID</b></td><td><b>Vulnerability" +
				" Score</b></td><td><b>Component Name</b></td></tr>\n\t\t<tr>\n\t\t\t<td> <a href=\"https://sap-staging.app.blackduck.com/api/vulnerabilities/CVE-2019-10744\"> CVE-2019-10744 </a> </td><td>9.1 CRITICAL</td><td>Lodash 4.17.10 " +
				"(npmjs:lodash/4.17.10)</td>\n\t\t\t</tr>\n\t\t<tr>\n\t\t\t<td> <a href=\"https://sap-staging.app.blackduck.com/api/vulnerabilities/CVE-2017-1000048\"> " +
				"CVE-2017-1000048 </a> </td><td>7.5 HIGH</td><td>qs -  QS Querystring 5.2.1 (npmjs:qs/5.2.1)</td>\n\t\t\t</tr>\n\t\t<tr>\n\t\t\t<td> " +
				"<a href=\"https://sap-staging.app.blackduck.com/api/vulnerabilities/CVE-2020-8203\"> CVE-2020-8203 </a> </td><td>7.4 HIGH</td><td>Lodash " +
				"4.17.10 (npmjs:lodash/4.17.10)</td>\n\t\t\t</tr>\n\t\t<tr>\n\t\t\t<td> <a href=\"https://sap-staging.app.blackduck.com/api/vulnerabilities/BDSA-2019-3842\"> " +
				"BDSA-2019-3842 </a> </td><td>7.1 HIGH</td><td>Lodash 4.17.10 (npmjs:lodash/4.17.10)</td>\n\t\t\t</tr>\n\t\t</table>\n</details>\n<details><summary>\n\n<h4> " +
				"3 Policy Violations of Test High Severity Vuln Filter </h4> \n</summary>\n\t<table>\n\t\t<tr><td><b>Vulnerability ID</b></td><td><b>Vulnerability " +
				"Score</b></td><td><b>Component Name</b></td></tr>\n\t\t<tr>\n\t\t\t<td> <a href=\"https://sap-staging.app.blackduck.com/api/vulnerabilities/CVE-2019-10744\"> " +
				"CVE-2019-10744 </a> </td><td>9.1 CRITICAL</td><td>Lodash 4.17.10 (npmjs:lodash/4.17.10)</td>\n\t\t\t</tr>\n\t\t<tr>\n\t\t\t<td> " +
				"<a href=\"https://sap-staging.app.blackduck.com/api/vulnerabilities/CVE-2020-8203\"> CVE-2020-8203 </a> </td><td>7.4 " +
				"HIGH</td><td>Lodash 4.17.10 (npmjs:lodash/4.17.10)</td>\n\t\t\t</tr>\n\t\t<tr>\n\t\t\t<td> <a href=\"https://sap-staging.app.blackduck.com/api/vulnerabilities/BDSA-2019-3842\"> " +
				"BDSA-2019-3842 </a> </td><td>7.1 HIGH</td><td>Lodash 4.17.10 (npmjs:lodash/4.17.10)</td>\n\t\t\t</tr>\n\t\t</table>\n</details>\n<details><summary>\n\n<h4> " +
				"2 Policy Violations of OutdatedFOSSLibraries </h4> \n</summary>\n\t<table>\n\t\t<tr><td><b>Component Name</b></td></tr>\n\t\t<tr><td>Chalk 1.1.3 " +
				"(npmjs:chalk/1.1.3)</td></tr>\n\t\t<tr><td>Lodash 4.17.10 (npmjs:lodash/4.17.10)</td></tr>\n\t\t</table>\n</details>\n\n",
		},
		{
			testName:   "No vulnerabilities && successful build",
			components: &Components{},
			expectedReport: "\n :heavy_check_mark: **OSS related checks passed successfully**\n  :clipboard: OSS related checks executed by Black Duck " +
				"- rapid scan passed successfully.\n" +
				" <h4><a href=\"https://sig-product-docs.synopsys.com/bundle/integrations-detect/page/runningdetect/rapidscan.html\">" +
				"RAPID SCAN</a></h4>\n\n\n",
		},
	}

	for _, c := range testCases {
		t.Run(c.testName, func(t *testing.T) {
			t.Parallel()

			buf, err := createMarkdownReport(c.components)

			assert.Equal(t, c.expectedErr, err)
			assert.Equal(t, c.expectedReport, buf.String())
		})
	}
}

// Testing getScore function
func TestGetScore(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName string
		message  string
		key      string
		expected string
	}{
		{
			testName: "Score 7.5",
			message: "Component qs -  QS Querystring version 5.2.1 with ID npmjs:qs/5.2.1 violates policy High " +
				"Vulnerability Security Issue: found vulnerability CVE-2017-1000048 with severity HIGH and CVSS score 7.5",
			key:      "score",
			expected: "7.5",
		},
		{
			testName: "CRITICAL severity",
			message: "Component minimist version 0.0.8 with ID npmjs:minimist/0.0.8 violates policy High " +
				"Vulnerability Security Issue: found vulnerability CVE-2021-44906 with severity CRITICAL and CVSS score 9.8",
			key:      "severity",
			expected: "CRITICAL",
		},
		{
			testName: "No severity",
			message: "Component minimist version 0.0.8 with ID npmjs:minimist/0.0.8 violates policy High " +
				"Vulnerability Security Issue: found vulnerability CVE-2021-44906 with CVSS score 9.8",
			key:      "severity",
			expected: "",
		},
	}

	for _, c := range testCases {
		t.Run(c.testName, func(t *testing.T) {
			t.Parallel()

			got := getScore(c.message, c.key)
			assert.Equal(t, c.expected, got)
		})
	}
}

// Testing scoreLogicSort function
func TestScoreLogicSort(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName   string
		leftScore  string
		rightScore string
		expected   bool
	}{
		{
			testName:   "left score is higher",
			leftScore:  "8.8 HIGH",
			rightScore: "8.1 HIGH",
			expected:   true,
		},
		{
			testName:   "right score is higher",
			leftScore:  "7.9 HIGH",
			rightScore: "9.3 CRITICAL",
			expected:   false,
		},
		{
			testName:   "left score equals 10.0",
			leftScore:  "10.0 CRITICAL",
			rightScore: "8.1 HIGH",
			expected:   true,
		},
		{
			testName:   "right score equals 10.0",
			leftScore:  "7.9 HIGH",
			rightScore: "10.0 CRITICAL",
			expected:   false,
		},
		{
			testName:   "both scores equal 10.0",
			leftScore:  "10.0 CRITICAL",
			rightScore: "10.0 CRITICAL",
			expected:   true,
		},
	}

	for _, c := range testCases {
		t.Run(c.testName, func(t *testing.T) {
			t.Parallel()

			got := scoreLogicSort(c.leftScore, c.rightScore)
			assert.Equal(t, c.expected, got)
		})
	}
}
