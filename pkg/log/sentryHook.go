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
	Entry().Debugf("Initializing Sentry with DSN %v", sentryDsn)
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              sentryDsn,
		AttachStacktrace: true,
	}); err != nil {
		Entry().Warnf("cannot initialize sentry: %v", err)
	}
	h := SentryHook{
		levels:        []logrus.Level{logrus.PanicLevel, logrus.FatalLevel},
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
	sentryHook.Event.Level = levelMap[entry.Level]
	sentryHook.Event.Message = entry.Message
	errValue := ""
	sentryHook.tags["correlationId"] = sentryHook.correlationID
	sentryHook.tags["category"] = GetErrorCategory().String()
	for k, v := range entry.Data {
		if k == "stepName" || k == "category" {
			sentryHook.tags[k] = fmt.Sprint(v)
		}
		if k == "error" {
			errValue = fmt.Sprint(v)
		}
		sentryHook.Event.Extra[k] = v
	}
	sentryHook.Hub.Scope().SetTags(sentryHook.tags)

	exception := sentry.Exception{
		Type:  entry.Message,
		Value: errValue,
	}

	// if error type found overwrite exception value and add stacktrace
	err, ok := entry.Data[logrus.ErrorKey].(error)
	if ok {
		exception.Value = err.Error()
		sentryHook.Event.Message = reflect.TypeOf(err).String()
		if sentryHook.Hub.Client().Options().AttachStacktrace {
			exception.Stacktrace = sentry.ExtractStacktrace(err)
		}

	}
	sentryHook.Event.Exception = []sentry.Exception{exception}

	sentryHook.Hub.CaptureEvent(sentryHook.Event)
	return nil
}
