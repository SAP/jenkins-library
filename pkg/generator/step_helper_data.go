package generator

import (
	"io"
	"os"
)

// StepHelperData is used to transport the needed parameters and functions from the step generator to the step generation.
type StepHelperData struct {
	OpenFile     func(s string) (io.ReadCloser, error)
	WriteFile    func(filename string, data []byte, perm os.FileMode) error
	ExportPrefix string
}
