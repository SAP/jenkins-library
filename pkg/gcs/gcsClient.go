package gcs

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"io"
	"os"
	"path/filepath"
)

// Client is an interface to mock gcsClient
type Client interface {
	UploadFile(bucketID string, sourcePath string, targetPath string) error
	DownloadFile(bucketID string, sourcePath string, targetPath string) error
	ListFiles(bucketID string) ([]string, error)
	Close() error
}

// gcsClient provides functions to interact with Google Cloud Storage API
type gcsClient struct {
	ctx        context.Context
	gcs        storage.Client
	gcsOptions []option.ClientOption
	openFile   func(name string) (io.ReadCloser, error)
	createFile func(name string) (io.WriteCloser, error)
}

type clientOptions func(*gcsClient)

// NewClient initializes the Google Cloud Storage client with the provided options
func NewClient(keyFile, token string, opts ...clientOptions) (Client, error) {
	ctx := context.Background()
	client := &gcsClient{
		ctx:        ctx,
		openFile:   openFileFromFS,
		createFile: createFileOnFS,
	}

	// Apply options
	for _, opt := range opts {
		opt(client)
	}

	gcs, err := initGcsClient(ctx, keyFile, token, client.gcsOptions...)
	if err != nil {
		return nil, err
	}

	client.gcs = *gcs
	return client, nil
}

func (cl *gcsClient) UploadFile(bucketID string, sourcePath string, targetPath string) error {
	sourcePath = filepath.Clean(sourcePath)
	log.Entry().Debugf("Uploading %v to %v\n", sourcePath, targetPath)

	sourceFile, err := cl.openFile(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	target := cl.gcs.Bucket(bucketID).Object(targetPath).NewWriter(cl.ctx)
	defer target.Close()

	task := func(ctx context.Context) error {
		if _, err = io.Copy(target, sourceFile); err != nil {
			return fmt.Errorf("upload failed: %w", err)
		}
		return nil
	}

	return retryWithLogging(cl.ctx, log.Entry(), task, initialBackoff, maxRetries, retryMultiplier)
}

// DownloadFile downloads a file from a Google Cloud Storage bucket
func (cl *gcsClient) DownloadFile(bucketID, sourcePath, targetPath string) error {
	targetPath = filepath.Clean(targetPath)
	log.Entry().Debugf("Downloading %v to %v\n", sourcePath, targetPath)

	target, err := cl.createFile(targetPath)
	if err != nil {
		return errors.Wrapf(err, "could not create target file: %v", err)
	}
	defer target.Close()

	source, err := cl.gcs.Bucket(bucketID).Object(sourcePath).NewReader(cl.ctx)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer source.Close()

	task := func(ctx context.Context) error {
		if _, err = io.Copy(target, source); err != nil {
			return fmt.Errorf("download failed: %w", err)
		}
		return nil
	}

	return retryWithLogging(cl.ctx, log.Entry(), task, initialBackoff, maxRetries, retryMultiplier)
}

// ListFiles lists all files in a specified Google Cloud Storage bucket
func (cl *gcsClient) ListFiles(bucketID string) ([]string, error) {
	fileNames := []string{}
	it := cl.gcs.Bucket(bucketID).Objects(cl.ctx, nil)
	for {
		attrs, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, errors.Wrap(err, "could not list files")
		}
		fileNames = append(fileNames, attrs.Name)
	}
	return fileNames, nil
}

// Close closes the client and removes previously set environment variables
func (cl *gcsClient) Close() error {
	if err := cl.gcs.Close(); err != nil {
		return err
	}

	return nil
}

// TODO: consider refactoring to avoid keeping functions that are used only for integration testing in the production code
// Below functions WithOpenFileFunction, WithCreateFileFunction and WithGCSClientOptions are
// used only for unit/integration testing.

// WithOpenFileFunction sets the openFile function in gcsClient
func WithOpenFileFunction(openFile func(name string) (io.ReadCloser, error)) clientOptions {
	return func(g *gcsClient) {
		g.openFile = openFile
	}
}

// WithCreateFileFunction sets the createFile function in gcsClient
func WithCreateFileFunction(createFile func(name string) (io.WriteCloser, error)) clientOptions {
	return func(g *gcsClient) {
		g.createFile = createFile
	}
}

// WithGCSClientOptions sets the Google Cloud Storage client options
func WithGCSClientOptions(opts ...option.ClientOption) clientOptions {
	return func(g *gcsClient) {
		g.gcsOptions = append(g.gcsOptions, opts...)
	}
}

// openFileFromFS and createFileOnFS functions are existing just because of the integration tests
// TODO: find a better way to mock filesystem operations in the integration tests
// openFileFromFS opens a file from the filesystem
func openFileFromFS(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

// createFileOnFS creates a file on the filesystem
func createFileOnFS(name string) (io.WriteCloser, error) {
	if err := os.MkdirAll(filepath.Dir(name), os.ModePerm); err != nil {
		return nil, err
	}
	return os.Create(name)
}

// NewClientLegacy is also still here because of integration tests
func NewClientLegacy(opts ...clientOptions) (Client, error) {
	ctx := context.Background()
	client := &gcsClient{
		ctx:        ctx,
		openFile:   openFileFromFS,
		createFile: createFileOnFS,
	}

	// Apply options
	for _, opt := range opts {
		opt(client)
	}

	gcs, err := storage.NewClient(ctx, client.gcsOptions...)
	if err != nil {
		return nil, err
	}

	client.gcs = *gcs
	return client, nil
}
