package cmd

import (
	"net/http/cookiejar"
	"time"

	"github.com/SAP/jenkins-library/pkg/command"
	piperhttp "github.com/SAP/jenkins-library/pkg/http"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func abapEnvironmentCheckoutBranch(config abapEnvironmentCheckoutBranchOptions, telemetryData *telemetry.CustomData) {
	// for command execution use Command
	c := command.Command{}

	// get abap communication arrangement information
	connectionDetails, errorGetInfo := getAbapCommunicationArrangementInfo(config, &c)
	if errorGetInfo != nil {
		log.Entry().WithError(errorGetInfo).Fatal("Parameters for the ABAP Connection not available")
	}

	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	// Configuring the HTTP Client and CookieJar
	client := piperhttp.Client{}
	cookieJar, errorCookieJar := cookiejar.New(nil)
	if errorCookieJar != nil {
		log.Entry().WithError(errorCookieJar).Fatal("Could not create a Cookie Jar")
	}
	clientOptions := piperhttp.ClientOptions{
		MaxRequestDuration: 180 * time.Second,
		CookieJar:          cookieJar,
		Username:           connectionDetails.User,
		Password:           connectionDetails.Password,
	}
	client.SetOptions(clientOptions)

	// pollIntervall := 10 * time.Second
	// log.Entry().Infof("Start pulling %v repositories", len(config.RepositoryNames))
	// for _, repositoryName := range config.RepositoryNames { }

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runAbapEnvironmentCheckoutBranch(&config, telemetryData, &c)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAbapEnvironmentCheckoutBranch(config *abapEnvironmentCheckoutBranchOptions, telemetryData *telemetry.CustomData, command execRunner) error {
	log.Entry().WithField("LogField", "Log field content").Info("This is just a demo for a simple step.")

	return nil
}

// Software Component Entity
type abapSoftwareComponentEntity struct {
	Metadata       abapMetadata `json:"__metadata"`
	UUID           string       `json:"uuid"`
	ScName         string       `json:"sc_name"`
	Namespace      string       `json:"namepsace"`
	Status         string       `json:"status"`
	StatusDescr    string       `json:"status_descr"`
	ToExecutionLog abapLogs     `json:"to_Execution_log"`
	ToTransportLog abapLogs     `json:"to_Transport_log"`
}

// Software Component Branch Entity
type abapBranchEntity struct {
	Namespace  string `json:"namepsace"`
	ScName     string `json:"sc_name"`
	BranchName string `json:"branch_name"`
}

// type abapMetadata struct {
// 	URI string `json:"uri"`
// }

// type abapLogs struct {
// 	Results []logResults `json:"results"`
// }

// type logResults struct {
// 	Index       string `json:"index_no"`
// 	Type        string `json:"type"`
// 	Description string `json:"descr"`
// 	Timestamp   string `json:"timestamp"`
// }

// type serviceKey struct {
// 	Abap     abapConenction `json:"abap"`
// 	Binding  abapBinding    `json:"binding"`
// 	Systemid string         `json:"systemid"`
// 	URL      string         `json:"url"`
// }

// type deferred struct {
// 	URI string `json:"uri"`
// }

// type abapConenction struct {
// 	CommunicationArrangementID string `json:"communication_arrangement_id"`
// 	CommunicationScenarioID    string `json:"communication_scenario_id"`
// 	CommunicationSystemID      string `json:"communication_system_id"`
// 	Password                   string `json:"password"`
// 	Username                   string `json:"username"`
// }

// type abapBinding struct {
// 	Env     string `json:"env"`
// 	ID      string `json:"id"`
// 	Type    string `json:"type"`
// 	Version string `json:"version"`
// }

// type connectionDetailsHTTP struct {
// 	User       string `json:"user"`
// 	Password   string `json:"password"`
// 	URL        string `json:"url"`
// 	XCsrfToken string `json:"xcsrftoken"`
// }

// type abapError struct {
// 	Code    string           `json:"code"`
// 	Message abapErrorMessage `json:"message"`
// }

// type abapErrorMessage struct {
// 	Lang  string `json:"lang"`
// 	Value string `json:"value"`
// }
