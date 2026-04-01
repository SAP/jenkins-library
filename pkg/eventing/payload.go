package eventing

// EventContext carries step-level data from the generated template into the eventing package.
type EventContext struct {
	StepName   string
	StageName  string
	ErrorCode  string
	PipelineID string
}
