package eventsdemo

// payload part of the event
// in general, consumers of the event package should not set fields of this struct directly,
// but use setter functions or constructors provided.
type TaskRunFinishedPayload struct {
	TaskName  string `json:"taskName"`
	StageName string `json:"stageName"`
	Outcome   string `json:"outcome"`
}

func NewTaskRunFinishedPayload(taskName, stageName, outcome string) TaskRunFinishedPayload {
	return TaskRunFinishedPayload{
		TaskName:  taskName,
		StageName: stageName,
		Outcome:   outcome,
	}
}

func (p *TaskRunFinishedPayload) SetTaskName(taskName string) {
	p.TaskName = taskName
}

func (p *TaskRunFinishedPayload) SetStageName(stageName string) {
	p.StageName = stageName
}

func (p *TaskRunFinishedPayload) SetOutcome(outcome string) {
	p.Outcome = outcome
}
