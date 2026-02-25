package events

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPayloadTaskRunFinished(t *testing.T) {
	cases := []struct {
		name       string
		taskName   string
		stageName  string
		returnCode string
		expected   string
	}{
		{name: "all fields set", taskName: "build", stageName: "dev", returnCode: "0", expected: `{TaskName:build StageName:dev Outcome:success}`},
		{name: "empty fields", taskName: "", stageName: "", returnCode: "", expected: `{TaskName: StageName: Outcome:failure}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert := assert.New(t)
			// test
			payload := NewPayloadTaskRunFinished(tc.stageName, tc.taskName, tc.returnCode)
			// assert
			assert.Equal(tc.expected, fmt.Sprintf("%+v", payload))
		})
	}
}
