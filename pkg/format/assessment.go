package format

import (
	"fmt"
	"io"
	"io/ioutil"

	"github.com/ghodss/yaml"

	"github.com/pkg/errors"
)

// Assessment format related JSON structs
type Assessment struct {
	Ignores    []Ignore `json:"ignore"`
}

type Ignore struct {
	Vulnerability string `json:"vulnerability"`
	Status AssessmentState `json:"status"`
	Analysis AssessmentAnalysis `json:"analysis"`
	Purls []Purl `json:"purls"`
}

type AssessmentState string

const (
	NotAssessed AssessmentState = "Not Assessed"
	Relevant AssessmentState = "Relevant (True Positive)"
	NotRelevant AssessmentState = "Not Relevant (False Positive)"
	InProcess AssessmentState = "In Process"
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