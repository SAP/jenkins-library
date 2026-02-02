package events

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_PayloadTaskRunFinished_ToJSON(t *testing.T) {
	cases := []struct {
		name    string
		payload PayloadTaskRunFinished
	}{
		{name: "all fields set", payload: PayloadTaskRunFinished{TaskName: "build", StageName: "dev", Outcome: "SUCCESS"}},
		{name: "empty fields", payload: PayloadTaskRunFinished{TaskName: "", StageName: "", Outcome: ""}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert := assert.New(t)

			// test
			gotStr := (&tc.payload).ToJSON()
			assert.NotEmpty(gotStr, "ToJSON returned empty string")
			var got PayloadTaskRunFinished
			assert.NoError(json.Unmarshal([]byte(gotStr), &got), "failed to unmarshal JSON from ToJSON()")
			// assert
			assert.Equal(tc.payload.TaskName, got.TaskName)
			assert.Equal(tc.payload.StageName, got.StageName)
			assert.Equal(tc.payload.Outcome, got.Outcome)
		})
	}
}
