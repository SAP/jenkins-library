package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/SAP/jenkins-library/pkg/mock"
	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type awsS3UploadMockUtils struct {
	*mock.ExecMockRunner
	*mock.FilesMock
}

func newAwsS3UploadTestsUtils() awsS3UploadMockUtils {
	utils := awsS3UploadMockUtils{
		ExecMockRunner: &mock.ExecMockRunner{},
		FilesMock:      &mock.FilesMock{},
	}
	return utils
}

type mockPutObjectAPI func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)

func (m mockPutObjectAPI) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return m(ctx, params, optFns...)
}

func TestRunAwsS3Upload(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// init
		config := awsS3UploadOptions{
			FilePath: filepath.Join("testdata", t.Name()+"_test.txt"),
		}

		client := func(t *testing.T) S3PutObjectAPI {
			return mockPutObjectAPI(func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
				t.Helper()
				if params.Bucket == nil {
					return nil, fmt.Errorf("expect bucket to not be nil")
				}
				if e, a := "fooBucket", *params.Bucket; e != a {
					return nil, fmt.Errorf("expect %v, got %v", e, a)
				}
				if params.Key == nil {
					return nil, fmt.Errorf("expect key to not be nil")
				}
				if e, a := filepath.ToSlash(filepath.Join("testdata", t.Name()+"_test.txt")), *params.Key; e != a {
					return nil, fmt.Errorf("expect %v, got %v", e, a)
				}
				if params.Body == nil {
					return nil, fmt.Errorf("expect Body / io.Reader not to be nil")
				}
				return &s3.PutObjectOutput{}, nil
			})
		}

		utils := newAwsS3UploadTestsUtils()

		// test
		err := runAwsS3Upload(&config, nil, utils, client(t), "fooBucket")

		// assert
		assert.NoError(t, err)
	})

	t.Run("no Path", func(t *testing.T) {
		t.Parallel()
		// init
		config := awsS3UploadOptions{}

		client := func(t *testing.T) S3PutObjectAPI {
			return mockPutObjectAPI(func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
				t.Helper()
				if params.Bucket == nil {
					return nil, fmt.Errorf("expect bucket to not be nil")
				}
				if e, a := "fooBucket", *params.Bucket; e != a {
					return nil, fmt.Errorf("expect %v, got %v", e, a)
				}
				if params.Key == nil {
					return nil, fmt.Errorf("expect key to not be nil")
				}
				if e, a := filepath.ToSlash(filepath.Join("testdata", t.Name()+"_test.txt")), *params.Key; e != a {
					return nil, fmt.Errorf("expect %v, got %v", e, a)
				}
				if params.Body == nil {
					return nil, fmt.Errorf("expect Body / io.Reader not to be nil")
				}
				return &s3.PutObjectOutput{}, nil
			})
		}

		utils := newAwsS3UploadTestsUtils()

		// test
		err := runAwsS3Upload(&config, nil, utils, client(t), "fooBucket")

		// assert
		assert.EqualError(t, err, "Empty File Path")
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()

		config := awsS3UploadOptions{
			FilePath: filepath.Join("testdata", t.Name()+"_test.txt"),
		}

		client := func(t *testing.T) S3PutObjectAPI {
			return mockPutObjectAPI(func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
				t.Helper()
				if params.Bucket == nil {
					return nil, fmt.Errorf("expect bucket to not be nil")
				}
				if e, a := "fooBucket", *params.Bucket; e != a {
					return nil, fmt.Errorf("expect %v, got %v", e, a)
				}
				return &s3.PutObjectOutput{}, nil
			})
		}

		utils := newAwsS3UploadTestsUtils()

		// test
		err := runAwsS3Upload(&config, nil, utils, client(t), "errorBucket")

		// assert
		assert.EqualError(t, err, "expect fooBucket, got errorBucket")
	})
}
