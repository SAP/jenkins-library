// Code generated by mockery v2.7.4. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Client is an autogenerated mock type for the Client type
type Client struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *Client) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DownloadFile provides a mock function with given fields: bucketID, sourcePath, targetPath
func (_m *Client) DownloadFile(bucketID string, sourcePath string, targetPath string) error {
	ret := _m.Called(bucketID, sourcePath, targetPath)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string) error); ok {
		r0 = rf(bucketID, sourcePath, targetPath)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ListFiles provides a mock function with given fields: bucketID
func (_m *Client) ListFiles(bucketID string) ([]string, error) {
	ret := _m.Called(bucketID)

	var r0 []string
	if rf, ok := ret.Get(0).(func(string) []string); ok {
		r0 = rf(bucketID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(bucketID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UploadFile provides a mock function with given fields: bucketID, sourcePath, targetPath
func (_m *Client) UploadFile(bucketID string, sourcePath string, targetPath string) error {
	ret := _m.Called(bucketID, sourcePath, targetPath)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string) error); ok {
		r0 = rf(bucketID, sourcePath, targetPath)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
