package versioning

import (
	"path/filepath"
	"strings"
	"unicode"

	"github.com/Masterminds/semver/v3"
	"helm.sh/helm/v3/pkg/chart"
)

// Maintainer describes a Chart maintainer.
type Maintainer struct {
	// Name is a user name or organization name
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Email is an optional email address to contact the named maintainer
	Email string `json:"email,omitempty" yaml:"email,omitempty"`
	// URL is an optional URL to an address for the named maintainer
	URL string `json:"url,omitempty" yaml:"url,omitempty"`
}

// Validate checks valid data and sanitizes string characters.
func (m *Maintainer) Validate() error {
	if m == nil {
		return chart.ValidationError("maintainers must not contain empty or null nodes")
	}
	m.Name = sanitizeString(m.Name)
	m.Email = sanitizeString(m.Email)
	m.URL = sanitizeString(m.URL)
	return nil
}

// Metadata for a Chart file. This models the structure of a Chart.yaml file.
type Metadata struct {
	// The name of the chart. Required.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// The URL to a relevant project page, git repo, or contact person
	Home string `json:"home,omitempty" yaml:"home,omitempty"`
	// Source is the URL to the source code of this chart
	Sources []string `json:"sources,omitempty" yaml:"sources,omitempty"`
	// A SemVer 2 conformant version string of the chart. Required.
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
	// A one-sentence description of the chart
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	// A list of string keywords
	Keywords []string `json:"keywords,omitempty" yaml:"keywords,omitempty"`
	// A list of name and URL/email address combinations for the maintainer(s)
	Maintainers []*Maintainer `json:"maintainers,omitempty" yaml:"maintainers,omitempty"`
	// The URL to an icon file.
	Icon string `json:"icon,omitempty" yaml:"icon,omitempty"`
	// The API Version of this chart. Required.
	APIVersion string `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	// The condition to check to enable chart
	Condition string `json:"condition,omitempty" yaml:"condition,omitempty"`
	// The tags to check to enable chart
	Tags string `json:"tags,omitempty" yaml:"tags,omitempty"`
	// The version of the application enclosed inside of this chart.
	AppVersion string `json:"appVersion,omitempty" yaml:"appVersion,omitempty"`
	// Whether or not this chart is deprecated
	Deprecated bool `json:"deprecated,omitempty" yaml:"deprecated,omitempty"`
	// Annotations are additional mappings uninterpreted by Helm,
	// made available for inspection by other applications.
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	// KubeVersion is a SemVer constraint specifying the version of Kubernetes required.
	KubeVersion string `json:"kubeVersion,omitempty" yaml:"kubeVersion,omitempty"`
	// Dependencies are a list of dependencies for a chart.
	Dependencies []*chart.Dependency `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
	// Specifies the chart type: application or library
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
}

// Validate checks the metadata for known issues and sanitizes string
// characters.
func (md *Metadata) Validate() error {
	if md == nil {
		return chart.ValidationError("chart.metadata is required")
	}

	md.Name = sanitizeString(md.Name)
	md.Description = sanitizeString(md.Description)
	md.Home = sanitizeString(md.Home)
	md.Icon = sanitizeString(md.Icon)
	md.Condition = sanitizeString(md.Condition)
	md.Tags = sanitizeString(md.Tags)
	md.AppVersion = sanitizeString(md.AppVersion)
	md.KubeVersion = sanitizeString(md.KubeVersion)
	for i := range md.Sources {
		md.Sources[i] = sanitizeString(md.Sources[i])
	}
	for i := range md.Keywords {
		md.Keywords[i] = sanitizeString(md.Keywords[i])
	}

	if md.APIVersion == "" {
		return chart.ValidationError("chart.metadata.apiVersion is required")
	}
	if md.Name == "" {
		return chart.ValidationError("chart.metadata.name is required")
	}

	if md.Name != filepath.Base(md.Name) {
		return chart.ValidationErrorf("chart.metadata.name %q is invalid", md.Name)
	}

	if md.Version == "" {
		return chart.ValidationError("chart.metadata.version is required")
	}
	if !isValidSemver(md.Version) {
		return chart.ValidationErrorf("chart.metadata.version %q is invalid", md.Version)
	}
	if !isValidChartType(md.Type) {
		return chart.ValidationError("chart.metadata.type must be application or library")
	}

	for _, m := range md.Maintainers {
		if err := m.Validate(); err != nil {
			return err
		}
	}

	// Aliases need to be validated here to make sure that the alias name does
	// not contain any illegal characters.
	dependencies := map[string]*chart.Dependency{}
	for _, dependency := range md.Dependencies {
		if err := dependency.Validate(); err != nil {
			return err
		}
		key := dependency.Name
		if dependency.Alias != "" {
			key = dependency.Alias
		}
		if dependencies[key] != nil {
			return chart.ValidationErrorf("more than one dependency with name or alias %q", key)
		}
		dependencies[key] = dependency
	}
	return nil
}

func isValidChartType(in string) bool {
	switch in {
	case "", "application", "library":
		return true
	}
	return false
}

func isValidSemver(v string) bool {
	_, err := semver.NewVersion(v)
	return err == nil
}

// sanitizeString normalize spaces and removes non-printable characters.
func sanitizeString(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return ' '
		}
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}, str)
}
