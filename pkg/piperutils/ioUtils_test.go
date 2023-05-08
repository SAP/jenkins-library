//go:build unit
// +build unit

package piperutils

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopyData(t *testing.T) {
	runInTempDir(t, "copying file succeeds small", func(t *testing.T) {
		srcName := "testFileSrc"
		src, err := os.OpenFile(srcName, os.O_CREATE|os.O_RDWR, 0700)
		if err != nil {
			t.Fatal("Failed to create src file")
		}
		data := make([]byte, 3)
		data[0] = byte(32)
		data[1] = byte(42)
		data[2] = byte(52)
		_, err = src.Write(data)
		if err != nil {
			t.Fatalf("Failed to write data to src file: %v", err)
		}
		src.Close()
		src, err = os.OpenFile(srcName, os.O_CREATE|os.O_RDWR, 0700)
		if err != nil {
			t.Fatal("Failed to reopen src file")
		}

		dstName := "testFileTgt"
		dst, err := os.OpenFile(dstName, os.O_CREATE|os.O_RDWR, 0700)
		if err != nil {
			t.Fatal("Failed to create dst file")
		}

		result, err := CopyData(dst, src)
		src.Close()
		dst.Close()
		dst, err = os.OpenFile(dstName, os.O_CREATE|os.O_RDWR, 0700)
		dataRead := make([]byte, 3)
		dst.Read(dataRead)
		dst.Close()

		assert.NoError(t, err, "Didn't expect error but got one")
		assert.Equal(t, int64(3), result, "Expected true but got false")
		assert.Equal(t, data, dataRead, "data written %v is different to data read %v")
	})
	runInTempDir(t, "copying file succeeds larger", func(t *testing.T) {
		srcName := "testFile"
		src, err := os.OpenFile(srcName, os.O_CREATE|os.O_RDWR, 0700)
		if err != nil {
			t.Fatal("Failed to create src file")
		}
		data := make([]byte, 300)
		for i := 0; i < 300; i++ {
			data[i] = byte(i)
		}
		_, err = src.Write(data)
		src.Close()
		src, err = os.OpenFile(srcName, os.O_CREATE|os.O_RDWR, 0700)

		dstName := "testFile2"
		dst, err := os.OpenFile(dstName, os.O_CREATE|os.O_RDWR, 0700)
		if err != nil {
			t.Fatal("Failed to create dst file")
		}

		result, err := CopyData(dst, src)
		src.Close()
		dst.Close()

		assert.NoError(t, err, "Didn't expect error but got one")
		assert.Equal(t, int64(300), result, "Expected true but got false")
	})
	runInTempDir(t, "copying file fails on read", func(t *testing.T) {
		srcName := "testFileExcl"
		src, err := os.OpenFile(srcName, os.O_CREATE|os.O_RDWR, 0700)
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
		dst, err := os.OpenFile(dstName, os.O_CREATE|os.O_RDWR, 0700)
		if err != nil {
			t.Fatal("Failed to create dst file")
		}

		result, err := CopyData(dst, src)
		src.Close()
		dst.Close()

		assert.Error(t, err, "Expected error but got none")
		assert.Equal(t, int64(0), result, "Expected true but got false")
	})
	runInTempDir(t, "copying file fails on write", func(t *testing.T) {
		srcName := "testFileExcl"
		src, err := os.OpenFile(srcName, os.O_CREATE|os.O_RDWR, 0700)
		if err != nil {
			t.Fatalf("Failed to create src file %v", err)
		}
		data := make([]byte, 300)
		for i := 0; i < 300; i++ {
			data[i] = byte(i)
		}
		_, err = src.Write(data)
		src.Close()
		src, err = os.OpenFile(srcName, os.O_CREATE|os.O_RDWR, 0700)

		dstName := "testFileExclus"
		dst, err := os.OpenFile(dstName, os.O_CREATE|os.O_RDWR, 0700)
		if err != nil {
			t.Fatalf("Failed to create dst file: %v", err)
		}
		dst.Close()
		dst, err = os.OpenFile(dstName, os.O_RDONLY, 0700)

		result, err := CopyData(dst, src)
		src.Close()
		dst.Close()

		assert.Error(t, err, "Expected error but got none")
		assert.Equal(t, int64(0), result, "Expected true but got false")
	})
}
