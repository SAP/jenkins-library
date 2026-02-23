package eventing

// TaskRunFinishedPayload is the data payload for a TaskRunFinished CloudEvent.
type TaskRunFinishedPayload struct {
	TaskName  string `json:"taskName"`
	StageName string `json:"stageName"`
	Outcome   string `json:"outcome"`
}