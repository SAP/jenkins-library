package events

import cloudevents "github.com/cloudevents/sdk-go/v2"

type Option func(o *cloudevents.EventContextV1)

func WithID(id string) Option {
	return func(o *cloudevents.EventContextV1) {
		o.SetID(id)
	}
}

func WithType(etype string) Option {
	return func(o *cloudevents.EventContextV1) {
		o.SetType(etype)
	}
}
