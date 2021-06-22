package piperutils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopyData(t *testing.T) {
	runInTempDir(t, "copying file succeeds", "dir1", func(t *testing.T) {
		srcName := "testFile"
		src, err := os.OpenFile(srcName, os.O_CREATE, 0700)
		if err != nil {
			t.Fatal("Failed to create src file")
		}
		_, err = src.Write([]byte{byte(1), byte(2), byte(3)})
		src.Close()
		src, err = os.OpenFile(srcName, os.O_CREATE, 0700)

		dstName := "testFile2"
		dst, err := os.OpenFile(dstName, os.O_CREATE, 0700)
		if err != nil {
			t.Fatal("Failed to create dst file")
		}

		result, err := CopyData(dst, src)

		assert.NoError(t, err, "Didn't expert error but got one")
		assert.Equal(t, int64(3), result, "Expected true but got false")
	})
	runInTempDir(t, "copying file succeeds", "dir2", func(t *testing.T) {
		srcName := "testFile"
		src, err := os.OpenFile(srcName, os.O_CREATE, 0700)
		if err != nil {
			t.Fatal("Failed to create src file")
		}
		data := make([]byte, 300)
		for i := 0; i < 300; i++ {
			data[i] = byte(i)
		}
		_, err = src.Write(data)
		src.Close()
		src, err = os.OpenFile(srcName, os.O_CREATE, 0700)

		dstName := "testFile2"
		dst, err := os.OpenFile(dstName, os.O_CREATE, 0700)
		if err != nil {
			t.Fatal("Failed to create dst file")
		}

		result, err := CopyData(dst, src)

		assert.NoError(t, err, "Didn't expert error but got one")
		assert.Equal(t, int64(300), result, "Expected true but got false")
	})
	runInTempDir(t, "copying file succeeds", "dir3", func(t *testing.T) {
		srcName := "testFileExcl"
		src, err := os.OpenFile(srcName, os.O_CREATE, 0700)
		if err != nil {
			t.Fatalf("Failed to create src file %v", err)
		}
		data := make([]byte, 300)
		for i := 0; i < 300; i++ {
			data[i] = byte(i)
		}
		_, err = src.Write(data)
		src.Close()
		src, err = os.OpenFile(srcName, os.O_WRONLY, 0700)

		dstName := "testFile2"
		dst, err := os.OpenFile(dstName, os.O_CREATE, 0700)
		if err != nil {
			t.Fatal("Failed to create dst file")
		}

		result, err := CopyData(dst, src)

		assert.Error(t, err, "Expected error but got none")
		assert.Equal(t, int64(0), result, "Expected true but got false")
	})
	runInTempDir(t, "copying file succeeds", "dir4", func(t *testing.T) {
		srcName := "testFileExcl"
		src, err := os.OpenFile(srcName, os.O_CREATE, 0700)
		if err != nil {
			t.Fatalf("Failed to create src file %v", err)
		}
		data := make([]byte, 300)
		for i := 0; i < 300; i++ {
			data[i] = byte(i)
		}
		_, err = src.Write(data)
		src.Close()
		src, err = os.OpenFile(srcName, os.O_CREATE, 0700)

		dstName := "testFile2"
		dst, err := os.OpenFile(dstName, os.O_CREATE, 0700)
		if err != nil {
			t.Fatal("Failed to create dst file")
		}
		dst.Close()
		dst, err = os.OpenFile(dstName, os.O_RDONLY, 0700)

		result, err := CopyData(dst, src)

		assert.Error(t, err, "Expected error but got none")
		assert.Equal(t, int64(0), result, "Expected true but got false")
	})
}
