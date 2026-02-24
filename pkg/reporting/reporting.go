package reporting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/template"
	"time"
)

// IssueDetail represents any content that can be transformed into the body of a GitHub issue
type IssueDetail interface {
	Title() string
	ToMarkdown() ([]byte, error)
	ToTxt() string
}

// ScanReport defines the elements of a scan report used by various scan steps
type ScanReport struct {
	StepName       string          `json:"stepName"`
	ReportTitle    string          `json:"title"`
	Subheaders     []Subheader     `json:"subheaders"`
	Overview       []OverviewRow   `json:"overview"`
	FurtherInfo    string          `json:"furtherInfo"`
	ReportTime     time.Time       `json:"reportTime"`
	DetailTable    ScanDetailTable `json:"detailTable"`
	SuccessfulScan bool            `json:"successfulScan"`
}

// ScanDetailTable defines a table containing scan result details
type ScanDetailTable struct {
	Headers       []string  `json:"headers"`
	Rows          []ScanRow `json:"rows"`
	WithCounter   bool      `json:"withCounter"`
	CounterHeader string    `json:"counterHeader"`
	NoRowsMessage string    `json:"noRowsMessage"`
}

// ScanRow defines one row of a scan result table
type ScanRow struct {
	Columns []ScanCell `json:"columns"`
}

// AddColumn adds a column to a dedicated ScanRow
func (s *ScanRow) AddColumn(content interface{}, style ColumnStyle) {
	if s.Columns == nil {
		s.Columns = []ScanCell{}
	}
	s.Columns = append(s.Columns, ScanCell{Content: fmt.Sprint(content), Style: style})
}

// ScanCell defines one column of a scan result table
type ScanCell struct {
	Content string      `json:"content"`
	Style   ColumnStyle `json:"style"`
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

// OverviewRow defines a row in the report's overview section
// it can consist of a description and some details where the details can have a style attached
type OverviewRow struct {
	Description string      `json:"description"`
	Details     string      `json:"details,omitempty"`
	Style       ColumnStyle `json:"style,omitempty"`
}

// Subheader defines a dedicated sub header in a report
type Subheader struct {
	Description string `json:"text"`
	Details     string `json:"details,omitempty"`
}

// AddSubHeader adds a sub header to the report containing of a text/title plus optional details
func (s *ScanReport) AddSubHeader(header, details string) {
	s.Subheaders = append(s.Subheaders, Subheader{Description: header, Details: details})
}

// StepReportDirectory specifies the default directory for markdown reports which can later be collected by step pipelineCreateSummary
const StepReportDirectory = ".pipeline/stepReports"

// ToJSON returns the report in JSON format
func (s *ScanReport) ToJSON() ([]byte, error) {
	return json.Marshal(s)
}

// ToTxt up to now returns the report in JSON format
func (s ScanReport) ToTxt() string {
	txt, _ := s.ToJSON()
	return string(txt)
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
		{{- $s.Description}}: {{$s.Details}}<br />
		{{end -}}
		</span>
	</h2>
	<div>
		<h3>
		{{range $o := .Overview}}
		{{- drawOverviewRow $o}}<br />
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
		"columnCount":     tableColumnCount,
		"drawCell":        drawCell,
		"drawOverviewRow": drawOverviewRow,
	}
	report := []byte{}
	tmpl, err := template.New("report").Funcs(funcMap).Parse(reportHTMLTemplate)
	if err != nil {
		return report, fmt.Errorf("failed to create HTML report template: %w", err)
	}
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, s)
	if err != nil {
		return report, fmt.Errorf("failed to execute HTML report template: %w", err)
	}
	return buf.Bytes(), nil
}

const reportMdTemplate = `## {{if .SuccessfulScan}}:white_check_mark:{{else}}:x:{{end}} {{.Title}}

<table>
{{range $s := .Subheaders -}}
	<tr><td><b>{{- $s.Description}}:</b></td><td>{{$s.Details}}</td></tr>
{{- end}}

{{range $o := .Overview -}}
{{drawOverviewRow $o}}
{{- end}}
</table>

{{.FurtherInfo}}

Snapshot taken: <i>{{reportTime .ReportTime}}</i>

{{if shouldDrawTable .DetailTable -}}
<details><summary><i>{{.Title}} details:</i></summary>
<p>

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
</p>
</details>
{{ end }}

`

// Title returns the title of the report
func (s ScanReport) Title() string {
	return s.ReportTitle
}

// ToMarkdown creates a markdown version of the report content
func (s ScanReport) ToMarkdown() ([]byte, error) {
	funcMap := template.FuncMap{
		"columnCount":     tableColumnCount,
		"drawCell":        drawCell,
		"shouldDrawTable": shouldDrawTable,
		"inc": func(i int) int {
			return i + 1
		},
		"reportTime": func(currentTime time.Time) string {
			return currentTime.Format("Jan 02, 2006 - 15:04:05 MST")
		},
		"drawOverviewRow": drawOverviewRowMarkdown,
	}
	report := []byte{}
	tmpl, err := template.New("report").Funcs(funcMap).Parse(reportMdTemplate)
	if err != nil {
		return report, fmt.Errorf("failed to create Markdown report template: %w", err)
	}
	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, s)
	if err != nil {
		return report, fmt.Errorf("failed to execute Markdown report template: %w", err)
	}
	return buf.Bytes(), nil
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

func shouldDrawTable(table ScanDetailTable) bool {
	if len(table.Headers) > 0 {
		return true
	}
	return false
}

func drawOverviewRow(row OverviewRow) string {
	// so far accept only accept max. two columns for overview table: description and content
	if len(row.Details) == 0 {
		return row.Description
	}
	// ToDo: allow styling of details
	return fmt.Sprintf("%v: %v", row.Description, row.Details)
}

func drawOverviewRowMarkdown(row OverviewRow) string {
	// so far accept only accept max. two columns for overview table: description and content
	if len(row.Details) == 0 {
		return row.Description
	}
	// ToDo: allow styling of details
	return fmt.Sprintf("<tr><td>%v:</td><td>%v</td></tr>", row.Description, row.Details)
}
