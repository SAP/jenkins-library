package ans

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEvent_MergeWithJSON(t *testing.T) {
	tests := []struct {
		name          string
		eventJSON     string
		existingEvent Event
		wantEvent     Event
		wantErr       bool
	}{
		{
			name:      "Proper event JSON yields correct event",
			eventJSON: `{"eventType": "my event","eventTimestamp":1647526655}`,
			wantEvent: Event{
				EventType:      "my event",
				EventTimestamp: 1647526655,
			},
			wantErr: false,
		},
		{
			name:      "Merging events includes all parts",
			eventJSON: `{"eventType": "my event", "eventTimestamp": 1647526655, "tags": {"we": "were", "here": "first"}, "resource": {"resourceInstance": "blarp", "resourceName": "was changed"}}`,
			existingEvent: Event{
				EventType: "Bleep",
				Subject:   "Bloop",
				Tags:      map[string]interface{}{"Some": 1.0, "Additional": "a string", "Tags": true},
				Resource: &Resource{
					ResourceType: "blurp",
					ResourceName: "blorp",
				},
			},
			wantEvent: Event{
				EventType:      "my event",
				EventTimestamp: 1647526655,
				Subject:        "Bloop",
				Tags:           map[string]interface{}{"we": "were", "here": "first", "Some": 1.0, "Additional": "a string", "Tags": true},
				Resource: &Resource{
					ResourceType:     "blurp",
					ResourceName:     "was changed",
					ResourceInstance: "blarp",
				},
			},
			wantErr: false,
		},
		{
			name:      "Faulty JSON yields error",
			eventJSON: `bli-da-blup`,
			wantErr:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEvent := tt.existingEvent
			err := gotEvent.MergeWithJSON([]byte(tt.eventJSON))
			if (err != nil) != tt.wantErr {
				t.Errorf("MergeWithJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.wantEvent, gotEvent, "Received Event is not as expected.")
		})
	}
}

func TestEvent_SetLogLevel(t *testing.T) {
	tests := []struct {
		name         string
		level        logrus.Level
		wantSeverity string
		wantCategory string
	}{
		{
			name:         "InfoLevel yields INFO and NOTIFICATION",
			level:        logrus.InfoLevel,
			wantSeverity: infoSeverity,
			wantCategory: notificationCategory,
		},
		{
			name:         "DebugLevel yields INFO and NOTIFICATION",
			level:        logrus.DebugLevel,
			wantSeverity: infoSeverity,
			wantCategory: notificationCategory,
		},
		{
			name:         "WarnLevel yields WARNING and ALERT",
			level:        logrus.WarnLevel,
			wantSeverity: warningSeverity,
			wantCategory: alertCategory,
		},
		{
			name:         "ErrorLevel yields ERROR and EXCEPTION",
			level:        logrus.ErrorLevel,
			wantSeverity: errorSeverity,
			wantCategory: exceptionCategory,
		},
		{
			name:         "FatalLevel yields FATAL and EXCEPTION",
			level:        logrus.FatalLevel,
			wantSeverity: fatalSeverity,
			wantCategory: exceptionCategory,
		},
		{
			name:         "PanicLevel yields FATAL and EXCEPTION",
			level:        logrus.PanicLevel,
			wantSeverity: fatalSeverity,
			wantCategory: exceptionCategory,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := &Event{}
			event.SetLogLevel(tt.level)
			assert.Equal(t, tt.wantSeverity, event.Severity, "Got wrong severity")
			assert.Equal(t, tt.wantCategory, event.Category, "Got wrong category")
		})
	}
}
