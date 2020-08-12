package cmd

import (
	"encoding/json"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
)

func aAKaaSCheckCVs(config aAKaaSCheckCVsOptions, telemetryData *telemetry.CustomData, commonPipelineEnvironment *aAKaaSCheckCVsCommonPipelineEnvironment) {
	// for command execution use Command
	c := command.Command{}
	// reroute command output to logging framework
	c.Stdout(log.Writer())
	c.Stderr(log.Writer())

	// for http calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	// error situations should stop execution through log.Entry().Fatal() call which leads to an os.Exit(1) in the end
	err := runAAKaaSCheckCVs(&config, telemetryData, &c, commonPipelineEnvironment)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAAKaaSCheckCVs(config *aAKaaSCheckCVsOptions, telemetryData *telemetry.CustomData, command command.ExecRunner, cpe *aAKaaSCheckCVsCommonPipelineEnvironment) error {
	log.Entry().WithField("LogField", "Log field content").Info("This is just a demo for a simple step.")
	// TODO mit Matthias wegen Namen sprechen, statt AAKaaS => abapAssemblyKit (AddonAssemblyKit?)
	//TODO große repositories structur in abapUtils (mit SPSLevel usw), gemeinsam mit Assembly PR in den Master
	log.Entry().Info("repos from cpe %v", config.Repositories)

	var repos []repositories
	json.Unmarshal([]byte(config.Repositories), &repos)
	log.Entry().Info("repos as struct %v", repos)
	for i := range repos {
		repos[i].Spslevel = "10"
	}
	log.Entry().Info("repos after spsLevel set %v", repos)

	reposBackToCPE, _ := json.Marshal(repos)
	cpe.abap.repositories = string(reposBackToCPE)
	return nil
}
