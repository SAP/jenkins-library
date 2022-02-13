package log

// ErrorCategory defines the category of a pipeline error
type ErrorCategory int

// Error categories which allow categorizing failures
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
var fatalError []byte

func (e ErrorCategory) String() string {
	return [...]string{
		"undefined",
		"build",
		"compliance",
		"config",
		"custom",
		"infrastructure",
		"service",
		"test",
	}[e]
}

// ErrorCategoryByString returns the error category based on the category text
func ErrorCategoryByString(category string) ErrorCategory {
	switch category {
	case "build":
		return ErrorBuild
	case "compliance":
		return ErrorCompliance
	case "config":
		return ErrorConfiguration
	case "custom":
		return ErrorCustom
	case "infrastructure":
		return ErrorInfrastructure
	case "service":
		return ErrorService
	case "test":
		return ErrorTest
	}
	return ErrorUndefined
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

// SetFatalErrorDetail sets the fatal error to be stored
func SetFatalErrorDetail(error []byte) {
	fatalError = error
}

// GetFatalErrorDetail retrieves the error which is currently known to the execution of a step
func GetFatalErrorDetail() []byte {
	return fatalError
}
