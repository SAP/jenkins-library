package log

import (
	"errors"
	"testing"

	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestSentryHookLevels(t *testing.T) {
	hook := NewSentryHook("", "")
	assert.Equal(t, []logrus.Level{logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel}, hook.Levels())
}

func TestSentryHookDsn(t *testing.T) {
	sentryDsn := "http://publickey@url.to.sentry/1"
	hook := NewSentryHook(sentryDsn, "")

	assert.Equal(t, hook.Hub.Client().Options().Dsn, sentryDsn)
}

func TestSentryHookNoErrorOnInvalidDsn(t *testing.T) {
	hook := NewSentryHook("invalid sentry dsn", "")

	entry := logrus.Entry{}
	err := hook.Fire(&entry)
	assert.NoError(t, err)
}

func TestSentryHookEventLevel(t *testing.T) {
	sentryDsn := "http://publickey@url.to.sentry/1"
	hook := NewSentryHook(sentryDsn, "")

	entry := logrus.Entry{}
	err := hook.Fire(&entry)

	assert.NoError(t, err)
	assert.Equal(t, sentry.LevelFatal, hook.Event.Level)
}

func TestSentryHookManualTag(t *testing.T) {
	sentryDsn := "http://publickey@url.to.sentry/1"
	hook := NewSentryHook(sentryDsn, "")
	key := "testkey"
	value := "testValue"
	hook.tags[key] = value

	assert.NotNil(t, hook.tags[key])
	assert.Equal(t, value, hook.tags[key])

	entry := logrus.Entry{}
	err := hook.Fire(&entry)

	assert.NoError(t, err)
	assert.Equal(t, value, hook.tags[key])
}

func TestSentryHookAutomaticTags(t *testing.T) {
	sentryDsn := "http://publickey@url.to.sentry/1"
	corrId := "correlationId"
	hook := NewSentryHook(sentryDsn, corrId)
	stepName := "stepName"
	category := "category"
	value := "testValue"

	entry := logrus.Entry{
		Data: logrus.Fields{
			stepName: value,
			category: value,
		},
	}
	err := hook.Fire(&entry)

	assert.NoError(t, err)
	assert.Equal(t, value, hook.tags[stepName])
	assert.Equal(t, value, hook.tags[category])
	assert.Equal(t, corrId, hook.tags[corrId])
}

func TestSentryHookException(t *testing.T) {
	sentryDsn := "http://publickey@url.to.sentry/1"
	hook := NewSentryHook(sentryDsn, "")
	hook.Event = sentry.NewEvent()
	errVal := errors.New("Exception value")
	errorMessage := "actual error message"
	exception := sentry.Exception{
		Type:  errorMessage,
		Value: errVal.Error(),
	}

	entry := logrus.Entry{
		Data: logrus.Fields{
			logrus.ErrorKey: errVal,
		},
		Message: errorMessage,
	}
	err := hook.Fire(&entry)

	assert.NoError(t, err)
	assert.Equal(t, []sentry.Exception{exception}, hook.Event.Exception)
}
