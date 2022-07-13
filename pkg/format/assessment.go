package format

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/ghodss/yaml"
	"github.com/package-url/packageurl-go"
	"github.com/pkg/errors"
)

// Assessment format related JSON structs
type Assessments struct {
	List    []Assessment `json:"ignore"`
}

type Assessment struct {
	Vulnerability string `json:"vulnerability"`
	Status AssessmentStatus `json:"status"`
	Analysis AssessmentAnalysis `json:"analysis"`
	Purls []Purl `json:"purls"`
}

type AssessmentStatus string

const (
	NotAssessed AssessmentStatus = "Not Assessed"
	Relevant AssessmentStatus = "Relevant (True Positive)"
	NotRelevant AssessmentStatus = "Not Relevant (False Positive)"
	InProcess AssessmentStatus = "In Process"
)

type AssessmentAnalysis string

const (
	WaitingForFix AssessmentAnalysis = "Waiting for OSS community fix"
	RiskAccepted AssessmentAnalysis = "Risk Accepted"
	Others AssessmentAnalysis = "Others"
	NotPresent AssessmentAnalysis = "Affected parts of the OSS library are not present"
	NotUsed AssessmentAnalysis = "Affected parts of the OSS library are not used"
	AssessmentPropagation AssessmentAnalysis = "Assessment Propagation"
	BuildVersionOutdated AssessmentAnalysis = "Build Version is outdated"
	FixedByDevTeam AssessmentAnalysis = "OSS Component fixed by development team"
	Mitigated AssessmentAnalysis = "Mitigated by the Application"
	WronglyReported AssessmentAnalysis = "Wrongly reported CVE"
)

type Purl struct {
	Purl string `json:"purl"`
}

func (p Purl) ToPackageUrl() (packageurl.PackageURL, error) {
	return packageurl.FromString(p.Purl)
}

// ReadAssessment loads the assessments and returns their contents
func (assessment *Assessment) ReadAssessment(assessmentFile io.ReadCloser) error {
	defer assessmentFile.Close()

	content, err := ioutil.ReadAll(assessmentFile)
	if err != nil {
		return errors.Wrapf(err, "error reading %v", assessmentFile)
	}

	err = yaml.Unmarshal(content, &assessment)
	if err != nil {
		return NewParseError(fmt.Sprintf("format of assessment file is invalid %q: %v", content, err))
	}
	return nil
}