//go:build unit
// +build unit

package log

import (
	"bytes"
	"net/url"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSecrets(t *testing.T) {
	t.Run("should log", func(t *testing.T) {
		secret := "password"

		outWriter := Entry().Logger.Out
		var buffer bytes.Buffer
		Entry().Logger.SetOutput(&buffer)
		defer func() { Entry().Logger.SetOutput(outWriter) }()

		Entry().Infof("My secret is %s.", secret)
		assert.Contains(t, buffer.String(), secret)

		buffer.Reset()
		RegisterSecret(secret)
		Entry().Infof("My secret is %s.", secret)
		assert.NotContains(t, buffer.String(), secret)
	})

	t.Run("should log url encoded", func(t *testing.T) {
		secret := "secret-token!0"
		encodedSecret := url.QueryEscape(secret)

		outWriter := Entry().Logger.Out
		var buffer bytes.Buffer
		Entry().Logger.SetOutput(&buffer)
		defer func() { Entry().Logger.SetOutput(outWriter) }()

		Entry().Infof("My secret is %s.", secret)
		assert.Contains(t, buffer.String(), secret)

		buffer.Reset()
		RegisterSecret(secret)
		Entry().Infof("My secret is %s.", encodedSecret)
		assert.NotContains(t, buffer.String(), encodedSecret)
	})
}

func TestWriteLargeBuffer(t *testing.T) {
	t.Run("should log large buffer without linebreaks via Writer()", func(t *testing.T) {
		b := []byte("a")
		size := 131072
		b = bytes.Repeat(b, size)
		written, err := Writer().Write(b)
		assert.Equal(t, size, written)
		assert.NoError(t, err)
	})
}

func TestPiperLogFormatter_Format(t *testing.T) {
	testTime := time.Date(2025, 12, 12, 15, 4, 5, 0, time.UTC)
	tests := []struct {
		name      string
		logFormat string
		level     logrus.Level
		msg       string
		fields    logrus.Fields
		errorVal  interface{}
		want      string
	}{
		// --- GHA cases ---
		{
			name:      "GHA info",
			logFormat: logFormatGitHubActions,
			level:     logrus.InfoLevel,
			msg:       "Info message",
			fields:    logrus.Fields{"stepName": "step1"},
			want:      "Info message\n",
		},
		{
			name:      "GHA warning",
			logFormat: logFormatGitHubActions,
			level:     logrus.WarnLevel,
			msg:       "Warn message",
			fields:    logrus.Fields{"stepName": "step2"},
			want:      "::warning::Warn message\n",
		},
		{
			name:      "GHA error",
			logFormat: logFormatGitHubActions,
			level:     logrus.ErrorLevel,
			msg:       "Error message",
			fields:    logrus.Fields{"stepName": "step3"},
			want:      "::error::Error message\n",
		},
		{
			name:      "GHA fatal",
			logFormat: logFormatGitHubActions,
			level:     logrus.FatalLevel,
			msg:       "Fatal message",
			fields:    logrus.Fields{"stepName": "step4"},
			want:      "::error::Fatal message\n",
		},
		{
			name:      "GHA debug",
			logFormat: logFormatGitHubActions,
			level:     logrus.DebugLevel,
			msg:       "Debug message",
			fields:    logrus.Fields{"stepName": "step5"},
			want:      "[DEBUG] Debug message\n",
		},
		{
			name:      "GHA error with error field",
			logFormat: logFormatGitHubActions,
			level:     logrus.ErrorLevel,
			msg:       "Error with error field",
			fields:    logrus.Fields{"stepName": "step6"},
			errorVal:  "some error",
			want:      "::error::Error with error field - some error\n",
		},
		{
			name:      "GHA mask secret",
			logFormat: logFormatGitHubActions,
			level:     logrus.InfoLevel,
			msg:       "My password is secret123",
			fields:    logrus.Fields{},
			want:      "My password is ****\n",
		},
		// --- Default format ---
		{
			name:      "Default info",
			logFormat: logFormatDefault,
			level:     logrus.InfoLevel,
			msg:       "Default info message",
			fields:    logrus.Fields{"stepName": "stepX", "foo": "bar"},
			want:      "info  stepX  - Default info message [foo=bar]\n",
		},
		// --- Plain format ---
		{
			name:      "Plain info",
			logFormat: logFormatPlain,
			level:     logrus.InfoLevel,
			msg:       "Plain info message",
			fields:    logrus.Fields{"stepName": "plainStep"},
			want:      "Plain info message\n",
		},
		// --- Timestamp format ---
		{
			name:      "Timestamp info",
			logFormat: logFormatWithTimestamp,
			level:     logrus.InfoLevel,
			msg:       "Timestamp info message",
			fields:    logrus.Fields{"stepName": "tsStep"},
			want:      testTime.Format("15:04:05") + " info  tsStep Timestamp info message\n",
		},
	}

	RegisterSecret("secret123")

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := &PiperLogFormatter{logFormat: tt.logFormat}
			entry := &logrus.Entry{
				Logger:  logrus.New(),
				Data:    tt.fields,
				Time:    testTime,
				Level:   tt.level,
				Message: tt.msg,
			}
			if tt.errorVal != nil {
				entry.Data[logrus.ErrorKey] = tt.errorVal
			}
			b, err := formatter.Format(entry)
			assert.NoError(t, err)
			assert.Equal(t, tt.want, string(b))
		})
	}
}
