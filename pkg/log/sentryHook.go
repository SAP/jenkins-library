package log

import (
	"fmt"
	"log"
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

type SentryHook struct {
	levels        []logrus.Level
	Hub           *sentry.Hub
	tags          map[string]string
	Event         *sentry.Event
	correlationId string
}

func NewSentryHook(sentryDsn, correlationId string) SentryHook {
	if err := sentry.Init(sentry.ClientOptions{
		Dsn:              sentryDsn,
		AttachStacktrace: true,
	}); err != nil {
		log.Printf("cannot initialize sentry: %v", err)
	}
	h := SentryHook{
		levels:        []logrus.Level{logrus.PanicLevel, logrus.FatalLevel, logrus.ErrorLevel},
		Hub:           sentry.CurrentHub(),
		tags:          make(map[string]string),
		Event:         sentry.NewEvent(),
		correlationId: correlationId,
	}
	return h
}

func (sentryHook *SentryHook) Levels() []logrus.Level {
	return sentryHook.levels
}

func (sentryHook *SentryHook) Fire(entry *logrus.Entry) error {
	event := sentry.NewEvent()
	event.Level = levelMap[entry.Level]

	sentryHook.tags["correlationId"] = sentryHook.correlationId
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
