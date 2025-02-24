package gcs

import (
	"context"
	"io"
	"os"
	"path/filepath"

	"cloud.google.com/go/storage"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
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
	context       context.Context
	envVars       []EnvVar
	gcs           storage.Client
	clientOptions []option.ClientOption
	openFile      func(name string) (io.ReadCloser, error)
	createFile    func(name string) (io.WriteCloser, error)
}

// EnvVar defines an environment variable including information about a potential modification to the variable
type EnvVar struct {
	Name     string
	Value    string
	Modified bool
}

type gcsOption func(*gcsClient)

// WithEnvVars sets environment variables in gcsClient
func WithEnvVars(envVars []EnvVar) gcsOption {
	return func(g *gcsClient) {
		g.envVars = envVars
	}
}

// NewClient initializes the Google Cloud Storage client with the provided options
func NewClient(opts ...gcsOption) (Client, error) {
	ctx := context.Background()
	client := &gcsClient{
		context:    ctx,
		openFile:   openFileFromFS,
		createFile: createFileOnFS,
	}

	// Apply options
	for _, opt := range opts {
		opt(client)
	}

	client.prepareEnv()
	gcs, err := storage.NewClient(ctx, client.clientOptions...)
	if err != nil {
		return nil, errors.Wrapf(err, "bucket connection failed: %v", err)
	}
	client.gcs = *gcs
	return client, nil
}

// UploadFile uploads a file to a Google Cloud Storage bucket
func (cl *gcsClient) UploadFile(bucketID string, sourcePath string, targetPath string) error {
	target := cl.gcs.Bucket(bucketID).Object(targetPath).NewWriter(cl.context)
	log.Entry().Debugf("uploading %v to %v", sourcePath, targetPath)
	sourceFile, err := cl.openFile(sourcePath)
	if err != nil {
		return errors.Wrapf(err, "could not open source file: %v", err)
	}
	defer sourceFile.Close()

	if _, err := io.Copy(target, sourceFile); err != nil {
		return errors.Wrapf(err, "upload failed: %v", err)
	}

	if err := target.Close(); err != nil {
		return errors.Wrapf(err, "closing bucket failed: %v", err)
	}
	return nil
}

// DownloadFile downloads a file from a Google Cloud Storage bucket
func (cl *gcsClient) DownloadFile(bucketID string, sourcePath string, targetPath string) error {
	log.Entry().Debugf("downloading %v to %v\n", sourcePath, targetPath)
	gcsReader, err := cl.gcs.Bucket(bucketID).Object(sourcePath).NewReader(cl.context)
	if err != nil {
		return errors.Wrapf(err, "could not open source file from Google Cloud Storage bucket: %v", err)
	}

	targetWriter, err := cl.createFile(targetPath)
	if err != nil {
		return errors.Wrapf(err, "could not create target file: %v", err)
	}
	defer targetWriter.Close()

	if _, err := io.Copy(targetWriter, gcsReader); err != nil {
		return errors.Wrapf(err, "download failed: %v", err)
	}
	if err := gcsReader.Close(); err != nil {
		return errors.Wrapf(err, "closing bucket failed: %v", err)
	}
	return nil
}

// ListFiles lists all files in a specified Google Cloud Storage bucket
func (cl *gcsClient) ListFiles(bucketID string) ([]string, error) {
	fileNames := []string{}
	it := cl.gcs.Bucket(bucketID).Objects(cl.context, nil)
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
	if err := cl.cleanupEnv(); err != nil {
		return err
	}
	return nil
}

// prepareEnv sets required environment variables if they are not already set
func (cl *gcsClient) prepareEnv() {
	for key, env := range cl.envVars {
		cl.envVars[key].Modified = setenvIfEmpty(env.Name, env.Value)
	}
}

// cleanupEnv removes environment variables set by prepareEnv
func (cl *gcsClient) cleanupEnv() error {
	for _, env := range cl.envVars {
		if err := removeEnvIfPreviouslySet(env.Name, env.Modified); err != nil {
			return err
		}
	}
	return nil
}

// setenvIfEmpty sets an environment variable if it is not already set
func setenvIfEmpty(env, val string) bool {
	if len(os.Getenv(env)) == 0 {
		os.Setenv(env, val)
		return true
	}
	return false
}

// removeEnvIfPreviouslySet removes an environment variable if it was previously set by setenvIfEmpty
func removeEnvIfPreviouslySet(env string, previouslySet bool) error {
	if previouslySet {
		if err := os.Setenv(env, ""); err != nil {
			return err
		}
	}
	return nil
}

// TODO: consider refactoring to avoid keeping functions that are used only for unit/integration testing in the production code
// Below functions WithOpenFileFunction, WithCreateFileFunction and WithClientOptions are
// used only for unit/integration testing.

// WithOpenFileFunction sets the openFile function in gcsClient
func WithOpenFileFunction(openFile func(name string) (io.ReadCloser, error)) gcsOption {
	return func(g *gcsClient) {
		g.openFile = openFile
	}
}

// WithCreateFileFunction sets the createFile function in gcsClient
func WithCreateFileFunction(createFile func(name string) (io.WriteCloser, error)) gcsOption {
	return func(g *gcsClient) {
		g.createFile = createFile
	}
}

// WithClientOptions sets the Google Cloud Storage client options
func WithClientOptions(opts ...option.ClientOption) gcsOption {
	return func(g *gcsClient) {
		g.clientOptions = append(g.clientOptions, opts...)
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
