package events

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_PayloadGeneric_Merge(t *testing.T) {
	cases := []struct {
		name           string
		payloadString  string
		otherString    string
		expectedString string
	}{
		{name: "fields set", payloadString: `{"name":"test","value":42}`, otherString: `{"extra":"data"}`, expectedString: `{"extra":"data","name":"test","value":42}`},
		{name: "empty object", payloadString: "{}", otherString: `{"newField":123}`, expectedString: `{"newField":123}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert := assert.New(t)
			payload := PayloadGeneric{JSONData: tc.payloadString}

			// test
			payload.Merge(tc.otherString)
			assert.NotEmpty(payload.JSONData, "Merge returned empty JSONData")
			// assert
			assert.Equal(tc.expectedString, payload.JSONData)
		})
	}
}

func Test_PayloadGeneric_ToJSON(t *testing.T) {
	cases := []struct {
		name          string
		payloadString string
	}{
		{name: "fields set", payloadString: `{"name":"test","value":42}`},
		{name: "empty object", payloadString: "{}"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert := assert.New(t)
			payload := PayloadGeneric{JSONData: tc.payloadString}

			// test
			gotStr := payload.ToJSON()
			assert.NotEmpty(gotStr, "ToJSON returned empty string")
			// assert
			assert.Equal(tc.payloadString, gotStr)
		})
	}
}
