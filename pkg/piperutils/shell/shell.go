package shell

import (
	"fmt"
	"strings"
)

// WrapInQuotes wraps a string in single quotes to be used in a shell call.
// This means single quotes within the string will be escaped accordingly.
// The function is helful when passing password parameters to shell calls.
func WrapInQuotes(s string) string {
	res := strings.ReplaceAll(s, "'", "'\"'\"'")
	return fmt.Sprintf("'%v'", res)
}
