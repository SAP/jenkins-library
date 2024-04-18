package gcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/pkg/errors"
)

const api_url = "https://pubsub.googleapis.com/v1/projects/%s/topics/%s:publish"

// https://pkg.go.dev/cloud.google.com/go/pubsub#Message
type EventMessage struct {
	Data []byte `json:"data"`
}

type Event struct {
	Messages []EventMessage `json:"messages"`
}

func Publish(projectNumber string, topic string, token string, data []byte) error {
	ctx := context.Background()

	// build event
	event := Event{
		Messages: []EventMessage{{
			Data: data,
		}},
	}

	// marshal event
	eventBytes, err := json.Marshal(event)
	if err != nil {
		return errors.Wrap(err, "failed to marshal event")
	}

	// create request
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf(api_url, projectNumber, topic), bytes.NewReader(eventBytes))
	if err != nil {
		return errors.Wrap(err, "failed to create request")
	}

	// add headers
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	// send request
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return errors.Wrap(err, "failed to send request")
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid status code: %v", response.StatusCode)
	}

	//TODO: read response & messageIds

	return nil
}
