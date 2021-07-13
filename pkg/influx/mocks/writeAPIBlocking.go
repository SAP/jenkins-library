// Code generated by mockery v2.7.4. DO NOT EDIT.

package mocks

import (
	context "context"

	write "github.com/influxdata/influxdb-client-go/v2/api/write"
	mock "github.com/stretchr/testify/mock"
)

// WriteAPIBlocking is an autogenerated mock type for the WriteAPIBlocking type
type WriteAPIBlocking struct {
	mock.Mock
}

// WritePoint provides a mock function with given fields: ctx, point
func (_m *WriteAPIBlocking) WritePoint(ctx context.Context, point ...*write.Point) error {
	_va := make([]interface{}, len(point))
	for _i := range point {
		_va[_i] = point[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, ...*write.Point) error); ok {
		r0 = rf(ctx, point...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// WriteRecord provides a mock function with given fields: ctx, line
func (_m *WriteAPIBlocking) WriteRecord(ctx context.Context, line ...string) error {
	_va := make([]interface{}, len(line))
	for _i := range line {
		_va[_i] = line[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, ...string) error); ok {
		r0 = rf(ctx, line...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
