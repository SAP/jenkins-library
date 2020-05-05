package log

import (
	"bytes"
	"sync"

	"github.com/sirupsen/logrus"
)

// logrusWriter can be plugged in between a tool's output and the
// logrus.Entry's writer in order to support writing chunks larger than 64K without linebreaks.
// Implementation copied from https://github.com/sirupsen/logrus/issues/564
type logrusWriter struct {
	entry  *logrus.Entry
	buffer bytes.Buffer
	mutex  sync.Mutex
}

func (w *logrusWriter) Write(b []byte) (int, error) {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	origLen := len(b)
	for {
		if len(b) == 0 {
			return origLen, nil
		}
		i := bytes.IndexByte(b, '\n')
		if i < 0 {
			w.buffer.Write(b)
			return origLen, nil
		}

		w.buffer.Write(b[:i])
		w.alwaysFlush()
		b = b[i+1:]
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
