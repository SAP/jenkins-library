package cmd

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type mockPutObjectAPI func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error)

func (m mockPutObjectAPI) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	return m(ctx, params, optFns...)
}

func TestRunAwsS3Upload(t *testing.T) {
	t.Parallel()

	t.Run("happy path", func(t *testing.T) {
		t.Parallel()
		// create temporary file
		f, err := os.CreateTemp("", "tmpfile-") // in Go version older than 1.17 you can use ioutil.TempFile
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		defer os.Remove(f.Name())
		data := []byte("test test test")
		if _, err := f.Write(data); err != nil {
			log.Fatal(err)
		}

		// initialization
		config := awsS3UploadOptions{
			FilePath: f.Name(),
		}
		client := mockS3Client
		// test
		err = runAwsS3Upload(&config, client(t, config.FilePath), "fooBucket")
		// assert
		assert.NoError(t, err)
	})

	t.Run("error path", func(t *testing.T) {
		t.Parallel()
		// initialization
		config := awsS3UploadOptions{
			FilePath: "nonExistingFilepath",
		}
		client := mockS3Client
		// test
		err := runAwsS3Upload(&config, client(t, config.FilePath), "fooBucket")
		// assert
		_, ok := err.(*fs.PathError)
		assert.True(t, ok)
	})

	t.Run("error bucket", func(t *testing.T) {
		t.Parallel()
		// create temporary file
		f, err := os.CreateTemp("", "tmpfile-") // in Go version older than 1.17 you can use ioutil.TempFile
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()
		defer os.Remove(f.Name())
		data := []byte("test test test")
		if _, err := f.Write(data); err != nil {
			log.Fatal(err)
		}

		// initialization
		config := awsS3UploadOptions{
			FilePath: f.Name(),
		}
		client := mockS3Client
		// test
		err = runAwsS3Upload(&config, client(t, config.FilePath), "errorBucket")
		// assert
		assert.EqualError(t, err, "expect fooBucket, got errorBucket")
	})
}

func mockS3Client(t *testing.T, fileName string) S3PutObjectAPI {
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
		if e, a := filepath.ToSlash(fileName), *params.Key; e != a {
			return nil, fmt.Errorf("expect %v, got %v", e, a)
		}
		if params.Body == nil {
			return nil, fmt.Errorf("expect Body / io.Reader not to be nil")
		}
		return &s3.PutObjectOutput{}, nil
	})
}
