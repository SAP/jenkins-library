package format

import (
	"fmt"
	"io"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/ghodss/yaml"
	"github.com/package-url/packageurl-go"
	"github.com/pkg/errors"
)

type Assessment struct {
	Vulnerability string             `json:"vulnerability"`
	Status        AssessmentStatus   `json:"status"`
	Analysis      AssessmentAnalysis `json:"analysis"`
	Purls         []Purl             `json:"purls"`
}

type AssessmentStatus string

const (
	//NotAssessed AssessmentStatus = "notAssessed" //"Not Assessed"
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

func (a Assessment) ToImpactAnalysisState() cdx.ImpactAnalysisState {
	switch a.Status {
	case Relevant:
		return cdx.IASExploitable
	case NotRelevant:
		return cdx.IASFalsePositive
	case InProcess:
		return cdx.IASInTriage
	}
	return cdx.IASExploitable
}

func (a Assessment) ToImpactJustification() cdx.ImpactAnalysisJustification {
	switch a.Analysis {
	case WaitingForFix:
		return cdx.IAJRequiresDependency
	case RiskAccepted:
		return cdx.IAJRequiresEnvironment
	case NotPresent:
		return cdx.IAJCodeNotPresent
	case NotUsed:
		return cdx.IAJCodeNotReachable
	case AssessmentPropagation:
		return cdx.IAJRequiresDependency
	case FixedByDevTeam:
		return cdx.IAJProtectedByMitigatingControl
	case Mitigated:
		return cdx.IAJProtectedByMitigatingControl
	case WronglyReported:
		return cdx.IAJCodeNotPresent
	}
	return cdx.IAJProtectedAtRuntime
}

func (a Assessment) ToImpactAnalysisResponse() *[]cdx.ImpactAnalysisResponse {
	switch a.Analysis {
	case WaitingForFix:
		return &[]cdx.ImpactAnalysisResponse{cdx.IARCanNotFix}
	case RiskAccepted:
		return &[]cdx.ImpactAnalysisResponse{cdx.IARWillNotFix}
	case NotPresent:
		return &[]cdx.ImpactAnalysisResponse{cdx.IARCanNotFix}
	case NotUsed:
		return &[]cdx.ImpactAnalysisResponse{cdx.IARWillNotFix}
	case AssessmentPropagation:
		return &[]cdx.ImpactAnalysisResponse{cdx.IARCanNotFix}
	case FixedByDevTeam:
		return &[]cdx.ImpactAnalysisResponse{cdx.IARUpdate}
	case Mitigated:
		return &[]cdx.ImpactAnalysisResponse{cdx.IARWorkaroundAvailable}
	case WronglyReported:
		return &[]cdx.ImpactAnalysisResponse{cdx.IARCanNotFix}
	}
	return &[]cdx.ImpactAnalysisResponse{cdx.IARWillNotFix}
}

// ReadAssessment loads the assessments and returns their contents
func ReadAssessments(assessmentFile io.ReadCloser) (*[]Assessment, error) {
	defer assessmentFile.Close()
	ignore := struct {
		Assessments []Assessment `json:"ignore"`
	}{
		Assessments: []Assessment{},
	}

	content, err := io.ReadAll(assessmentFile)
	if err != nil {
		return nil, errors.Wrapf(err, "error reading %v", assessmentFile)
	}

	err = yaml.Unmarshal(content, &ignore)
	if err != nil {
		return nil, NewParseError(fmt.Sprintf("format of assessment file is invalid %q: %v", content, err))
	}
	return &ignore.Assessments, nil
}
