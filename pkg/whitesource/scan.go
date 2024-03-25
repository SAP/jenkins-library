package whitesource

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/versioning"
)

// Scan stores information about scanned WhiteSource projects (modules).
type Scan struct {
	// AggregateProjectName stores the name of the WhiteSource project where scans shall be aggregated.
	// It does not include the ProductVersion.
	AggregateProjectName string
	// ProductVersion is the global version that is used across all Projects (modules) during the scan.
	BuildTool                   string
	ProductToken                string
	ProductVersion              string
	scannedProjects             map[string]Project
	scanTimes                   map[string]time.Time
	AgentName                   string
	AgentVersion                string
	Coordinates                 versioning.Coordinates
	SkipProjectsWithEmptyTokens bool
}

func (s *Scan) init() {
	if s.scannedProjects == nil {
		s.scannedProjects = make(map[string]Project)
	}
	if s.scanTimes == nil {
		s.scanTimes = make(map[string]time.Time)
	}
}

func (s *Scan) versionSuffix() string {
	return " - " + s.ProductVersion
}

// AppendScannedProject checks that no Project with the same name is already contained in the list of scanned projects,
// and appends a new Project with the given name. The global product version is appended to the name.
func (s *Scan) AppendScannedProject(projectName string) error {
	if len(projectName) == 0 {
		return fmt.Errorf("projectName must not be empty")
	}
	if strings.HasSuffix(projectName, s.versionSuffix()) {
		return fmt.Errorf("projectName is not expected to include the product version already")
	}
	return s.AppendScannedProjectVersion(projectName + s.versionSuffix())
}

// AppendScannedProjectVersion checks that no Project with the same name is already contained in the list of scanned
// projects,  and appends a new Project with the given name (which is expected to include the product version).
func (s *Scan) AppendScannedProjectVersion(projectName string) error {
	if !strings.HasSuffix(projectName, s.versionSuffix()) {
		return fmt.Errorf("projectName is expected to include the product version")
	}
	if len(projectName) == len(s.versionSuffix()) {
		return fmt.Errorf("projectName consists only of the product version")
	}
	s.init()
	_, exists := s.scannedProjects[projectName]
	if exists {
		log.Entry().Errorf("A module with the name '%s' was already scanned. "+
			"Your project's modules must have unique names.", projectName)
		return fmt.Errorf("project with name '%s' was already scanned", projectName)
	}
	s.scannedProjects[projectName] = Project{Name: projectName}
	s.scanTimes[projectName] = time.Now()
	return nil
}

// ProjectByName returns a WhiteSource Project previously established via AppendScannedProject().
func (s *Scan) ProjectByName(projectName string) (Project, bool) {
	project, exists := s.scannedProjects[projectName]
	return project, exists
}

// ScannedProjects returns the WhiteSource projects that have been added via AppendScannedProject() as a slice.
func (s *Scan) ScannedProjects() []Project {
	var projects []Project
	for _, project := range s.scannedProjects {
		if len(project.Token) == 0 && s.SkipProjectsWithEmptyTokens {
			log.Entry().Debugf("Project will be skipped as the token is empty, project_name: %s", project.Name)
			continue
		}
		projects = append(projects, project)
	}
	return projects
}

// ScannedProjectNames returns a sorted list of all scanned project names
func (s *Scan) ScannedProjectNames() []string {
	projectNames := []string{}
	for _, project := range s.ScannedProjects() {
		projectNames = append(projectNames, project.Name)
	}
	// Sorting helps the list become stable across pipeline runs (and in the unit tests),
	// as the order in which we travers map keys is not deterministic.
	sort.Strings(projectNames)
	return projectNames
}

// ScannedProjectTokens returns a sorted list of all scanned project's tokens
func (s *Scan) ScannedProjectTokens() []string {
	projectTokens := []string{}
	for _, project := range s.ScannedProjects() {
		if len(project.Token) > 0 {
			projectTokens = append(projectTokens, project.Token)
		}
	}
	// Sorting helps the list become stable across pipeline runs (and in the unit tests),
	// as the order in which we travers map keys is not deterministic.
	sort.Strings(projectTokens)
	return projectTokens
}

// ScanTime returns the time at which the respective WhiteSource Project was scanned, or the the
// zero value of time.Time, if AppendScannedProject() was not called with that name.
func (s *Scan) ScanTime(projectName string) time.Time {
	if s.scanTimes == nil {
		return time.Time{}
	}
	return s.scanTimes[projectName]
}

type whitesource interface {
	GetProjectsMetaInfo(productToken string) ([]Project, error)
	GetProjectRiskReport(projectToken string) ([]byte, error)
	GetProjectVulnerabilityReport(projectToken string, format string) ([]byte, error)
}

// UpdateProjects pulls the current backend metadata for all WhiteSource projects in the product with
// the given productToken, and updates all scanned projects with the obtained information.
func (s *Scan) UpdateProjects(productToken string, sys whitesource) error {
	s.init()
	projects, err := sys.GetProjectsMetaInfo(productToken)
	if err != nil {
		return fmt.Errorf("failed to retrieve WhiteSource projects meta info: %w", err)
	}

	var projectsToUpdate []string
	for projectName := range s.scannedProjects {
		projectsToUpdate = append(projectsToUpdate, projectName)
	}

	for _, project := range projects {
		_, exists := s.scannedProjects[project.Name]
		if exists {
			s.scannedProjects[project.Name] = project
			projectsToUpdate, _ = piperutils.RemoveAll(projectsToUpdate, project.Name)
		}
	}
	if len(projectsToUpdate) != 0 {
		log.Entry().Warnf("Could not fetch metadata for projects %v", projectsToUpdate)
	}
	return nil
}
