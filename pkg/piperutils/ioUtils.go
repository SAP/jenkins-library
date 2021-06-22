package piperutils

import (
	"fmt"
	"io"

	"github.com/pkg/errors"
)

func CopyData(dst io.Writer, src io.Reader) (int64, error) {
	tmp := make([]byte, 256)
	bytesRead := int64(0)
	bytesWritten := int64(0)
	done := false

	for {
		n, err := src.Read(tmp)
		bytesRead += int64(n)
		if err != nil {
			if err != io.EOF {
				return bytesRead, errors.Wrap(err, "read error")
			}
			done = true
		}
		n, err = dst.Write(tmp[:n])
		bytesWritten += int64(n)
		if err != nil {
			return bytesWritten, errors.Wrap(err, "write error")
		}
		if done {
			break
		}
	}
	if bytesRead != bytesWritten {
		return bytesRead, errors.New(fmt.Sprintf("transfer error: read %v bytes but wrote %v bytes", bytesRead, bytesWritten))
	}
	return bytesWritten, nil
}
