//go:build unit
// +build unit

package ans

import (
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		},
		{
			name:      "Merging events includes all parts",
			eventJSON: `{"eventType": "my event", "eventTimestamp": 1647526655, "tags": {"we": "were", "here": "first"}, "resource": {"resourceInstance": "myResourceInstance", "resourceName": "was changed"}}`,
			existingEvent: Event{
				EventType: "test",
				Subject:   "test",
				Tags:      map[string]interface{}{"Some": 1.0, "Additional": "a string", "Tags": true},
				Resource: &Resource{
					ResourceType: "myResourceType",
					ResourceName: "myResourceName",
				},
			},
			wantEvent: Event{
				EventType:      "my event",
				EventTimestamp: 1647526655,
				Subject:        "test",
				Tags:           map[string]interface{}{"we": "were", "here": "first", "Some": 1.0, "Additional": "a string", "Tags": true},
				Resource: &Resource{
					ResourceType:     "myResourceType",
					ResourceName:     "was changed",
					ResourceInstance: "myResourceInstance",
				},
			},
		},
		{
			name:      "Faulty JSON yields error",
			eventJSON: `faulty json`,
			wantErr:   true,
		},
		{
			name:      "Non-existent field yields error",
			eventJSON: `{"unknownKey": "yields error"}`,
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
			event.SetSeverityAndCategory(tt.level)
			assert.Equal(t, tt.wantSeverity, event.Severity, "Got wrong severity")
			assert.Equal(t, tt.wantCategory, event.Category, "Got wrong category")
		})
	}
}

func TestEvent_Validate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		eventJSON string
		errMsg    string
	}{
		{
			errMsg:    "Category must be one of [EXCEPTION ALERT NOTIFICATION]",
			eventJSON: `{"category": "WRONG_CATEGORY"}`,
		},
		{
			errMsg:    "Severity must be one of [INFO NOTICE WARNING ERROR FATAL]",
			eventJSON: `{"severity": "WRONG_SEVERITY"}`,
		},
		{
			errMsg:    "Priority must be 1,000 or less",
			eventJSON: `{"priority": 1001}`,
		},
		{
			errMsg:    "Priority must be 1 or greater",
			eventJSON: `{"priority": -1}`,
		},
		{
			errMsg:    "EventTimestamp must be 0 or greater",
			eventJSON: `{"eventTimestamp": -1}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.errMsg, func(t *testing.T) {
			event := defaultEvent()
			require.NoError(t, event.MergeWithJSON([]byte(tt.eventJSON)))
			assert.EqualError(t, event.Validate(), fmt.Sprintf("%s: %s", tt.errMsg, standardErrMsg))
		})
	}
}

const standardErrMsg = "event JSON failed the validation"

func defaultEvent() Event {
	return Event{
		EventType:      "MyEvent",
		EventTimestamp: 1653485928,
		Severity:       "INFO",
		Category:       "NOTIFICATION",
		Subject:        "mySubject",
		Body:           "myBody",
		Priority:       123,
		Resource: &Resource{
			ResourceName:     "myResourceName",
			ResourceType:     "myResourceType",
			ResourceInstance: "myResourceInstance",
		},
	}
}

func TestEvent_Copy(t *testing.T) {
	t.Parallel()
	t.Run("good", func(t *testing.T) {
		originalEvent := defaultEvent()
		newEvent, err := originalEvent.Copy()
		require.NoError(t, err)
		assert.Equal(t, originalEvent, newEvent, "Events should be the same after copying.")
		newEvent.Resource.ResourceType = "different"
		assert.NotEqual(t, originalEvent, newEvent, "Events should not affect each other after copying")
	})
}
