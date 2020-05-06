package log

import (
	"bytes"
	"sync"

	"github.com/sirupsen/logrus"
)

// logrusWriter can be used as the destination for a tool's std output and forwards
// chunks between linebreaks to the logrus framework. This works around a problem
// with using Entry().Writer() directly, since that doesn't support chunks
// larger than 64K without linebreaks.
// Implementation copied from https://github.com/sirupsen/logrus/issues/564
type logrusWriter struct {
	entry  *logrus.Entry
	buffer bytes.Buffer
	mutex  sync.Mutex
}

func (w *logrusWriter) Write(buffer []byte) (int, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	origLen := len(buffer)
	for {
		if len(buffer) == 0 {
			return origLen, nil
		}
		i := bytes.IndexByte(buffer, '\n')
		if i < 0 {
			w.buffer.Write(buffer)
			return origLen, nil
		}

		w.buffer.Write(buffer[:i])
		w.alwaysFlush()
		buffer = buffer[i+1:]
	}
}

func (w *logrusWriter) alwaysFlush() {
	w.entry.Info(w.buffer.String())
	w.buffer.Reset()
}

func (w *logrusWriter) Flush() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.buffer.Len() != 0 {
		w.alwaysFlush()
	}
}
