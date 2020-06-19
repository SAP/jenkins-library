package log

// ErrorCategory defines the category of a pipeline error
type ErrorCategory int

const (
	ErrorUndefined ErrorCategory = iota
	ErrorBuild
	ErrorCompliance
	ErrorConfiguration
	ErrorCustom
	ErrorInfrastructure
	ErrorService
	ErrorTest
)

var errorCategory ErrorCategory = ErrorUndefined

func (e ErrorCategory) String() string {
	return [...]string{
		"undefined",
		"build",
		"compliance",
		"configuration",
		"custom",
		"infrastructure",
		"service",
		"test",
	}[e]
}

// SetErrorCategory sets the error category
// This can be used later by calling log.GetErrorCategory()
// In addition it will be used when exiting the program with
// log.FatalError(err, message)
func SetErrorCategory(category ErrorCategory) {
	errorCategory = category
}

// GetErrorCategory retrieves the error category which is currently known to the execution of a step
func GetErrorCategory() ErrorCategory {
	return errorCategory
}
