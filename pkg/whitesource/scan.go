package whitesource

import (
	"fmt"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"strings"
	"time"
)

// Scan stores information about scanned WhiteSource projects (modules).
type Scan struct {
	// AggregateProjectName stores the name of the WhiteSource project where scans shall be aggregated.
	// It does not include the ProductVersion.
	AggregateProjectName string
	// ProductVersion is the global version that is used across all Projects (modules) during the scan.
	ProductVersion  string
	scannedProjects map[string]Project
	scanTimes       map[string]time.Time
}

func (s *Scan) init() {
	if s.scannedProjects == nil {
		s.scannedProjects = make(map[string]Project)
	}
	if s.scanTimes == nil {
		s.scanTimes = make(map[string]time.Time)
	}
}

// AppendScannedProject checks that no Project with the same name is already contained in the list of scanned projects,
// and appends a new Project with the given name. The global product version is appended to the name.
func (s *Scan) AppendScannedProject(projectName string) error {
	return s.AppendScannedProjectVersion(projectName + " - " + s.ProductVersion)
}

// AppendScannedProjectVersion checks that no Project with the same name is already contained in the list of scanned
// projects,  and appends a new Project with the given name (which is expected to include the product version).
func (s *Scan) AppendScannedProjectVersion(projectName string) error {
	if !strings.HasSuffix(projectName, " - "+s.ProductVersion) {
		return fmt.Errorf("projectName is expected to include the product version")
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
		projects = append(projects, project)
	}
	return projects
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
