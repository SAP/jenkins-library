package log

import (
	"fmt"
	"reflect"

	"github.com/getsentry/sentry-go"
	"github.com/sirupsen/logrus"
)

// conversion from logrus log level to sentry log level
var (
	levelMap = map[logrus.Level]sentry.Level{
		logrus.TraceLevel: sentry.LevelDebug,
		logrus.DebugLevel: sentry.LevelDebug,
		logrus.InfoLevel:  sentry.LevelInfo,
		logrus.WarnLevel:  sentry.LevelWarning,
		logrus.ErrorLevel: sentry.LevelError,
		logrus.FatalLevel: sentry.LevelFatal,
		logrus.PanicLevel: sentry.LevelFatal,
	}
)

// SentryHook provides a logrus hook which enables error logging to sentry platform.
// This is helpful in order to provide better monitoring and alerting on errors
// as well as the given error details can help to find the root cause of bugs.
type SentryHook struct {
	levels        []logrus.Level
	Hub           *sentry.Hub
	tags          map[string]string
	Event         *sentry.Event
	correlationID string
}

// NewSentryHook initializes sentry sdk with dsn and creates new hook
func NewSentryHook(sentryDsn, correlationID string) SentryHook {
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              sentryDsn,
		AttachStacktrace: true,
	}); err != nil {
		Entry().Infof("cannot initialize sentry: %v", err)
	}
	h := SentryHook{
		levels:        []logrus.Level{logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel},
		Hub:           sentry.CurrentHub(),
		tags:          make(map[string]string),
		Event:         sentry.NewEvent(),
		correlationID: correlationID,
	}
	return h
}

// Levels returns the supported log level of the hook.
func (sentryHook *SentryHook) Levels() []logrus.Level {
	return sentryHook.levels
}

// Fire creates a new event from the error and sends it to sentry
func (sentryHook *SentryHook) Fire(entry *logrus.Entry) error {
	event := sentry.NewEvent()
	event.Level = levelMap[entry.Level]

	sentryHook.tags["correlationId"] = sentryHook.correlationID
	for k, v := range entry.Data {
		if k == "stepName" || k == "category" {
			sentryHook.tags[k] = fmt.Sprint(v)
		}
		event.Extra[k] = v
	}
	sentryHook.Hub.Scope().SetTags(sentryHook.tags)

	err, ok := entry.Data[logrus.ErrorKey].(error)
	if ok {
		exception := sentry.Exception{
			Type:  entry.Message,
			Value: err.Error(),
		}
		event.Message = reflect.TypeOf(err).String()

		if sentryHook.Hub.Client().Options().AttachStacktrace {
			exception.Stacktrace = sentry.ExtractStacktrace(err)
		}

		event.Exception = []sentry.Exception{exception}
	}

	sentryHook.Event.Level = event.Level
	sentryHook.Event.Exception = event.Exception
	sentryHook.Hub.CaptureEvent(event)

	return nil
}
