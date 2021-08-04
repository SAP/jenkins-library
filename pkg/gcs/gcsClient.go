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
)

// GCSClientInterface is an interface to mock GCSClient
type ClientInterface interface {
	UploadFile(bucketID string, sourcePath string, targetPath string) error
	DownloadFile(bucketID string, sourcePath string, targetPath string) error
	ListFiles(bucketID string) ([]string, error)
}

// GCSClient provides functions to interact with google cloud storage API
type gcsClient struct {
	context context.Context
	client  storage.Client
}

// Init intitializes the google cloud storage client
func NewClient(gcpJsonKeyFilePath string) (*gcsClient, error) {
	if gcpJsonKeyFilePath != "" {
		setenvIfEmpty("GOOGLE_APPLICATION_CREDENTIALS", gcpJsonKeyFilePath)
	} else {
		return nil, errors.New("GCP JSON Key file Path must not be empty")
	}
	context := context.Background()
	client, err := storage.NewClient(context)
	if err != nil {
		return nil, errors.Wrapf(err, "bucket connection failed: %v", err)
	}
	gcsClient := gcsClient{
		context: context,
		client:  *client,
	}
	return &gcsClient, nil
}

// UploadFile uploads a file into a google cloud storage bucket
func (g *gcsClient) UploadFile(bucketID string, sourcePath string, targetPath string) error {
	target := g.client.Bucket(bucketID).Object(targetPath).NewWriter(g.context)
	log.Entry().Debugf("uploading %v to %v\n", sourcePath, targetPath)
	sourceFile, err := os.Open(sourcePath)
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
func (g *gcsClient) DownloadFile(bucketID string, sourcePath string, targetPath string) error {
	log.Entry().Debugf("downloading %v to %v\n", sourcePath, targetPath)
	gcsReader, err := g.client.Bucket(bucketID).Object(sourcePath).NewReader(g.context)
	if err != nil {
		return errors.Wrapf(err, "could not open source file from a google cloud storage bucket: %v", err)
	}

	if err = os.MkdirAll(filepath.Dir(targetPath), os.ModePerm); err != nil {
		return errors.Wrapf(err, "could not create target path: %v", err)
	}

	targetWriter, err := os.Create(targetPath)
	if err != nil {
		return errors.Wrapf(err, "could not create target file: %v", err)
	}

	if err := g.copy(gcsReader, targetWriter); err != nil {
		return errors.Wrapf(err, "download failed: %v", err)
	}
	if err := gcsReader.Close(); err != nil {
		return errors.Wrapf(err, "closing bucket failed: %v", err)
	}
	return nil
}

// ListFiles lists all files in certain cumulus bucket
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

func setenvIfEmpty(env, val string) {
	if len(os.Getenv(env)) == 0 {
		os.Setenv(env, val)
	}
}
