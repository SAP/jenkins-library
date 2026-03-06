//go:build unit

package gcs

import (
	"context"
	"errors"
	"testing"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

// Mock functions to replace real GCS client creation in tests
var (
	mockInitWithKeyFile = func(ctx context.Context, keyFile string, opts ...option.ClientOption) (*storage.Client, error) {
		if keyFile == "bad-key" {
			return nil, errors.New("bad key file")
		}
		return &storage.Client{}, nil
	}
	mockInitWithToken = func(ctx context.Context, token string, opts ...option.ClientOption) (*storage.Client, error) {
		if token == "bad-token" {
			return nil, errors.New("bad token")
		}
		return &storage.Client{}, nil
	}
)

func TestInitGcsClient(t *testing.T) {
	// Patch the real functions with mocks
	origKeyFile := initWithKeyFile
	origToken := initWithToken
	initWithKeyFile = mockInitWithKeyFile
	initWithToken = mockInitWithToken
	defer func() {
		initWithKeyFile = origKeyFile
		initWithToken = origToken
	}()

	ctx := context.Background()

	t.Run("no credentials", func(t *testing.T) {
		_, err := initGcsClient(ctx, "", "")
		if err == nil {
			t.Error("expected error when no credentials provided")
		}
	})

	t.Run("token only", func(t *testing.T) {
		client, err := initGcsClient(ctx, "", "good-token")
		if err != nil || client == nil {
			t.Errorf("expected success with token, got err: %v", err)
		}
	})

	t.Run("keyFile only", func(t *testing.T) {
		client, err := initGcsClient(ctx, "good-key", "")
		if err != nil || client == nil {
			t.Errorf("expected success with keyFile, got err: %v", err)
		}
	})

	t.Run("both, token succeeds", func(t *testing.T) {
		client, err := initGcsClient(ctx, "good-key", "good-token")
		if err != nil || client == nil {
			t.Errorf("expected success with both, got err: %v", err)
		}
	})

	t.Run("both, token fails, keyFile succeeds", func(t *testing.T) {
		client, err := initGcsClient(ctx, "good-key", "bad-token")
		if err != nil || client == nil {
			t.Errorf("expected fallback to keyFile, got err: %v", err)
		}
	})

	t.Run("both, both fail", func(t *testing.T) {
		_, err := initGcsClient(ctx, "bad-key", "bad-token")
		if err == nil {
			t.Error("expected error when both token and keyFile fail")
		}
	})
}
