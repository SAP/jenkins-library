package events

type PayloadTaskRunFinished struct {
	TaskName  string `json:"taskName"`
	StageName string `json:"stageName"`
	Outcome   string `json:"outcome"`
}

func NewPayloadTaskRunFinished(stageName, taskName, returnCode string) PayloadTaskRunFinished {
	outcome := "failure"
	if returnCode == "0" {
		outcome = "success"
	}
	return PayloadTaskRunFinished{
		TaskName:  taskName,
		StageName: stageName,
		Outcome:   outcome,
	}
}
