package log

import (
	"bytes"
	"strings"
	"sync"
)

type logTarget interface {
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
}

// logrusWriter can be used as the destination for a tool's std output and forwards
// chunks between linebreaks to the logrus framework. This works around a problem
// with using Entry().Writer() directly, since that doesn't support chunks
// larger than 64K without linebreaks.
// Implementation copied from https://github.com/sirupsen/logrus/issues/564
type logrusWriter struct {
	logger logTarget
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
		linebreakIndex := bytes.IndexByte(buffer, '\n')
		if linebreakIndex < 0 {
			w.buffer.Write(buffer)
			return origLen, nil
		}

		w.buffer.Write(buffer[:linebreakIndex])
		w.alwaysFlush()
		buffer = buffer[linebreakIndex+1:]
	}
}

func (w *logrusWriter) alwaysFlush() {
	message := w.buffer.String()
	w.buffer.Reset()
	
	// Check for step error patterns first
	matched, enhancedMessage := checkErrorPatterns(message)
	if matched {
		// Use enhanced message if pattern matched
		message = enhancedMessage
	}
	
	// Align level with underlying tool (like maven or npm)
	// This is to avoid confusion when maven or npm print errors or warnings which piper would print as "info"
	if strings.Contains(message, "ERROR") || strings.Contains(message, "ERR!") {
		w.logger.Error(message)
	} else if strings.Contains(message, "WARN") {
		w.logger.Warn(message)
	} else {
		w.logger.Info(message)
	}
}

func (w *logrusWriter) Flush() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.buffer.Len() != 0 {
		w.alwaysFlush()
	}
}
