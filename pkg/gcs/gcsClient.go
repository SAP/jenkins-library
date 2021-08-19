package gcs

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/pkg/errors"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// GCSClientInterface is an interface to mock GCSClient
type ClientInterface interface {
	UploadFile(sourcePath string) error
	DownloadFile(sourcePath string) error
	ListFiles() ([]string, error)
	Close() error
}

// GCSClient provides functions to interact with google cloud storage API
type gcsClient struct {
	envVars      []EnvVar
	context      context.Context
	client       storage.Client
	openFile     func(name string) (io.ReadCloser, error)
	createFile   func(name string) (io.WriteCloser, error)
	bucketID     string
	targetFolder string
}

// EnvVar defines an  environment variable incl. information about a potential modification to the variable
type EnvVar struct {
	Name     string
	Value    string
	Modified bool
}

// Init intitializes the google cloud storage client
func NewClient(envVars []EnvVar, openFile func(name string) (io.ReadCloser, error), createFile func(name string) (io.WriteCloser, error),
	bucketID string, targetFolder string, opts ...option.ClientOption) (*gcsClient, error) {
	if bucketID == "" {
		return nil, errors.New("bucketID mustn't be empty")
	}
	ctx := context.Background()
	gcsClient := gcsClient{
		context:      ctx,
		envVars:      envVars,
		openFile:     openFile,
		createFile:   createFile,
		bucketID:     bucketID,
		targetFolder: targetFolder,
	}
	gcsClient.prepareEnv()
	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, errors.Wrapf(err, "bucket connection failed: %v", err)
	}
	gcsClient.client = *client
	return &gcsClient, nil
}

// prepareEnv sets required environment variables in case they are not set yet
func (g *gcsClient) prepareEnv() {
	for key, env := range g.envVars {
		g.envVars[key].Modified = setenvIfEmpty(env.Name, env.Value)
	}
}

// cleanupEnv removes environment variables set by PrepareEnv
func (g *gcsClient) cleanupEnv() error {
	for _, env := range g.envVars {
		if err := removeEnvIfPreviouslySet(env.Name, env.Modified); err != nil {
			return err
		}
	}
	return nil
}

// UploadFile uploads a file into a google cloud storage bucket
func (g *gcsClient) UploadFile(sourcePath string) error {
	targetPath := strings.Trim(g.targetFolder, "/") + "/" + strings.TrimPrefix(sourcePath, "/")
	target := g.client.Bucket(g.bucketID).Object(targetPath).NewWriter(g.context)
	log.Entry().Debugf("uploading %v to %v\n", sourcePath, targetPath)
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

func (g *gcsClient) copy(source io.Reader, target io.Writer) error {
	if _, err := io.Copy(target, source); err != nil {
		return err
	}
	return nil
}

// DownloadFile downloads a file from a google cloud storage bucket
func (g *gcsClient) DownloadFile(sourcePath string) error {
	targetPath := strings.TrimPrefix(sourcePath, "/")
	log.Entry().Debugf("downloading %v to %v\n", sourcePath, targetPath)
	gcsReader, err := g.client.Bucket(g.bucketID).Object(sourcePath).NewReader(g.context)
	if err != nil {
		return errors.Wrapf(err, "could not open source file from a google cloud storage bucket: %v", err)
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

// ListFiles lists all files in certain cumulus bucket
func (g *gcsClient) ListFiles() ([]string, error) {
	fileNames := []string{}
	it := g.client.Bucket(g.bucketID).Objects(g.context, nil)
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

func (g *gcsClient) Close() error {
	if err := g.client.Close(); err != nil {
		return err
	}
	if err := g.cleanupEnv(); err != nil {
		return err
	}
	return nil
}

func setenvIfEmpty(env, val string) bool {
	if len(os.Getenv(env)) == 0 {
		os.Setenv(env, val)
		return true
	}
	return false
}

func removeEnvIfPreviouslySet(env string, previouslySet bool) error {
	if previouslySet {
		if err := os.Setenv(env, ""); err != nil {
			return err
		}
	}
	return nil
}

func OpenFileFromFS(name string) (io.ReadCloser, error) {
	return os.Open(name)
}

func CreateFileOnFS(name string) (io.WriteCloser, error) {
	if err := os.MkdirAll(filepath.Dir(name), os.ModePerm); err != nil {
		return nil, err
	}
	return os.Create(name)
}
