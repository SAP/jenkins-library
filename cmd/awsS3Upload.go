package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/pkg/errors"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3PutObjectAPI defines the interface for the PutObject function.
// We use this interface to test the function using a mocked service.
type S3PutObjectAPI interface {
	PutObject(ctx context.Context,
		params *s3.PutObjectInput,
		optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

// PutFile uploads a file to an AWS S3 bucket
// The function needs a context (including the AWS Region) and a PutObjectInput for the service call
// The return value is a PutObjectOutput with the result of the upload
func PutFile(c context.Context, api S3PutObjectAPI, input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	return api.PutObject(c, input)
}

// Struct to store the AWS credentials from a specified JSON string
type awsCredentials struct {
	AwsAccessKeyID     string `json:"access_key_id"`
	Bucket             string `json:"bucket"`
	AwsSecretAccessKey string `json:"secret_access_key"`
	AwsRegion          string `json:"region"`
}

func awsS3Upload(configOptions awsS3UploadOptions, telemetryData *telemetry.CustomData) {
	// Prepare Credentials
	log.Entry().Infoln("Start reading AWS Credentials")
	var obj awsCredentials

	err := json.Unmarshal([]byte(configOptions.JSONCredentialsAWS), &obj)
	if err != nil {
		log.Entry().
			WithError(err).
			Fatal("Could not read JSONCredentialsAWS")
	}

	// Set environment variables which are needed to initialize S3 Client
	log.Entry().Infoln("Successfully read AWS Credentials. Setting up environment variables")
	awsRegionSet := setenvIfEmpty("AWS_REGION", obj.AwsRegion)
	awsAccessKeyIDSet := setenvIfEmpty("AWS_ACCESS_KEY_ID", obj.AwsAccessKeyID)
	awsSecretAccessKeySet := setenvIfEmpty("AWS_SECRET_ACCESS_KEY", obj.AwsSecretAccessKey)

	defer removeEnvIfPreviouslySet("AWS_REGION", awsRegionSet)
	defer removeEnvIfPreviouslySet("AWS_ACCESS_KEY_ID", awsAccessKeyIDSet)
	defer removeEnvIfPreviouslySet("AWS_SECRET_ACCESS_KEY", awsSecretAccessKeySet)

	// Initialize S3 Client
	log.Entry().Infoln("Loading Configuration for S3 Client")
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Entry().
			WithError(err).
			Fatal("AWS Client Configuration failed")
	}
	client := s3.NewFromConfig(cfg)

	err = runAwsS3Upload(&configOptions, client, obj.Bucket)
	if err != nil {
		log.Entry().WithError(err).Fatal("Step execution failed")
	}
}

func runAwsS3Upload(configOptions *awsS3UploadOptions, client S3PutObjectAPI, bucket string) error {
	// Iterate through directories
	err := filepath.Walk(configOptions.FilePath, func(currentFilePath string, f os.FileInfo, err error) error {
		// Handle Failure to prevent panic (e.g. in case of an invalid filepath)
		if err != nil {
			return fmt.Errorf("failed to access path: '%v', error: %w", currentFilePath, err)
		}
		// Skip directories, only upload files
		if !f.IsDir() {
			log.Entry().Infof("Current target path is: '%v'", currentFilePath)

			// Open File
			currentFile, err := os.Open(currentFilePath)
			if err != nil {
				return errors.Wrapf(err, "failed to open file: '%v'", currentFilePath)
			}
			defer currentFile.Close()

			// AWS SDK needs UNIX file paths to automatically create directories
			key := filepath.ToSlash(currentFilePath)

			// Intitialize S3 PutObjectInput
			inputObject := &s3.PutObjectInput{
				Bucket: &bucket,
				Key:    &key,
				Body:   currentFile,
			}

			// Upload File
			log.Entry().Infof("Start upload of file '%v'", currentFilePath)
			_, err = PutFile(context.TODO(), client, inputObject)
			if err != nil {
				return errors.Wrapf(err, "failed to upload file '%v'", currentFilePath)
			}

			log.Entry().Infof("Upload of file '%v' was successful!", currentFilePath)
			return nil
		}
		return nil
	})

	if err != nil {
		return err
	}
	log.Entry().Infoln("Upload has successfully finished!")
	return err
}

// Function to set environment variables if they are not already set
func setenvIfEmpty(env, val string) bool {
	if len(os.Getenv(env)) == 0 {
		os.Setenv(env, val)
		return true
	}
	return false
}

// Function to remove environment variables if they are set
func removeEnvIfPreviouslySet(env string, previouslySet bool) {
	if previouslySet {
		os.Setenv(env, "")
	}
}
