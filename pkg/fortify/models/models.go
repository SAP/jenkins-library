// Package models provides the data model of the Fortify SSC REST API as
// used by the fortifyExecuteScan step. The type definitions were derived from
// the go-swagger generated models of github.com/piper-validation/fortify-client-go
// (commit 7b3e9a72af01), reduced to plain data structs.
package models

type Artifact struct {
	Embed             *EmbeddedScans       `json:"_embed,omitempty"`
	AllowApprove      bool                 `json:"allowApprove,omitempty"`
	AllowDelete       bool                 `json:"allowDelete,omitempty"`
	AllowPurge        bool                 `json:"allowPurge,omitempty"`
	ApprovalComment   string               `json:"approvalComment,omitempty"`
	ApprovalDate      Iso8601MilliDateTime `json:"approvalDate,omitempty"`
	ApprovalUsername  string               `json:"approvalUsername,omitempty"`
	ArtifactType      string               `json:"artifactType,omitempty"`
	AuditUpdated      bool                 `json:"auditUpdated,omitempty"`
	FileName          string               `json:"fileName,omitempty"`
	FileSize          int64                `json:"fileSize,omitempty"`
	FileURL           string               `json:"fileURL,omitempty"`
	ID                int64                `json:"id,omitempty"`
	InModifyingStatus bool                 `json:"inModifyingStatus,omitempty"`
	Indexed           bool                 `json:"indexed,omitempty"`
	LastScanDate      Iso8601MilliDateTime `json:"lastScanDate,omitempty"`
	MessageCount      int64                `json:"messageCount,omitempty"`
	Messages          string               `json:"messages,omitempty"`
	OriginalFileName  string               `json:"originalFileName,omitempty"`
	OtherStatus       string               `json:"otherStatus,omitempty"`
	Purged            bool                 `json:"purged,omitempty"`
	RuntimeStatus     string               `json:"runtimeStatus,omitempty"`
	ScaStatus         string               `json:"scaStatus,omitempty"`
	ScanErrorsCount   int64                `json:"scanErrorsCount,omitempty"`
	Status            string               `json:"status,omitempty"`
	UploadDate        Iso8601MilliDateTime `json:"uploadDate,omitempty"`
	UploadIP          string               `json:"uploadIP,omitempty"`
	UserName          string               `json:"userName,omitempty"`
	VersionNumber     int32                `json:"versionNumber,omitempty"`
	WebInspectStatus  string               `json:"webInspectStatus,omitempty"`
}

type Attribute struct {
	AttributeDefinitionID *int64             `json:"attributeDefinitionId"`
	ID                    int64              `json:"id,omitempty"`
	Value                 *string            `json:"value"`
	Values                []*AttributeOption `json:"values"`
}

type AttributeOption struct {
	Description          string  `json:"description,omitempty"`
	GUID                 *string `json:"guid"`
	Hidden               bool    `json:"hidden,omitempty"`
	ID                   *int64  `json:"id"`
	InUse                bool    `json:"inUse,omitempty"`
	Index                *int32  `json:"index"`
	Name                 *string `json:"name"`
	ObjectVersion        int32   `json:"objectVersion,omitempty"`
	ProjectMetaDataDefID *int64  `json:"projectMetaDataDefId"`
	PublishVersion       int32   `json:"publishVersion,omitempty"`
}

type AuthenticationEntity struct {
	Embed       *EmbeddedRoles `json:"_embed,omitempty"`
	DisplayName string         `json:"displayName,omitempty"`
	Email       string         `json:"email,omitempty"`
	EntityName  string         `json:"entityName,omitempty"`
	FirstName   string         `json:"firstName,omitempty"`
	ID          int64          `json:"id,omitempty"`
	IsLdap      bool           `json:"isLdap,omitempty"`
	LastName    string         `json:"lastName,omitempty"`
	LdapDn      string         `json:"ldapDn,omitempty"`
	Type        string         `json:"type,omitempty"`
	UserPhoto   *UserPhoto     `json:"userPhoto,omitempty"`
}

type EmbeddedReportDefinition struct {
	FieldsToNullWithExclusions []string          `json:"fieldsToNullWithExclusions"`
	ReportDefinition           *ReportDefinition `json:"reportDefinition,omitempty"`
}

type EmbeddedRoles struct {
	Roles []*Role `json:"roles"`
}

type EmbeddedScans struct {
	Scans []*Scan `json:"scans"`
}

type EngineType struct {
	Name           string `json:"name,omitempty"`
	ServedByPlugin *bool  `json:"servedByPlugin,omitempty"`
}

type FileToken struct {
	FileTokenType *string `json:"fileTokenType"`
	Token         string  `json:"token,omitempty"`
}

type FilterSet struct {
	DefaultFilterSet bool         `json:"defaultFilterSet"`
	Description      string       `json:"description"`
	Folders          []*FolderDto `json:"folders"`
	GUID             string       `json:"guid"`
	Title            string       `json:"title"`
}

type FolderDto struct {
	Color string `json:"color,omitempty"`
	GUID  string `json:"guid,omitempty"`
	ID    int64  `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
}

type InputReportParameter struct {
	Identifier *string     `json:"identifier"`
	Name       *string     `json:"name"`
	ParamValue interface{} `json:"paramValue"`
	Type       *string     `json:"type"`
}

type IssueAuditComment struct {
	AuditTime          Iso8601MilliDateTime `json:"auditTime"`
	Comment            *string              `json:"comment"`
	IssueEngineType    string               `json:"issueEngineType,omitempty"`
	IssueID            int64                `json:"issueId"`
	IssueInstanceID    string               `json:"issueInstanceId,omitempty"`
	IssueName          string               `json:"issueName,omitempty"`
	ProjectName        string               `json:"projectName,omitempty"`
	ProjectVersionID   int64                `json:"projectVersionId,omitempty"`
	ProjectVersionName string               `json:"projectVersionName,omitempty"`
	SeqNumber          int32                `json:"seqNumber"`
	UserName           string               `json:"userName"`
}

type IssueFilterSelector struct {
	Description        string            `json:"description"`
	DisplayName        string            `json:"displayName"`
	EntityType         string            `json:"entityType"`
	FilterSelectorType string            `json:"filterSelectorType"`
	GUID               string            `json:"guid"`
	SelectorOptions    []*SelectorOption `json:"selectorOptions"`
	Value              string            `json:"value"`
}

type IssueFilterSelectorSet struct {
	FilterBySet []*IssueFilterSelector `json:"filterBySet"`
	GroupBySet  []*IssueSelector       `json:"groupBySet"`
}

type IssueSelector struct {
	Description *string `json:"description"`
	DisplayName *string `json:"displayName"`
	EntityType  *string `json:"entityType"`
	GUID        *string `json:"guid"`
	Value       *string `json:"value"`
}

type IssueStatistics struct {
	FilterSetID                *int64 `json:"filterSetId"`
	HiddenCount                *int32 `json:"hiddenCount"`
	HiddenDisplayableCount     int32  `json:"hiddenDisplayableCount"`
	ProjectVersionID           *int64 `json:"projectVersionId"`
	RemovedCount               *int32 `json:"removedCount"`
	RemovedDisplayableCount    int32  `json:"removedDisplayableCount,omitempty"`
	SuppressedCount            *int32 `json:"suppressedCount"`
	SuppressedDisplayableCount int32  `json:"suppressedDisplayableCount,omitempty"`
}

type Project struct {
	CreatedBy       *string              `json:"createdBy"`
	CreationDate    Iso8601MilliDateTime `json:"creationDate"`
	Description     string               `json:"description,omitempty"`
	ID              int64                `json:"id,omitempty"`
	IssueTemplateID *string              `json:"issueTemplateId"`
	Name            *string              `json:"name"`
}

type ProjectVersion struct {
	Active                    *bool                 `json:"active"`
	AssignedIssuesCount       int64                 `json:"assignedIssuesCount,omitempty"`
	AttachmentsOutOfDate      bool                  `json:"attachmentsOutOfDate,omitempty"`
	AutoPredict               bool                  `json:"autoPredict,omitempty"`
	BugTrackerEnabled         *bool                 `json:"bugTrackerEnabled"`
	BugTrackerPluginID        *string               `json:"bugTrackerPluginId"`
	Committed                 *bool                 `json:"committed"`
	CreatedBy                 *string               `json:"createdBy"`
	CreationDate              *Iso8601MilliDateTime `json:"creationDate"`
	CurrentState              *ProjectVersionState  `json:"currentState,omitempty"`
	CustomTagValuesAutoApply  bool                  `json:"customTagValuesAutoApply,omitempty"`
	Description               *string               `json:"description"`
	ID                        int64                 `json:"id,omitempty"`
	IssueTemplateID           *string               `json:"issueTemplateId"`
	IssueTemplateModifiedTime *int64                `json:"issueTemplateModifiedTime"`
	IssueTemplateName         *string               `json:"issueTemplateName"`
	LatestScanID              *int64                `json:"latestScanId"`
	LoadProperties            string                `json:"loadProperties,omitempty"`
	MasterAttrGUID            *string               `json:"masterAttrGuid"`
	MigrationVersion          float32               `json:"migrationVersion,omitempty"`
	Mode                      string                `json:"mode,omitempty"`
	Name                      *string               `json:"name"`
	ObfuscatedID              string                `json:"obfuscatedId,omitempty"`
	Owner                     *string               `json:"owner"`
	PredictionPolicy          string                `json:"predictionPolicy,omitempty"`
	Project                   *Project              `json:"project,omitempty"`
	RefreshRequired           bool                  `json:"refreshRequired,omitempty"`
	SecurityGroup             string                `json:"securityGroup,omitempty"`
	ServerVersion             *float32              `json:"serverVersion"`
	SiteID                    string                `json:"siteId,omitempty"`
	SnapshotOutOfDate         *bool                 `json:"snapshotOutOfDate"`
	SourceBasePath            string                `json:"sourceBasePath,omitempty"`
	StaleIssueTemplate        *bool                 `json:"staleIssueTemplate"`
	Status                    string                `json:"status,omitempty"`
	TracesOutOfDate           bool                  `json:"tracesOutOfDate,omitempty"`
}

type ProjectVersionCopyCurrentStateRequest struct {
	PreviousProjectVersionID *int64 `json:"previousProjectVersionId"`
	ProjectVersionID         *int64 `json:"projectVersionId"`
}

type ProjectVersionCopyPartialRequest struct {
	CopyAnalysisProcessingRules *bool  `json:"copyAnalysisProcessingRules"`
	CopyBugTrackerConfiguration *bool  `json:"copyBugTrackerConfiguration"`
	CopyCustomTags              *bool  `json:"copyCustomTags"`
	PreviousProjectVersionID    *int64 `json:"previousProjectVersionId"`
	ProjectVersionID            *int64 `json:"projectVersionId"`
}

type ProjectVersionIssue struct {
	Analyzer                   *string               `json:"analyzer"`
	Audited                    bool                  `json:"audited,omitempty"`
	BugURL                     *string               `json:"bugURL"`
	Confidence                 *float32              `json:"confidence"`
	DisplayEngineType          *string               `json:"displayEngineType"`
	EngineCategory             *string               `json:"engineCategory"`
	EngineType                 *string               `json:"engineType"`
	ExternalBugID              *string               `json:"externalBugId"`
	FolderGUID                 *string               `json:"folderGuid"`
	FolderID                   *int64                `json:"folderId"`
	FoundDate                  *Iso8601MilliDateTime `json:"foundDate"`
	Friority                   *string               `json:"friority"`
	FullFileName               *string               `json:"fullFileName"`
	HasAttachments             *bool                 `json:"hasAttachments"`
	HasCorrelatedIssues        *bool                 `json:"hasCorrelatedIssues"`
	Hidden                     *bool                 `json:"hidden"`
	ID                         int64                 `json:"id,omitempty"`
	Impact                     *float32              `json:"impact"`
	IssueInstanceID            *string               `json:"issueInstanceId"`
	IssueName                  *string               `json:"issueName"`
	IssueStatus                string                `json:"issueStatus,omitempty"`
	Kingdom                    *string               `json:"kingdom"`
	LastScanID                 int64                 `json:"lastScanId,omitempty"`
	Likelihood                 *float32              `json:"likelihood"`
	LineNumber                 *int32                `json:"lineNumber"`
	PrimaryLocation            *string               `json:"primaryLocation"`
	PrimaryRuleGUID            *string               `json:"primaryRuleGuid"`
	PrimaryTag                 *string               `json:"primaryTag"`
	PrimaryTagValueAutoApplied bool                  `json:"primaryTagValueAutoApplied,omitempty"`
	ProjectName                *string               `json:"projectName"`
	ProjectVersionID           *int64                `json:"projectVersionId"`
	ProjectVersionName         *string               `json:"projectVersionName"`
	Removed                    *bool                 `json:"removed"`
	RemovedDate                *Iso8601MilliDateTime `json:"removedDate"`
	Reviewed                   *string               `json:"reviewed"`
	Revision                   *int32                `json:"revision"`
	ScanStatus                 *string               `json:"scanStatus"`
	Severity                   *float32              `json:"severity"`
	Suppressed                 *bool                 `json:"suppressed"`
	HasComments                *bool                 `json:"hasComments"`
}

type ProjectVersionIssueGroup struct {
	AuditedCount *int32  `json:"auditedCount"`
	CleanName    *string `json:"cleanName"`
	ID           *string `json:"id"`
	Name         *string `json:"name"`
	TotalCount   *int32  `json:"totalCount"`
	VisibleCount *int32  `json:"visibleCount"`
}

type ProjectVersionState struct {
	AnalysisResultsExist                      *bool                 `json:"analysisResultsExist"`
	AnalysisUploadEnabled                     *bool                 `json:"analysisUploadEnabled"`
	AttentionRequired                         *bool                 `json:"attentionRequired"`
	AuditEnabled                              *bool                 `json:"auditEnabled"`
	BatchBugSubmissionExists                  *bool                 `json:"batchBugSubmissionExists"`
	Committed                                 *bool                 `json:"committed"`
	CriticalPriorityIssueCountDelta           *int32                `json:"criticalPriorityIssueCountDelta"`
	DeltaPeriod                               *int32                `json:"deltaPeriod"`
	ExtraMessage                              *string               `json:"extraMessage"`
	HasCustomIssues                           *bool                 `json:"hasCustomIssues"`
	ID                                        *int64                `json:"id"`
	IssueCountDelta                           *int32                `json:"issueCountDelta"`
	LastFprUploadDate                         *Iso8601MilliDateTime `json:"lastFprUploadDate"`
	MetricEvaluationDate                      *Iso8601MilliDateTime `json:"metricEvaluationDate"`
	PercentAuditedDelta                       *float32              `json:"percentAuditedDelta"`
	PercentCriticalPriorityIssuesAuditedDelta *float32              `json:"percentCriticalPriorityIssuesAuditedDelta"`
}

type ReportAuthEntity struct {
	FirstName string  `json:"firstName,omitempty"`
	ID        int64   `json:"id,omitempty"`
	LastName  string  `json:"lastName,omitempty"`
	UserName  *string `json:"userName"`
}

type ReportDefinition struct {
	CrossApp        bool               `json:"crossApp,omitempty"`
	Description     string             `json:"description,omitempty"`
	FileName        string             `json:"fileName,omitempty"`
	GUID            string             `json:"guid,omitempty"`
	ID              int64              `json:"id,omitempty"`
	InUse           bool               `json:"inUse,omitempty"`
	Name            *string            `json:"name"`
	ObjectVersion   int32              `json:"objectVersion,omitempty"`
	Parameters      []*ReportParameter `json:"parameters"`
	PublishVersion  int32              `json:"publishVersion,omitempty"`
	RenderingEngine string             `json:"renderingEngine,omitempty"`
	TemplateDocID   int64              `json:"templateDocId,omitempty"`
	Type            *string            `json:"type"`
	TypeDefaultText string             `json:"typeDefaultText,omitempty"`
}

type ReportParameter struct {
	Description            string                   `json:"description,omitempty"`
	ID                     int64                    `json:"id,omitempty"`
	Identifier             *string                  `json:"identifier"`
	Name                   string                   `json:"name,omitempty"`
	ParamOrder             int32                    `json:"paramOrder,omitempty"`
	ReportDefinitionID     *int64                   `json:"reportDefinitionId"`
	ReportParameterOptions []*ReportParameterOption `json:"reportParameterOptions"`
	Type                   *string                  `json:"type"`
}

type ReportParameterOption struct {
	DefaultValue bool    `json:"defaultValue,omitempty"`
	Description  string  `json:"description,omitempty"`
	DisplayValue *string `json:"displayValue"`
	ID           int64   `json:"id,omitempty"`
	Order        int32   `json:"order,omitempty"`
	ReportValue  *string `json:"reportValue"`
}

type ReportProject struct {
	ID                   int64                   `json:"id,omitempty"`
	Name                 string                  `json:"name,omitempty"`
	ProjectVersionsCount int32                   `json:"projectVersionsCount,omitempty"`
	Versions             []*ReportProjectVersion `json:"versions"`
}

type ReportProjectVersion struct {
	DevelopmentPhase *string `json:"developmentPhase"`
	ID               int64   `json:"id,omitempty"`
	Name             string  `json:"name,omitempty"`
}

type Role struct {
	AllApplicationRole *bool    `json:"allApplicationRole"`
	AssignedToNonUsers *bool    `json:"assignedToNonUsers"`
	BuiltIn            *bool    `json:"builtIn"`
	Default            bool     `json:"default,omitempty"`
	Deletable          bool     `json:"deletable,omitempty"`
	Description        string   `json:"description,omitempty"`
	ID                 string   `json:"id,omitempty"`
	Name               string   `json:"name,omitempty"`
	ObjectVersion      int32    `json:"objectVersion,omitempty"`
	PermissionIds      []string `json:"permissionIds"`
	PublishVersion     int32    `json:"publishVersion,omitempty"`
	UserOnly           *bool    `json:"userOnly"`
}

type SavedReport struct {
	Embed                 *EmbeddedReportDefinition `json:"_embed,omitempty"`
	AuthEntity            *ReportAuthEntity         `json:"authEntity,omitempty"`
	Format                *string                   `json:"format"`
	FormatDefaultText     string                    `json:"formatDefaultText,omitempty"`
	GenerationDate        Iso8601MilliDateTime      `json:"generationDate,omitempty"`
	ID                    int64                     `json:"id,omitempty"`
	InputReportParameters []*InputReportParameter   `json:"inputReportParameters"`
	IsPublished           bool                      `json:"isPublished,omitempty"`
	Name                  string                    `json:"name,omitempty"`
	Note                  string                    `json:"note,omitempty"`
	Projects              []*ReportProject          `json:"projects"`
	Published             bool                      `json:"published,omitempty"`
	ReportDefinitionID    *int64                    `json:"reportDefinitionId"`
	ReportProjectsCount   int32                     `json:"reportProjectsCount,omitempty"`
	Status                string                    `json:"status,omitempty"`
	StatusDefaultText     string                    `json:"statusDefaultText,omitempty"`
	Type                  *string                   `json:"type"`
	TypeDefaultText       string                    `json:"typeDefaultText,omitempty"`
}

type Scan struct {
	ArtifactID            int64                `json:"artifactId"`
	BuildID               string               `json:"buildId"`
	BuildLabel            string               `json:"buildLabel"`
	BuildProject          string               `json:"buildProject"`
	BuildVersion          string               `json:"buildVersion"`
	Certification         string               `json:"certification"`
	ElapsedTime           string               `json:"elapsedTime"`
	EngineVersion         string               `json:"engineVersion"`
	ExecLOC               int32                `json:"execLOC"`
	FortifyAnnotationsLOC int32                `json:"fortifyAnnotationsLOC"`
	GUID                  string               `json:"guid"`
	Hostname              string               `json:"hostname"`
	ID                    int64                `json:"id,omitempty"`
	NoOfFiles             int32                `json:"noOfFiles"`
	TotalLOC              int32                `json:"totalLOC"`
	Type                  string               `json:"type"`
	UploadDate            Iso8601MilliDateTime `json:"uploadDate"`
}

type SelectorOption struct {
	DisplayName string `json:"displayName"`
	GUID        string `json:"guid"`
	Value       string `json:"value"`
}

type UserPhoto struct {
	Photo         []byte  `json:"photo"`
	PhotoMimeType *string `json:"photoMimeType"`
}
