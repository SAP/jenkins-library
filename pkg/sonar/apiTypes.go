package sonar

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

// The types and response handling in this file are copied from github.com/magicsong/sonargo (Apache-2.0)
// https://github.com/magicsong/sonargo/tree/103eda7abc20bd192a064b6eb94ba26329e339f1/sonar
// reduced to the parts used in this package.

// Paging is used in many API responses
type Paging struct {
	PageIndex int `json:"pageIndex,omitempty"`
	PageSize  int `json:"pageSize,omitempty"`
	Total     int `json:"total,omitempty"`
}

// CeTaskOption represents the query options of the ce/task API endpoint
type CeTaskOption struct {
	AdditionalFields string `url:"additionalFields,omitempty"` // Description:"Comma-separated list of the optional fields to be returned in response.",ExampleValue:""
	Id               string `url:"id,omitempty"`               // Description:"Id of task",ExampleValue:"AU-Tpxb--iU5OvuD2FLy"
}

// CeTaskObject represents the response of the ce/task API endpoint
type CeTaskObject struct {
	Task *Task `json:"task,omitempty"`
}

// Task represents a compute engine task
type Task struct {
	AnalysisID         string   `json:"analysisId,omitempty"`
	ComponentID        string   `json:"componentId,omitempty"`
	ComponentKey       string   `json:"componentKey,omitempty"`
	ComponentName      string   `json:"componentName,omitempty"`
	ComponentQualifier string   `json:"componentQualifier,omitempty"`
	ErrorMessage       string   `json:"errorMessage,omitempty"`
	ErrorStacktrace    string   `json:"errorStacktrace,omitempty"`
	ErrorType          string   `json:"errorType,omitempty"`
	ExecutedAt         string   `json:"executedAt,omitempty"`
	ExecutionTimeMs    int64    `json:"executionTimeMs,omitempty"`
	FinishedAt         string   `json:"finishedAt,omitempty"`
	HasErrorStacktrace bool     `json:"hasErrorStacktrace,omitempty"`
	HasScannerContext  bool     `json:"hasScannerContext,omitempty"`
	ID                 string   `json:"id,omitempty"`
	Logs               bool     `json:"logs,omitempty"`
	Organization       string   `json:"organization,omitempty"`
	ScannerContext     string   `json:"scannerContext,omitempty"`
	StartedAt          string   `json:"startedAt,omitempty"`
	Status             string   `json:"status,omitempty"`
	SubmittedAt        string   `json:"submittedAt,omitempty"`
	SubmitterLogin     string   `json:"submitterLogin,omitempty"`
	TaskType           string   `json:"taskType,omitempty"`
	Type               string   `json:"type,omitempty"`
	WarningCount       int      `json:"warningCount,omitempty"`
	Warnings           []string `json:"warnings,omitempty"`
}

// IssuesSearchObject represents the response of the issues/search API endpoint
type IssuesSearchObject struct {
	Components  []*Component `json:"components,omitempty"`
	EffortTotal int          `json:"effortTotal,omitempty"`
	DebtTotal   int          `json:"debtTotal,omitempty"`
	Issues      []*Issue     `json:"issues,omitempty"`
	P           int          `json:"p,omitempty"`
	Ps          int          `json:"ps,omitempty"`
	Paging      *Paging      `json:"paging,omitempty"`
	Total       int          `json:"total,omitempty"`
	Facets      []string     `json:"facets,omitempty"`
}

// Issue represents a SonarQube issue
type Issue struct {
	Actions      []string   `json:"actions,omitempty"`
	Assignee     string     `json:"assignee,omitempty"`
	Author       string     `json:"author,omitempty"`
	Comments     []*Comment `json:"comments,omitempty"`
	Component    string     `json:"component,omitempty"`
	CreationDate string     `json:"creationDate,omitempty"`
	Debt         string     `json:"debt,omitempty"`
	Effort       string     `json:"effort,omitempty"`
	Flows        []any      `json:"flows,omitempty"`
	Hash         string     `json:"hash,omitempty"`
	Key          string     `json:"key,omitempty"`
	Line         int        `json:"line,omitempty"`
	Message      string     `json:"message,omitempty"`
	Organization string     `json:"organization,omitempty"`
	Project      string     `json:"project,omitempty"`
	Rule         string     `json:"rule,omitempty"`
	Severity     string     `json:"severity,omitempty"`
	Status       string     `json:"status,omitempty"`
	Tags         []string   `json:"tags,omitempty"`
	TextRange    *TextRange `json:"textRange,omitempty"`
	Transitions  []string   `json:"transitions,omitempty"`
	Type         string     `json:"type,omitempty"`
	UpdateDate   string     `json:"updateDate,omitempty"`
	FromHotspot  bool       `json:"fromHotspot,omitempty"`
	Resolution   string     `json:"resolution,omitempty"`
	CloseDate    string     `json:"closeDate,omitempty"`
}

// Comment represents a comment on an issue
type Comment struct {
	CreatedAt string `json:"createdAt,omitempty"`
	HTMLText  string `json:"htmlText,omitempty"`
	Key       string `json:"key,omitempty"`
	Login     string `json:"login,omitempty"`
	Markdown  string `json:"markdown,omitempty"`
	Updatable bool   `json:"updatable,omitempty"`
}

// TextRange represents the location of an issue in a file
type TextRange struct {
	EndLine     int `json:"endLine,omitempty"`
	EndOffset   int `json:"endOffset,omitempty"`
	StartLine   int `json:"startLine,omitempty"`
	StartOffset int `json:"startOffset,omitempty"`
}

// Component represents a SonarQube component with its measures
type Component struct {
	AnalysisDate     string          `json:"analysisDate,omitempty"`
	Description      string          `json:"description,omitempty"`
	Enabled          bool            `json:"enabled,omitempty"`
	ID               string          `json:"id,omitempty"`
	Key              string          `json:"key,omitempty"`
	Language         string          `json:"language,omitempty"`
	LastAnalysisDate string          `json:"lastAnalysisDate,omitempty"`
	LeakPeriodDate   string          `json:"leakPeriodDate,omitempty"`
	LongName         string          `json:"longName,omitempty"`
	Measures         []*SonarMeasure `json:"measures,omitempty"`
	Name             string          `json:"name,omitempty"`
	Organization     string          `json:"organization,omitempty"`
	Path             string          `json:"path,omitempty"`
	Project          string          `json:"project,omitempty"`
	Qualifier        string          `json:"qualifier,omitempty"`
	Tags             []string        `json:"tags,omitempty"`
	UUID             string          `json:"uuid,omitempty"`
	Version          string          `json:"version,omitempty"`
	Visibility       string          `json:"visibility,omitempty"`
}

// MeasuresComponentObject represents the response of the measures/component API endpoint
type MeasuresComponentObject struct {
	Component *Component `json:"component,omitempty"`
	Periods   []*Period  `json:"periods,omitempty"`
}

// SonarMeasure represents a single measure of a component
type SonarMeasure struct {
	Metric    string     `json:"metric,omitempty"`
	Periods   []*Period  `json:"periods,omitempty"`
	Value     string     `json:"value,omitempty"`
	Histories []*History `json:"history,omitempty"`
	BestValue bool       `json:"bestValue,omitempty"`
}

// Period represents a new code period
type Period struct {
	Date      string `json:"date,omitempty"`
	Index     int64  `json:"index,omitempty"`
	Mode      string `json:"mode,omitempty"`
	Parameter string `json:"parameter,omitempty"`
	Value     string `json:"value,omitempty"`
	BestValue bool   `json:"bestValue,omitempty"`
}

// History represents a historical measure value
type History struct {
	Date  string `json:"date,omitempty"`
	Value string `json:"value,omitempty"`
}

// ErrorResponse is returned by CheckResponse for API responses with an error status code
type ErrorResponse struct {
	Body     []byte
	Response *http.Response
	Message  string
}

// Error implements the error interface
func (e *ErrorResponse) Error() string {
	path, _ := url.QueryUnescape(e.Response.Request.URL.Path)
	u := fmt.Sprintf("%s://%s%s", e.Response.Request.URL.Scheme, e.Response.Request.URL.Host, path)
	return fmt.Sprintf("%s %s: %d %s", e.Response.Request.Method, u, e.Response.StatusCode, e.Message)
}

// CheckResponse returns an ErrorResponse in case the response has an error status code
func CheckResponse(r *http.Response) error {
	switch r.StatusCode {
	case 200, 201, 202, 204, 304:
		return nil
	}

	errorResponse := &ErrorResponse{Response: r}
	data, err := io.ReadAll(r.Body)
	if err == nil && data != nil {
		errorResponse.Body = data

		var raw any
		if err := json.Unmarshal(data, &raw); err != nil {
			errorResponse.Message = string(data)
		} else {
			errorResponse.Message = parseError(raw)
		}
	}

	return errorResponse
}

func parseError(raw any) string {
	switch raw := raw.(type) {
	case string:
		return raw
	case []any:
		var errs []string
		for _, v := range raw {
			errs = append(errs, parseError(v))
		}
		return fmt.Sprintf("[%s]", strings.Join(errs, ", "))
	case map[string]any:
		var errs []string
		for k, v := range raw {
			errs = append(errs, fmt.Sprintf("{%s: %s}", k, parseError(v)))
		}
		sort.Strings(errs)
		return strings.Join(errs, ", ")
	default:
		return fmt.Sprintf("failed to parse unexpected error type: %T", raw)
	}
}
