package log

import (
	"reflect"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestCollectorHook_Fire(t *testing.T) {
	type fields struct {
		CorrelationID string
		Messages      []Message
	}
	type args struct {
		entry *logrus.Entry
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{"Test Fire",
			fields{
				CorrelationID: "123",
				Messages:      []Message{},
			},
			args{entry: &logrus.Entry{
				Time:    time.Now(),
				Data:    logrus.Fields{"test": "test value"},
				Message: "Test Message",
			},
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialMessageLength := len(tt.fields.Messages)
			f := &CollectorHook{
				CorrelationID: tt.fields.CorrelationID,
				Messages:      tt.fields.Messages,
			}
			// Check if hook was triggered
			if err := f.Fire(tt.args.entry); (err != nil) != tt.wantErr {
				t.Errorf("Fire() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check if the message was successfully added
			if len(f.Messages) != initialMessageLength+1 {
				t.Errorf("Fire() error - Messages not added to array - Message count %v", len(f.Messages))
			}
		})
	}
}
func TestCollectorHook_Levels(t *testing.T) {
	type fields struct {
		CorrelationID string
		Messages      []Message
	}
	tests := []struct {
		name   string
		fields fields
		want   []logrus.Level
	}{
		{"Test Levels",
			fields{
				CorrelationID: "123",
				Messages: []Message{
					{
						Time:    time.Now(),
						Level:   logrus.DebugLevel,
						Message: "Test Message",
					},
				},
			},
			[]logrus.Level{logrus.InfoLevel, logrus.DebugLevel, logrus.WarnLevel, logrus.ErrorLevel, logrus.PanicLevel, logrus.FatalLevel},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &CollectorHook{
				CorrelationID: tt.fields.CorrelationID,
				Messages:      tt.fields.Messages,
			}
			if got := f.Levels(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Levels() = %v, want %v", got, tt.want)
			}
		})
	}
}
