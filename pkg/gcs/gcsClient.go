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
	client        storage.Client
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

// NewClient initializes the Google Cloud Storage client with the provided options
func NewClient(opts ...gcsOption) (*gcsClient, error) {
	var (
		defaultOpenFile   = openFileFromFS
		defaultCreateFile = createFileOnFS
	)

	ctx := context.Background()
	gcsClient := &gcsClient{
		context:    ctx,
		openFile:   defaultOpenFile,
		createFile: defaultCreateFile,
	}

	// Apply options
	for _, opt := range opts {
		opt(gcsClient)
	}

	gcsClient.prepareEnv()
	client, err := storage.NewClient(ctx, gcsClient.clientOptions...)
	if err != nil {
		return nil, errors.Wrapf(err, "bucket connection failed: %v", err)
	}
	gcsClient.client = *client
	return gcsClient, nil
}

// UploadFile uploads a file to a Google Cloud Storage bucket
func (g *gcsClient) UploadFile(bucketID string, sourcePath string, targetPath string) error {
	target := g.client.Bucket(bucketID).Object(targetPath).NewWriter(g.context)
	log.Entry().Debugf("uploading %v to %v", sourcePath, targetPath)
	sourceFile, err := g.openFile(sourcePath)
	if err != nil {
		return errors.Wrapf(err, "could not open source file: %v", err)
	}
	defer sourceFile.Close()

	if err := g.copy(sourceFile, target); err != nil {
		return errors.Wrapf(err, "upload failed: %v", err)
	}

	if err := target.Close(); err != nil {
		return errors.Wrapf(err, "closing bucket failed: %v", err)
	}
	return nil
}

// DownloadFile downloads a file from a Google Cloud Storage bucket
func (g *gcsClient) DownloadFile(bucketID string, sourcePath string, targetPath string) error {
	log.Entry().Debugf("downloading %v to %v\n", sourcePath, targetPath)
	gcsReader, err := g.client.Bucket(bucketID).Object(sourcePath).NewReader(g.context)
	if err != nil {
		return errors.Wrapf(err, "could not open source file from Google Cloud Storage bucket: %v", err)
	}

	targetWriter, err := g.createFile(targetPath)
	if err != nil {
		return errors.Wrapf(err, "could not create target file: %v", err)
	}
	defer targetWriter.Close()

	if err := g.copy(gcsReader, targetWriter); err != nil {
		return errors.Wrapf(err, "download failed: %v", err)
	}
	if err := gcsReader.Close(); err != nil {
		return errors.Wrapf(err, "closing bucket failed: %v", err)
	}
	return nil
}

// ListFiles lists all files in a specified Google Cloud Storage bucket
func (g *gcsClient) ListFiles(bucketID string) ([]string, error) {
	fileNames := []string{}
	it := g.client.Bucket(bucketID).Objects(g.context, nil)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		fileNames = append(fileNames, attrs.Name)
	}
	return fileNames, nil
}

// Close closes the client and removes previously set environment variables
func (g *gcsClient) Close() error {
	if err := g.client.Close(); err != nil {
		return err
	}
	if err := g.cleanupEnv(); err != nil {
		return err
	}
	return nil
}

// copy copies data from the source reader to the target writer
func (g *gcsClient) copy(source io.Reader, target io.Writer) error {
	if _, err := io.Copy(target, source); err != nil {
		return err
	}
	return nil
}

// prepareEnv sets required environment variables if they are not already set
func (g *gcsClient) prepareEnv() {
	for key, env := range g.envVars {
		g.envVars[key].Modified = setenvIfEmpty(env.Name, env.Value)
	}
}

// cleanupEnv removes environment variables set by prepareEnv
func (g *gcsClient) cleanupEnv() error {
	for _, env := range g.envVars {
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
