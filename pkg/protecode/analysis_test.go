package protecode

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSevere(t *testing.T) {
	t.Run("with severe cvss v3 vulnerability", func(t *testing.T) {
		// init
		vulnerability := Vulnerability{
			Exact:  true,
			Triage: []Triage{},
			Vuln: Vuln{
				Cve:        "Cve2",
				Cvss:       8.0,
				Cvss3Score: "7.3",
			},
		}
		// test && assert
		assert.Equal(t, true, isSevere(vulnerability))
	})
	t.Run("with severe cvss v2 vulnerability", func(t *testing.T) {
		// init
		vulnerability := Vulnerability{
			Exact:  true,
			Triage: []Triage{},
			Vuln: Vuln{
				Cve:        "Cve2",
				Cvss:       8.0,
				Cvss3Score: "0.0",
			},
		}
		// test && assert
		assert.Equal(t, true, isSevere(vulnerability))
	})
	t.Run("with non-severe cvss v3 vulnerability", func(t *testing.T) {
		// init
		vulnerability := Vulnerability{
			Exact:  true,
			Triage: []Triage{},
			Vuln: Vuln{
				Cve:        "Cve2",
				Cvss:       4.0,
				Cvss3Score: "4.0",
			},
		}
		// test && assert
		assert.Equal(t, false, isSevere(vulnerability))
	})
	t.Run("with non-severe cvss v2 vulnerability", func(t *testing.T) {
		// init
		vulnerability := Vulnerability{
			Exact:  true,
			Triage: []Triage{},
			Vuln: Vuln{
				Cve:        "Cve2",
				Cvss:       4.0,
				Cvss3Score: "0.0",
			},
		}
		// test && assert
		assert.Equal(t, false, isSevere(vulnerability))
	})
	t.Run("with non-severe vulnerability with missing cvss v3 rating", func(t *testing.T) {
		// init
		vulnerability := Vulnerability{
			Exact:  true,
			Triage: []Triage{},
			Vuln: Vuln{
				Cve:        "Cve2",
				Cvss:       4.0,
				Cvss3Score: "",
			},
		}
		// test && assert
		assert.Equal(t, false, isSevere(vulnerability))
	})
}

func TestHasSevereVulnerabilities(t *testing.T) {
	severeV3 := Vulnerability{Exact: true, Triage: []Triage{}, Vuln: Vuln{Cve: "Cve1", Cvss: 4.0, Cvss3Score: "8.0"}}
	severeV2 := Vulnerability{Exact: true, Triage: []Triage{}, Vuln: Vuln{Cve: "Cve2", Cvss: 8.0, Cvss3Score: "0.0"}}
	nonSevere1 := Vulnerability{Exact: true, Triage: []Triage{}, Vuln: Vuln{Cve: "Cve3", Cvss: 4.0, Cvss3Score: "4.0"}}
	nonSevere2 := Vulnerability{Exact: true, Triage: []Triage{}, Vuln: Vuln{Cve: "Cve4", Cvss: 4.0, Cvss3Score: "4.0"}}
	excluded := Vulnerability{Exact: true, Triage: []Triage{}, Vuln: Vuln{Cve: "Cve5", Cvss: 8.0, Cvss3Score: "8.0"}}
	triaged := Vulnerability{Exact: true, Triage: []Triage{{ID: 1}}, Vuln: Vuln{Cve: "Cve6", Cvss: 8.0, Cvss3Score: "8.0"}}
	historic := Vulnerability{Exact: false, Triage: []Triage{}, Vuln: Vuln{Cve: "Cve7", Cvss: 8.0, Cvss3Score: "8.0"}}

	t.Run("with severe v3 vulnerabilities", func(t *testing.T) {
		// init
		data := Result{Components: []Component{{Vulns: []Vulnerability{nonSevere1, severeV3}}}}
		// test && assert
		assert.Equal(t, true, HasSevereVulnerabilities(data, ""))
	})
	t.Run("with severe v2 vulnerabilities", func(t *testing.T) {
		// init
		data := Result{Components: []Component{{Vulns: []Vulnerability{nonSevere1, severeV2}}}}
		// test && assert
		assert.Equal(t, true, HasSevereVulnerabilities(data, ""))
	})
	t.Run("without severe vulnerabilities", func(t *testing.T) {
		// init
		data := Result{Components: []Component{{Vulns: []Vulnerability{nonSevere1, nonSevere2}}}}
		// test && assert
		assert.Equal(t, false, HasSevereVulnerabilities(data, ""))
	})
	t.Run("with historic vulnerabilities", func(t *testing.T) {
		// init
		data := Result{Components: []Component{{Vulns: []Vulnerability{nonSevere1, triaged}}}}
		// test && assert
		assert.Equal(t, false, HasSevereVulnerabilities(data, ""))
	})
	t.Run("with excluded vulnerabilities", func(t *testing.T) {
		// init
		data := Result{Components: []Component{{Vulns: []Vulnerability{nonSevere1, excluded}}}}
		// test && assert
		assert.Equal(t, false, HasSevereVulnerabilities(data, "Cve5,Cve14"))
	})
	t.Run("with historic vulnerabilities", func(t *testing.T) {
		// init
		data := Result{Components: []Component{{Vulns: []Vulnerability{nonSevere1, historic}}}}
		// test && assert
		assert.Equal(t, false, HasSevereVulnerabilities(data, ""))
	})
}
