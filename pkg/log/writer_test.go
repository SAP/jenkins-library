package log

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

type mockLogger struct {
	infoLines  []string
	warnLines  []string
	errorLines []string
}

func newMockLogger() *mockLogger {
	var logger = mockLogger{}
	return &logger
}

func (l *mockLogger) Info(args ...interface{}) {
	l.infoLines = append(l.infoLines, fmt.Sprint(args...))
}

func (l *mockLogger) Warn(args ...interface{}) {
	l.warnLines = append(l.warnLines, fmt.Sprint(args...))
}

func (l *mockLogger) Error(args ...interface{}) {
	l.errorLines = append(l.errorLines, fmt.Sprint(args...))
}

func TestWriter(t *testing.T) {
	t.Run("should buffer and append correctly", func(t *testing.T) {
		mockLogger := newMockLogger()
		writer := logrusWriter{logger: mockLogger}

		written, err := writer.Write([]byte("line "))
		assert.Equal(t, len("line "), written)
		assert.NoError(t, err)

		written, err = writer.Write([]byte("without "))
		assert.Equal(t, len("without "), written)
		assert.NoError(t, err)

		written, err = writer.Write([]byte("linebreaks"))
		assert.Equal(t, len("linebreaks"), written)
		assert.NoError(t, err)

		assert.Equal(t, 0, len(mockLogger.infoLines))
		assert.Equal(t, 0, len(mockLogger.warnLines))
		assert.Equal(t, 0, len(mockLogger.errorLines))

		assert.Equal(t, "line without linebreaks", writer.buffer.String())
	})

	t.Run("should forward to info log", func(t *testing.T) {
		mockLogger := newMockLogger()
		writer := logrusWriter{logger: mockLogger}

		line := "line with linebreak\n"

		written, err := writer.Write([]byte(line))
		assert.Equal(t, len(line), written)
		assert.NoError(t, err)

		if assert.Equal(t, 1, len(mockLogger.infoLines)) {
			assert.Equal(t, "line with linebreak", mockLogger.infoLines[0])
		}
		assert.Equal(t, 0, len(mockLogger.warnLines))
		assert.Equal(t, 0, len(mockLogger.errorLines))

		assert.Equal(t, 0, writer.buffer.Len())
	})

	t.Run("should split at line breaks", func(t *testing.T) {
		mockLogger := newMockLogger()
		writer := logrusWriter{logger: mockLogger}

		input := "line\nwith\nlinebreaks\n"

		written, err := writer.Write([]byte(input))
		assert.Equal(t, len(input), written)
		assert.NoError(t, err)

		if assert.Equal(t, 3, len(mockLogger.infoLines)) {
			assert.Equal(t, "line", mockLogger.infoLines[0])
			assert.Equal(t, "with", mockLogger.infoLines[1])
			assert.Equal(t, "linebreaks", mockLogger.infoLines[2])
		}
		assert.Equal(t, 0, len(mockLogger.warnLines))
		assert.Equal(t, 0, len(mockLogger.errorLines))

		assert.Equal(t, 0, writer.buffer.Len())
	})
	t.Run("should output empty lines", func(t *testing.T) {
		mockLogger := newMockLogger()
		writer := logrusWriter{logger: mockLogger}

		input := "\n\n"

		written, err := writer.Write([]byte(input))
		assert.Equal(t, len(input), written)
		assert.NoError(t, err)

		if assert.Equal(t, 2, len(mockLogger.infoLines)) {
			assert.Equal(t, "", mockLogger.infoLines[0])
			assert.Equal(t, "", mockLogger.infoLines[1])
		}
		assert.Equal(t, 0, len(mockLogger.warnLines))
		assert.Equal(t, 0, len(mockLogger.errorLines))

		assert.Equal(t, 0, writer.buffer.Len())
	})
}
