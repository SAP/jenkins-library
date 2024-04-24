package abaputils

import (
	"errors"
	"time"

	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
)

type SoftwareComponentApiManagerInterface interface {
	GetAPI(con ConnectionDetailsHTTP, repo Repository) (SoftwareComponentApiInterface, error)
	GetPollIntervall() time.Duration
}

type SoftwareComponentApiManager struct {
	Client        piperhttp.Sender
	PollIntervall time.Duration
	Force0510     bool
}

func (manager *SoftwareComponentApiManager) GetAPI(con ConnectionDetailsHTTP, repo Repository) (SoftwareComponentApiInterface, error) {

	var err0948 error
	if !manager.Force0510 {
		// Initialize SAP_COM_0948, if it does not work, use SAP_COM_0510
		sap_com_0948 := SAP_COM_0948{}
		sap_com_0948.init(con, manager.Client, repo)
		err0948 = sap_com_0948.initialRequest()
		if err0948 == nil {
			return &sap_com_0948, nil
		}
	}

	sap_com_0510 := SAP_COM_0510{}
	sap_com_0510.init(con, manager.Client, repo)
	err0510 := sap_com_0510.initialRequest()
	if err0510 == nil {
		log.Entry().Infof("SAP_COM_0510 will be replaced by SAP_COM_0948 starting from the SAP BTP, ABAP environment release 2402.")
		return &sap_com_0510, nil
	}

	log.Entry().Errorf("Could not connect via SAP_COM_0948: %s", err0948)
	log.Entry().Errorf("Could not connect via SAP_COM_0510: %s", err0510)

	return nil, errors.New("Could not initialize API")
}

func (manager *SoftwareComponentApiManager) GetPollIntervall() time.Duration {
	if manager.PollIntervall == 0 {
		manager.PollIntervall = 5 * time.Second
	}
	return manager.PollIntervall
}

type SoftwareComponentApiInterface interface {
	init(con ConnectionDetailsHTTP, client piperhttp.Sender, repo Repository)
	initialRequest() error
	setSleepTimeConfig(timeUnit time.Duration, maxSleepTime time.Duration)
	getSleepTime(n int) (time.Duration, error)
	getUUID() string
	GetRepository() (bool, string, error)
	Clone() error
	Pull() error
	CheckoutBranch() error
	GetAction() (string, error)
	CreateTag(tag Tag) error
	GetLogOverview() ([]LogResultsV2, error)
	GetLogProtocol(LogResultsV2, int) (result []LogProtocol, count int, err error)
	GetExecutionLog() (ExecutionLog, error)
}

/****************************************
 *	Structs for the A4C_A2G_GHA service *
 ****************************************/

// ActionEntity struct for the Pull/Import entity A4C_A2G_GHA_SC_IMP
type ActionEntity struct {
	Metadata          AbapMetadata `json:"__metadata"`
	UUID              string       `json:"uuid"`
	Namespace         string       `json:"namespace"`
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
	Metadata     AbapMetadata `json:"__metadata"`
	ScName       string       `json:"sc_name"`
	ActiveBranch string       `json:"active_branch"`
	AvailOnInst  bool         `json:"avail_on_inst"`
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

type ExecutionLog struct {
	Value []ExecutionLogValue `json:"value"`
}

type ExecutionLogValue struct {
	IndexNo   int    `json:"index_no"`
	Type      string `json:"type"`
	Descr     string `json:"descr"`
	Timestamp string `json:"timestamp"`
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

type LogProtocolResultsV4 struct {
	Results []LogProtocol `json:"value"`
	Count   int           `json:"@odata.count"`
}

type LogProtocol struct {
	// Metadata      AbapMetadata `json:"__metadata"`
	OverviewIndex int    `json:"log_index"`
	ProtocolLine  int    `json:"index_no"`
	Type          string `json:"type"`
	Description   string `json:"descr"`
	Timestamp     string `json:"timestamp"`
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

type CreateTagBacklog struct {
	RepositoryName string
	CommitID       string
	Tags           []Tag
}

type Tag struct {
	TagName        string
	TagDescription string
}

type CreateTagBody struct {
	RepositoryName string `json:"sc_name"`
	CommitID       string `json:"commit_id"`
	Tag            string `json:"tag_name"`
	Description    string `json:"tag_description"`
}

type CreateTagResponse struct {
	UUID string `json:"uuid"`
}
