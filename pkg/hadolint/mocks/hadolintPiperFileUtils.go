// Code generated by mockery v2.0.0-alpha.13. DO NOT EDIT.

package mocks

import (
	os "os"

	mock "github.com/stretchr/testify/mock"
)

// HadolintPiperFileUtils is an autogenerated mock type for the HadolintPiperFileUtils type
type HadolintPiperFileUtils struct {
	mock.Mock
}

// FileExists provides a mock function with given fields: filename
func (_m *HadolintPiperFileUtils) FileExists(filename string) (bool, error) {
	ret := _m.Called(filename)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(filename)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(filename)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FileWrite provides a mock function with given fields: filename, data, perm
func (_m *HadolintPiperFileUtils) FileWrite(filename string, data []byte, perm os.FileMode) error {
	ret := _m.Called(filename, data, perm)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, []byte, os.FileMode) error); ok {
		r0 = rf(filename, data, perm)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
