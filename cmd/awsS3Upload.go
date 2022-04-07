package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/SAP/jenkins-library/pkg/command"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperutils"
	"github.com/SAP/jenkins-library/pkg/telemetry"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type awsS3UploadUtils interface {
	command.ExecRunner

	FileExists(filename string) (bool, error)

	// Add more methods here, or embed additional interfaces, or remove/replace as required.
	// The awsS3UploadUtils interface should be descriptive of your runtime dependencies,
	// i.e. include everything you need to be able to mock in tests.
	// Unit tests shall be executable in parallel (not depend on global state), and don't (re-)test dependencies.
}

type awsS3UploadUtilsBundle struct {
	*command.Command
	*piperutils.Files

	// Embed more structs as necessary to implement methods or interfaces you add to awsS3UploadUtils.
	// Structs embedded in this way must each have a unique set of methods attached.
	// If there is no struct which implements the method you need, attach the method to
	// awsS3UploadUtilsBundle and forward to the implementation of the dependency.
}

// S3PutObjectAPI defines the interface for the PutObject function.
// We use this interface to test the function using a mocked service.
type S3PutObjectAPI interface {
	PutObject(ctx context.Context,
		params *s3.PutObjectInput,
		optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)
}

// PutFile uploads a file to an Amazon Simple Storage Service (Amazon S3) bucket
// Inputs:
//     c is the context of the method call, which includes the AWS Region
//     api is the interface that defines the method call
//     input defines the input arguments to the service call.
// Output:
//     If success, a PutObjectOutput object containing the result of the service call and nil
//     Otherwise, nil and an error from the call to PutObject
func PutFile(c context.Context, api S3PutObjectAPI, input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	return api.PutObject(c, input)
}

func newAwsS3UploadUtils() awsS3UploadUtils {
	utils := awsS3UploadUtilsBundle{
		Command: &command.Command{},
		Files:   &piperutils.Files{},
	}
	// Reroute command output to logging framework
	utils.Stdout(log.Writer())
	utils.Stderr(log.Writer())
	return &utils
}

type AWSCredentials struct {
	access_key_id     string
	bucket            string
	secret_access_key string
	region            string
}

func awsS3Upload(configOptions awsS3UploadOptions, telemetryData *telemetry.CustomData) {
	// Utils can be used wherever the command.ExecRunner interface is expected.
	// It can also be used for example as a mavenExecRunner.
	utils := newAwsS3UploadUtils()

	// For HTTP calls import  piperhttp "github.com/SAP/jenkins-library/pkg/http"
	// and use a  &piperhttp.Client{} in a custom system
	// Example: step checkmarxExecuteScan.go

	//Prepare Credentials
	data, err := ioutil.ReadFile(configOptions.JSONKeyFilePath)
	if err != nil {
		fmt.Print(err)
	}

	var obj AWSCredentials

	err = json.Unmarshal(data, &obj)
	if err != nil {
		fmt.Println("error:", err)
	}

	//Set environment variables which are needed to initialize S3 Client
	setenvIfEmpty("AWS_REGION", obj.region)
	setenvIfEmpty("AWS_ACCESS_KEY_ID", obj.access_key_id)
	setenvIfEmpty("AWS_SECRET_ACCESS_KEY", obj.secret_access_key)

	//Initialize S3 Client
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic("configuration error, " + err.Error())
	}
	client := s3.NewFromConfig(cfg)

	// Error situations should be bubbled up until they reach the line below which will then stop execution
	// through the log.Entry().Fatal() call leading to an os.Exit(1) in the end.
	err = runAwsS3Upload(&configOptions, telemetryData, utils, client, obj.bucket)
	if err != nil {
		log.Entry().WithError(err).Fatal("step execution failed")
	}
}

func runAwsS3Upload(configOptions *awsS3UploadOptions, telemetryData *telemetry.CustomData, utils awsS3UploadUtils, client S3PutObjectAPI, bucket string) error {
	//Open File
	file, err := os.Open(configOptions.FilePath)

	if err != nil {
		fmt.Println("Unable to open file " + configOptions.FilePath)
		return err
	}

	defer file.Close()

	//Intitialize S3 PutObjectInput
	input := &s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &configOptions.FilePath,
		Body:   file,
	}

	//Upload File
	_, err = PutFile(context.TODO(), client, input)
	if err != nil {
		fmt.Println("Got error uploading file:")
		fmt.Println(err)
	}

	return err
}

func setenvIfEmpty(env, val string) bool {
	if len(os.Getenv(env)) == 0 {
		os.Setenv(env, val)
		return true
	}
	return false
}
