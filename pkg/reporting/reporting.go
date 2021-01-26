package reporting

import (
	"bytes"
	"fmt"
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
	ReportTime  time.Time
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
	Columns []ScanCell
}

// ScanCell defines one column of a scan result table
type ScanCell struct {
	Content string
	Style   ColumnStyle
}

// ColumnStyle defines style for a specific column
type ColumnStyle int

// enum for style types
const (
	Green = iota + 1
	Yellow
	Red
	Grey
	Black
)

func (c ColumnStyle) String() string {
	return [...]string{"", "green-cell", "yellow-cell", "red-cell", "grey-cell", "black-cell"}[c]
}

const reportHTMLTemplate = `<!DOCTYPE html>
<html>
<head>
	<title>{{.Title}}</title>
	<style type="text/css">
	body {
		font-family: Arial, Verdana;
	}
	table {
		border-collapse: collapse;
	}
	div.code {
		font-family: "Courier New", "Lucida Console";
	}
	th {
		border-top: 1px solid #ddd;
	}
	th, td {
		padding: 12px;
		text-align: left;
		border-bottom: 1px solid #ddd;
		border-right: 1px solid #ddd;
	}
	tr:nth-child(even) {
		background-color: #f2f2f2;
	}
	.bold {
		font-weight: bold;
	}
	.green{
		color: olivedrab;
	}
	.red{
		color: orangered;
	}
	.nobullets {
		list-style-type:none;
		padding-left: 0;
		padding-bottom: 0;
		margin: 0;
	}
	.green-cell {
		background-color: #e1f5a9;
		padding: 5px
	}
	.yellow-cell {
		background-color: #ffff99;
		padding: 5px
	}
	.red-cell {
		background-color: #ffe5e5;
		padding: 5px
	}
	.grey-cell{
		background-color: rgba(212, 212, 212, 0.7);
		padding: 5px;
	}
	.black-cell{
		background-color: rgba(0, 0, 0, 0.75);
		padding: 5px;
	}
	</style>
</head>
<body>
	<h1>{{.Title}}</h1>
	<h2>
		<span>
		{{range $s := .Subheaders}}
		{{- $s}}<br />
		{{end -}}
		</span>
	</h2>
	<div>
		<h3>
		{{range $o := .Overview}}
		{{- $o}}<br />
		{{end -}}
		</h3>
		<span>{{.FurtherInfo}}</span>
	</div>
	<p>Snapshot taken: {{reportTime .ReportTime}}</p>
	<table>
	<tr>
		{{if .DetailTable.WithCounter}}<th>{{.DetailTable.CounterHeader}}</th>{{end}}
		{{- range $h := .DetailTable.Headers}}
		<th>{{$h}}</th>
		{{- end}}
	</tr>
	{{range $i, $r := .DetailTable.Rows}}
	<tr>
		{{if $.DetailTable.WithCounter}}<td>{{inc $i}}</td>{{end}}
		{{- range $c := $r.Columns}}
		{{drawCell $c}}
		{{- end}}
	</tr>
	{{else}}
	<tr><td colspan="{{columnCount .DetailTable}}">{{.DetailTable.NoRowsMessage}}</td></tr>
	{{- end}}
	</table>
</body>
</html>
`

// ToHTML creates a HTML version of the report
func (s *ScanReport) ToHTML() ([]byte, error) {
	funcMap := template.FuncMap{
		"inc": func(i int) int {
			return i + 1
		},
		"reportTime": func(currentTime time.Time) string {
			return currentTime.Format("Jan 02, 2006 - 15:04:05 MST")
		},
		"columnCount": tableColumnCount,
		"drawCell":    drawCell,
	}
	report := []byte{}
	tmpl, err := template.New("report").Funcs(funcMap).Parse(reportHTMLTemplate)
	if err != nil {
		return report, errors.Wrap(err, "failed to create HTML report template")
	}
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, s)
	if err != nil {
		return report, errors.Wrap(err, "failed to execute HTML report template")
	}
	return buf.Bytes(), nil
}

// ToMarkdown creates a markdown version of the report content
func (s *ScanReport) ToMarkdown() string {
	//ToDo: create collapsible markdown?
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

func tableColumnCount(scanDetails ScanDetailTable) int {
	colCount := len(scanDetails.Headers)
	if scanDetails.WithCounter {
		colCount++
	}
	return colCount
}

func drawCell(cell ScanCell) string {
	if cell.Style > 0 {
		return fmt.Sprintf(`<td class="%v">%v</td>`, cell.Style, cell.Content)
	}
	return fmt.Sprintf(`<td>%v</td>`, cell.Content)
}
