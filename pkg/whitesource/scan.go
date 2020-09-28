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
	aggregateProjectName string
	productVersion       string
	scannedProjects      map[string]Project
	scanTimes            map[string]time.Time
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
	return s.AppendScannedProjectVersion(projectName + " - " + s.productVersion)
}

// AppendScannedProjectVersion checks that no Project with the same name is already contained in the list of scanned
// projects,  and appends a new Project with the given name (which is expected to include the product version).
func (s *Scan) AppendScannedProjectVersion(projectName string) error {
	if !strings.HasSuffix(projectName, " - "+s.productVersion) {
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

type whitesource interface {
	GetProjectsMetaInfo(productToken string) ([]Project, error)
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

// ScanOptions describes configuration option for a WhiteSource Scan.
type ScanOptions struct {
	// AggregateProjectName controls whether the scan results for all modules are aggregated into the
	// same WhiteSource Project, or (if left empty) are each stored in their individual Project.
	AggregateProjectName string
	// ProductVersion is the global version that is used across all Projects (modules) during the scan.
	ProductVersion string
}

// NewScan returns a new Scan instance, configured from the provided ScanOptions.
func NewScan(config ScanOptions) *Scan {
	return &Scan{
		aggregateProjectName: config.AggregateProjectName,
		productVersion:       config.ProductVersion,
	}
}
