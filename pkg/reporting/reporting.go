package reporting

import (
	"bytes"
	"text/template"
	"time"

	"github.com/pkg/errors"
)

// ScanReport defines the elements of a scan report used by various scan steps
type ScanReport struct {
	Title       string
	Subheaders  []string
	Overview    []string
	FurtherInfo string
	Timestamp   time.Time
	DetailTable ScanDetailTable
}

// ScanDetailTable defines a table containing scan result details
type ScanDetailTable struct {
	Headers       []string
	Rows          []ScanRow
	WithCounter   bool
	CounterHeader string
	NoRowsMessage string
}

// ScanRow defines one row of a scan result table
type ScanRow struct {
	Columns []ScanColumn
}

// ScanColumn defines one column of a scan result table
type ScanColumn struct {
	Content string
	Style   ColumnStyle
}

// ColumnStyle defines style for a specific column
type ColumnStyle int

// enum for style types
const (
	Green = iota
	Yellow
	Red
	Grey
	Black
)

func (c ColumnStyle) String() string {
	return [...]string{"green", "yellow", "red", "grey", "black"}[c]
}

const reportHTMLTemplate = `<!DOCTYPE html>
<html>
<head>
	<title>{{.Title}}</title>
	<style type="text/css">{{.Style}}</style>
</head>
<body>
	<h1>{{.Title}}</h1>
	<h2>
		<span>
		{{range $s := .Subheaders }}
		{{$s}}<br />
		{{end}}
		</span>
	</h2>
	<div>
		<h3>
		{{range $o := .Overview }}
		{{$o}}<br />
		{{end}}
		</h3>
		{{.FurtherInfo}}
	</div>
	<p>Snapshot taken:{{.CurrentTime reportTime}}</p>
	<table>
	<tr>
		{{if .DetailTable.WithCounter}}<th>{{.DetailTable.CounterHeader}}</th>
		{{range $h := .DetailTable.Headers }}
		<th>{{$h}}</th>
		{{end}}
	</tr>

	{{if not .DetailTable.Rows}}
	{{.DetailTable.NoRowsMessage}}
	{{else}}
	{{range $i, $r := .DetailTable.Rows}}
	<tr>
	{{if .DetailTable.WithCounter}}<td>{{inc $i}}</td>
	{{range $c := $r.Columns}}
		<td {{if $c.Style}}class="{{c.Sytle}}">{{$c.Content}}</td>
	{{end}}
	</tr>
	{{end}}
	</table>
</body>
</html>`

// ToHTML creates a HTML version of the report
func (s *ScanReport) ToHTML() []byte {
	report := []byte{}
	tmpl, err := template.New("report").Funcs(funcMap).Parse(reportHTMLTemplate)
	if err != nil {
		return report, errors.Wrap(err, "failed to create HTML report template")
	}
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, reportInput)
	if err != nil {
		return report, errors.Wrap(err, "failed to execute HTML report template")
	}
	return buf.Bytes(), nil
}

// ToMarkdown creates a markdown version of the report content
func (s *ScanReport) ToMarkdown() string {
	/*
		## collapsible markdown?

		<details><summary>CLICK ME</summary>
		<p>

		#### yes, even hidden code blocks!

		```python
		print("hello world!")
		```

		</p>
		</details>
	*/
	return ""
}
