package abaputils

import (
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
)

type SoftwareComponentApiManagerInterface interface {
	GetAPI(con ConnectionDetailsHTTP, client piperhttp.Sender, repo Repository) (SoftwareComponentApiInterface, error)
}

type SoftwareComponentApiManager struct{}

func (manager *SoftwareComponentApiManager) GetAPI(con ConnectionDetailsHTTP, client piperhttp.Sender, repo Repository) (SoftwareComponentApiInterface, error) {
	sap_com_0510 := SAP_COM_0510{}
	sap_com_0510.init(con, client, repo)

	// Initialize all APIs, use the one that returns a response
	// Currently SAP_COM_0510, later SAP_COM_0948
	err := sap_com_0510.initialRequest()
	return &sap_com_0510, err
}

type SoftwareComponentApiInterface interface {
	init(con ConnectionDetailsHTTP, client piperhttp.Sender, repo Repository)
	initialRequest() error
	Clone() error
	GetRepository() (bool, string, error)
	GetAction() (string, error)
	GetLogOverview() (ActionEntity, error)
	GetLogProtocol(LogResultsV2, int) (body LogProtocolResults, err error)
}

/****************************************
 *	Structs for the A4C_A2G_GHA service *
 ****************************************/

// ActionEntity struct for the Pull/Import entity A4C_A2G_GHA_SC_IMP
type ActionEntity struct {
	Metadata          AbapMetadata `json:"__metadata"`
	UUID              string       `json:"uuid"`
	Namespace         string       `json:"namepsace"`
	ScName            string       `json:"sc_name"`
	ImportType        string       `json:"import_type"`
	BranchName        string       `json:"branch_name"`
	StartedByUser     string       `json:"user_name"`
	Status            string       `json:"status"`
	StatusDescription string       `json:"status_descr"`
	CommitID          string       `json:"commit_id"`
	StartTime         string       `json:"start_time"`
	ChangeTime        string       `json:"change_time"`
	ToExecutionLog    AbapLogs     `json:"to_Execution_log"`
	ToTransportLog    AbapLogs     `json:"to_Transport_log"`
	ToLogOverview     AbapLogsV2   `json:"to_Log_Overview"`
}

// BranchEntity struct for the Branch entity A4C_A2G_GHA_SC_BRANCH
type BranchEntity struct {
	Metadata      AbapMetadata `json:"__metadata"`
	ScName        string       `json:"sc_name"`
	Namespace     string       `json:"namepsace"`
	BranchName    string       `json:"branch_name"`
	ParentBranch  string       `json:"derived_from"`
	CreatedBy     string       `json:"created_by"`
	CreatedOn     string       `json:"created_on"`
	IsActive      bool         `json:"is_active"`
	CommitID      string       `json:"commit_id"`
	CommitMessage string       `json:"commit_message"`
	LastCommitBy  string       `json:"last_commit_by"`
	LastCommitOn  string       `json:"last_commit_on"`
}

// CloneEntity struct for the Clone entity A4C_A2G_GHA_SC_CLONE
type CloneEntity struct {
	Metadata          AbapMetadata `json:"__metadata"`
	UUID              string       `json:"uuid"`
	ScName            string       `json:"sc_name"`
	BranchName        string       `json:"branch_name"`
	ImportType        string       `json:"import_type"`
	Namespace         string       `json:"namepsace"`
	Status            string       `json:"status"`
	StatusDescription string       `json:"status_descr"`
	StartedByUser     string       `json:"user_name"`
	StartTime         string       `json:"start_time"`
	ChangeTime        string       `json:"change_time"`
}

type RepositoryEntity struct {
	Metadata            AbapMetadata `json:"__metadata"`
	ScName              string       `json:"sc_name"`
	ActiveBranch        string       `json:"active_branch"`
	AvailableOnInstance bool         `json:"avail_on_inst"`
}

// AbapLogs struct for ABAP logs
type AbapLogs struct {
	Results []LogResults `json:"results"`
}

type AbapLogsV2 struct {
	Results []LogResultsV2 `json:"results"`
}

type LogResultsV2 struct {
	Metadata      AbapMetadata        `json:"__metadata"`
	Index         int                 `json:"log_index"`
	Name          string              `json:"log_name"`
	Status        string              `json:"type_of_found_issues"`
	Timestamp     string              `json:"timestamp"`
	ToLogProtocol LogProtocolDeferred `json:"to_Log_Protocol"`
}

type LogProtocolDeferred struct {
	Deferred URI `json:"__deferred"`
}

type URI struct {
	URI string `json:"uri"`
}

type LogProtocolResults struct {
	Results []LogProtocol `json:"results"`
	Count   string        `json:"__count"`
}

type LogProtocol struct {
	Metadata      AbapMetadata `json:"__metadata"`
	OverviewIndex int          `json:"log_index"`
	ProtocolLine  int          `json:"index_no"`
	Type          string       `json:"type"`
	Description   string       `json:"descr"`
	Timestamp     string       `json:"timestamp"`
}

// LogResults struct for Execution and Transport Log entities A4C_A2G_GHA_SC_LOG_EXE and A4C_A2G_GHA_SC_LOG_TP
type LogResults struct {
	Index       string `json:"index_no"`
	Type        string `json:"type"`
	Description string `json:"descr"`
	Timestamp   string `json:"timestamp"`
}

// RepositoriesConfig struct for parsing one or multiple branches and repositories configurations
type RepositoriesConfig struct {
	BranchName      string
	CommitID        string
	RepositoryName  string
	RepositoryNames []string
	Repositories    string
}

type EntitySetsForManageGitRepository struct {
	EntitySets []string `json:"EntitySets"`
}
