// Code generated by mockery v2.7.5. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	time "time"
)

// Task is an autogenerated mock type for the Task type
type Task struct {
	mock.Mock
}

// BuildNumber provides a mock function with given fields:
func (_m *Task) BuildNumber() (int64, error) {
	ret := _m.Called()

	var r0 int64
	if rf, ok := ret.Get(0).(func() int64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// HasStarted provides a mock function with given fields:
func (_m *Task) HasStarted() bool {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// Poll provides a mock function with given fields: _a0
func (_m *Task) Poll(_a0 context.Context) (int, error) {
	ret := _m.Called(_a0)

	var r0 int
	if rf, ok := ret.Get(0).(func(context.Context) int); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(int)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// WaitToStart provides a mock function with given fields: ctx, pollInterval
func (_m *Task) WaitToStart(ctx context.Context, pollInterval time.Duration) (int64, error) {
	ret := _m.Called(ctx, pollInterval)

	var r0 int64
	if rf, ok := ret.Get(0).(func(context.Context, time.Duration) int64); ok {
		r0 = rf(ctx, pollInterval)
	} else {
		r0 = ret.Get(0).(int64)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, time.Duration) error); ok {
		r1 = rf(ctx, pollInterval)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
