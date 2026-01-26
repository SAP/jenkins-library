package piperutils

import (
	"fmt"
	"io"

	"github.com/pkg/errors"
)

// CopyData transfers the bytes from src to dst without doing close handling implicitly.
func CopyData(dst io.Writer, src io.Reader) (int64, error) {
	tmp := make([]byte, 256)
	bytesRead := int64(0)
	bytesWritten := int64(0)
	done := false

	for !done {
		nr, err := src.Read(tmp)
		bytesRead += int64(nr)
		if err != nil {
			if err != io.EOF {
				return bytesRead, errors.Wrap(err, "read error")
			}
			done = true
		}
		nw, err := dst.Write(tmp[:nr])
		bytesWritten += int64(nw)
		if err != nil {
			return bytesWritten, errors.Wrap(err, "write error")
		}
	}
	if bytesRead != bytesWritten {
		return bytesRead, errors.New(fmt.Sprintf("transfer error: read %v bytes but wrote %v bytes", bytesRead, bytesWritten))
	}
	return bytesWritten, nil
}
