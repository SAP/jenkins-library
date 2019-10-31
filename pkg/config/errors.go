package config

// ParseError defines an error type for configuration parsing errors
type ParseError struct {
	message string
}

// NewParseError creates a new ParseError
func NewParseError(message string) *ParseError {
	return &ParseError{
		message: message,
	}
}

// Error returns the message of the ParseError
func (e *ParseError) Error() string {
	return e.message
}
