package eventing

// taskRunFinishedPayload is the data payload for a TaskRunFinished CloudEvent.
type taskRunFinishedPayload struct {
	TaskName  string `json:"taskName"`
	StageName string `json:"stageName"`
	Outcome   string `json:"outcome"`
}
