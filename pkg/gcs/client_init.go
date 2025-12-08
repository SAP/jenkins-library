package gcs

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/storage"
	"github.com/SAP/jenkins-library/pkg/log"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"
)

func initGcsClient(ctx context.Context, keyFile, token string, opts ...option.ClientOption) (client *storage.Client, err error) {
	switch {
	case keyFile == "" && token == "":
		return nil, errors.New("please provide either the keyFile or token")
	case keyFile == "":
		log.Entry().Debug("Authenticating with token")
		if client, err = initWithToken(ctx, token, opts...); err != nil {
			return nil, fmt.Errorf("token auth failed: %w", err)
		}
		log.Entry().Debug("Successfully initialized GCS client")
		return client, nil
	case token == "":
		log.Entry().Debug("Authenticating with JSON key file")
		if client, err = initWithKeyFile(ctx, keyFile, opts...); err != nil {
			return nil, fmt.Errorf("key file auth failed: %w", err)
		}
		log.Entry().Debug("Successfully initialized GCS client")
		return client, nil
	default: // Both keyFile and token are provided - token takes precedence
		log.Entry().Debug("Authenticating with token")
		client, errTok := initWithToken(ctx, token, opts...)
		if errTok == nil {
			log.Entry().Debug("Successfully initialized GCS client")
			return client, nil
		}
		log.Entry().Warn("Failed to authenticate with token, falling back to key file")
		client, errKey := initWithKeyFile(ctx, keyFile, opts...)
		if errKey == nil {
			log.Entry().Debug("Successfully initialized GCS client")
			return client, nil
		}
		return nil, fmt.Errorf("failed to authenticate with both token (%v) and key file (%v)", errTok, errKey)
	}
}

// This is the old way of initializing the GCS client, kept for backward compatibility
// OLD: Keyfile takes precedence over token, if both are provided
// NEW: Token takes precedence over keyfile, if both are provided
func initGcsClientLegacy(ctx context.Context, keyFile, token string, opts ...option.ClientOption) (client *storage.Client, err error) {
	switch {
	case keyFile == "" && token == "":
		return nil, errors.New("please provide either the keyFile or token")
	case keyFile == "":
		log.Entry().Debug("Authenticating with token")
		if client, err = initWithToken(ctx, token, opts...); err != nil {
			return nil, fmt.Errorf("token auth failed: %w", err)
		}
	default: // Keyfile not empty
		log.Entry().Debug("Authenticating with JSON key file")
		if client, err = initWithKeyFile(ctx, keyFile, opts...); err != nil {
			if token == "" {
				return nil, fmt.Errorf("key file auth failed: %w", err)
			}
			log.Entry().Debug("Falling back to token authentication")
			if client, err = initWithToken(ctx, token, opts...); err != nil {
				return nil, fmt.Errorf("token auth failed: %w", err)
			}
		}
	}

	log.Entry().Debug("Successfully initialized GCS client")
	return client, nil
}

func initWithKeyFile(ctx context.Context, keyFile string, opts ...option.ClientOption) (*storage.Client, error) {
	o := append([]option.ClientOption{option.WithCredentialsFile(keyFile)}, opts...)
	return storage.NewClient(ctx, o...)
}

func initWithToken(ctx context.Context, token string, opts ...option.ClientOption) (*storage.Client, error) {
	o := append([]option.ClientOption{option.WithTokenSource(oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token}))}, opts...)
	return storage.NewClient(ctx, o...)
}
