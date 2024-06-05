// Code generated by mockery v2.43.0. DO NOT EDIT.

package mocks

import (
	context "context"

	gojenkins "github.com/bndr/gojenkins"

	mock "github.com/stretchr/testify/mock"
)

// Jenkins is an autogenerated mock type for the Jenkins type
type Jenkins struct {
	mock.Mock
}

type Jenkins_Expecter struct {
	mock *mock.Mock
}

func (_m *Jenkins) EXPECT() *Jenkins_Expecter {
	return &Jenkins_Expecter{mock: &_m.Mock}
}

// BuildJob provides a mock function with given fields: ctx, name, params
func (_m *Jenkins) BuildJob(ctx context.Context, name string, params map[string]string) (int64, error) {
	ret := _m.Called(ctx, name, params)

	if len(ret) == 0 {
		panic("no return value specified for BuildJob")
	}

	var r0 int64
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, map[string]string) (int64, error)); ok {
		return rf(ctx, name, params)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, map[string]string) int64); ok {
		r0 = rf(ctx, name, params)
	} else {
		r0 = ret.Get(0).(int64)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, map[string]string) error); ok {
		r1 = rf(ctx, name, params)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Jenkins_BuildJob_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'BuildJob'
type Jenkins_BuildJob_Call struct {
	*mock.Call
}

// BuildJob is a helper method to define mock.On call
//   - ctx context.Context
//   - name string
//   - params map[string]string
func (_e *Jenkins_Expecter) BuildJob(ctx interface{}, name interface{}, params interface{}) *Jenkins_BuildJob_Call {
	return &Jenkins_BuildJob_Call{Call: _e.mock.On("BuildJob", ctx, name, params)}
}

func (_c *Jenkins_BuildJob_Call) Run(run func(ctx context.Context, name string, params map[string]string)) *Jenkins_BuildJob_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string), args[2].(map[string]string))
	})
	return _c
}

func (_c *Jenkins_BuildJob_Call) Return(_a0 int64, _a1 error) *Jenkins_BuildJob_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Jenkins_BuildJob_Call) RunAndReturn(run func(context.Context, string, map[string]string) (int64, error)) *Jenkins_BuildJob_Call {
	_c.Call.Return(run)
	return _c
}

// GetBuildFromQueueID provides a mock function with given fields: ctx, job, queueid
func (_m *Jenkins) GetBuildFromQueueID(ctx context.Context, job *gojenkins.Job, queueid int64) (*gojenkins.Build, error) {
	ret := _m.Called(ctx, job, queueid)

	if len(ret) == 0 {
		panic("no return value specified for GetBuildFromQueueID")
	}

	var r0 *gojenkins.Build
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *gojenkins.Job, int64) (*gojenkins.Build, error)); ok {
		return rf(ctx, job, queueid)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *gojenkins.Job, int64) *gojenkins.Build); ok {
		r0 = rf(ctx, job, queueid)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gojenkins.Build)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *gojenkins.Job, int64) error); ok {
		r1 = rf(ctx, job, queueid)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Jenkins_GetBuildFromQueueID_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetBuildFromQueueID'
type Jenkins_GetBuildFromQueueID_Call struct {
	*mock.Call
}

// GetBuildFromQueueID is a helper method to define mock.On call
//   - ctx context.Context
//   - job *gojenkins.Job
//   - queueid int64
func (_e *Jenkins_Expecter) GetBuildFromQueueID(ctx interface{}, job interface{}, queueid interface{}) *Jenkins_GetBuildFromQueueID_Call {
	return &Jenkins_GetBuildFromQueueID_Call{Call: _e.mock.On("GetBuildFromQueueID", ctx, job, queueid)}
}

func (_c *Jenkins_GetBuildFromQueueID_Call) Run(run func(ctx context.Context, job *gojenkins.Job, queueid int64)) *Jenkins_GetBuildFromQueueID_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(*gojenkins.Job), args[2].(int64))
	})
	return _c
}

func (_c *Jenkins_GetBuildFromQueueID_Call) Return(_a0 *gojenkins.Build, _a1 error) *Jenkins_GetBuildFromQueueID_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *Jenkins_GetBuildFromQueueID_Call) RunAndReturn(run func(context.Context, *gojenkins.Job, int64) (*gojenkins.Build, error)) *Jenkins_GetBuildFromQueueID_Call {
	_c.Call.Return(run)
	return _c
}

// GetJobObj provides a mock function with given fields: ctx, name
func (_m *Jenkins) GetJobObj(ctx context.Context, name string) *gojenkins.Job {
	ret := _m.Called(ctx, name)

	if len(ret) == 0 {
		panic("no return value specified for GetJobObj")
	}

	var r0 *gojenkins.Job
	if rf, ok := ret.Get(0).(func(context.Context, string) *gojenkins.Job); ok {
		r0 = rf(ctx, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*gojenkins.Job)
		}
	}

	return r0
}

// Jenkins_GetJobObj_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetJobObj'
type Jenkins_GetJobObj_Call struct {
	*mock.Call
}

// GetJobObj is a helper method to define mock.On call
//   - ctx context.Context
//   - name string
func (_e *Jenkins_Expecter) GetJobObj(ctx interface{}, name interface{}) *Jenkins_GetJobObj_Call {
	return &Jenkins_GetJobObj_Call{Call: _e.mock.On("GetJobObj", ctx, name)}
}

func (_c *Jenkins_GetJobObj_Call) Run(run func(ctx context.Context, name string)) *Jenkins_GetJobObj_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(context.Context), args[1].(string))
	})
	return _c
}

func (_c *Jenkins_GetJobObj_Call) Return(_a0 *gojenkins.Job) *Jenkins_GetJobObj_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *Jenkins_GetJobObj_Call) RunAndReturn(run func(context.Context, string) *gojenkins.Job) *Jenkins_GetJobObj_Call {
	_c.Call.Return(run)
	return _c
}

// NewJenkins creates a new instance of Jenkins. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewJenkins(t interface {
	mock.TestingT
	Cleanup(func())
}) *Jenkins {
	mock := &Jenkins{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
