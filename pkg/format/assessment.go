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
	List []Assessment `json:"ignore"`
}

type Assessment struct {
	Vulnerability string             `json:"vulnerability"`
	Status        AssessmentStatus   `json:"status"`
	Analysis      AssessmentAnalysis `json:"analysis"`
	Purls         []Purl             `json:"purls"`
}

type AssessmentStatus string

const (
	NotAssessed AssessmentStatus = "notAssessed" //"Not Assessed"
	Relevant    AssessmentStatus = "relevant"    //"Relevant (True Positive)"
	NotRelevant AssessmentStatus = "notRelevant" //"Not Relevant (False Positive)"
	InProcess   AssessmentStatus = "inProcess"   //"In Process"
)

type AssessmentAnalysis string

const (
	WaitingForFix         AssessmentAnalysis = "waitingForFix"         //"Waiting for OSS community fix"
	RiskAccepted          AssessmentAnalysis = "riskAccepted"          //"Risk Accepted"
	NotPresent            AssessmentAnalysis = "notPresent"            //"Affected parts of the OSS library are not present"
	NotUsed               AssessmentAnalysis = "notUsed"               //"Affected parts of the OSS library are not used"
	AssessmentPropagation AssessmentAnalysis = "assessmentPropagation" //"Assessment Propagation"
	FixedByDevTeam        AssessmentAnalysis = "fixedByDevTeam"        //"OSS Component fixed by development team"
	Mitigated             AssessmentAnalysis = "mitigated"             //"Mitigated by the Application"
	WronglyReported       AssessmentAnalysis = "wronglyReported"       //"Wrongly reported CVE"
)

type Purl struct {
	Purl string `json:"purl"`
}

func (p Purl) ToPackageUrl() (packageurl.PackageURL, error) {
	return packageurl.FromString(p.Purl)
}

// ReadAssessment loads the assessments and returns their contents
func ReadAssessments(assessmentFile io.ReadCloser) (*[]Assessment, error) {
	defer assessmentFile.Close()
	assessments := &[]Assessment{}

	content, err := ioutil.ReadAll(assessmentFile)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading %v", assessmentFile)
	}

	err = yaml.Unmarshal(content, assessments)
	if err != nil {
		return nil, NewParseError(fmt.Sprintf("format of assessment file is invalid %q: %v", content, err))
	}
	return assessments, nil
}
