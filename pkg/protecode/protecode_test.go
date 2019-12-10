package protecode

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseProteCodeResultSuccess(t *testing.T) {

	var result ProteCodeResult = ProteCodeResult{
		ProductId: "ProductId",
		ReportUrl: "ReportUrl",
		Status:    "B",
		Components: []ProteCodeComponent{
			{Vulnerability: []ProteCodeVulnerability{
				{Exact: true, Triage: "", Vuln: ProteCodeVuln{Cve: "Cve1", Cvss: 7.2, Cvss3Score: "0.0"}},
				{Exact: true, Triage: "triage2", Vuln: ProteCodeVuln{Cve: "Cve2", Cvss: 2.2, Cvss3Score: "2.3"}},
				{Exact: true, Triage: "", Vuln: ProteCodeVuln{Cve: "Cve2b", Cvss: 0.0, Cvss3Score: "0.0"}},
			},
			},
			{Vulnerability: []ProteCodeVulnerability{
				{Exact: true, Triage: "", Vuln: ProteCodeVuln{Cve: "Cve3", Cvss: 3.2, Cvss3Score: "7.3"}},
				{Exact: true, Triage: "", Vuln: ProteCodeVuln{Cve: "Cve4", Cvss: 8.0, Cvss3Score: "8.0"}},
				{Exact: false, Triage: "", Vuln: ProteCodeVuln{Cve: "Cve4b", Cvss: 8.0, Cvss3Score: "8.0"}},
			},
			},
		},
	}
	m := ParseResultToInflux(result, "Excluded CVES: Cve4,")
	t.Run("Parse Protecode Results", func(t *testing.T) {
		assert.Equal(t, 1, m["historical_vulnerabilities"])
		assert.Equal(t, 1, m["triaged_vulnerabilities"])
		assert.Equal(t, 1, m["excluded_vulnerabilities"])
		assert.Equal(t, 1, m["minor_vulnerabilities"])
		assert.Equal(t, 2, m["major_vulnerabilities"])
		assert.Equal(t, 3, m["vulnerabilities"])
	})
}

func TestCmdExecGetProtecodeResultSuccess(t *testing.T) {

	cases := []struct {
		cmdName   string
		cmdString string
		want      ProteCodeResult
	}{
		{"echo", "test", ProteCodeResult{ProductId: "productID2"}},
		{"echo", "Dummy-DeLiMiTeR-status=200", ProteCodeResult{ProductId: "productID1"}},
	}
	for _, c := range cases {

		got := CmdExecGetProtecodeResult(c.cmdName, c.cmdString)
		assert.Equal(t, c.want, got)
	}
}

func TestTest(t *testing.T) {
	cmd := exec.Command("echo", "-n", `{"Name": "Bob", "Age": 32}`)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		log.Fatal(err)
	}
	var person struct {
		Name string
		Age  int
	}
	if err := json.NewDecoder(stdout).Decode(&person); err != nil {
		log.Fatal(err)
	}
	if err := cmd.Wait(); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s is %d years old\n", person.Name, person.Age)
}
