// Code generated by mockery v2.50.0. DO NOT EDIT.

//go:build !compile

package ws

import mock "github.com/stretchr/testify/mock"

// Mockreader is an autogenerated mock type for the reader type
type Mockreader struct {
	mock.Mock
}

type Mockreader_Expecter struct {
	mock *mock.Mock
}

func (_m *Mockreader) EXPECT() *Mockreader_Expecter {
	return &Mockreader_Expecter{mock: &_m.Mock}
}

// Read provides a mock function with given fields: p
func (_m *Mockreader) Read(p []byte) (int, error) {
	ret := _m.Called(p)

	if len(ret) == 0 {
		panic("no return value specified for Read")
	}

	var r0 int
	var r1 error
	if rf, ok := ret.Get(0).(func([]byte) (int, error)); ok {
		return rf(p)
	}
	if rf, ok := ret.Get(0).(func([]byte) int); ok {
		r0 = rf(p)
	} else {
		r0 = ret.Get(0).(int)
	}

	if rf, ok := ret.Get(1).(func([]byte) error); ok {
		r1 = rf(p)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Mockreader_Read_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Read'
type Mockreader_Read_Call struct {
	*mock.Call
}

// Read is a helper method to define mock.On call
//   - p []byte
func (_e *Mockreader_Expecter) Read(p interface{}) *Mockreader_Read_Call {
	return &Mockreader_Read_Call{Call: _e.mock.On("Read", p)}
}

func (_c *Mockreader_Read_Call) Run(run func(p []byte)) *Mockreader_Read_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].([]byte))
	})
	return _c
}

func (_c *Mockreader_Read_Call) Return(n int, err error) *Mockreader_Read_Call {
	_c.Call.Return(n, err)
	return _c
}

func (_c *Mockreader_Read_Call) RunAndReturn(run func([]byte) (int, error)) *Mockreader_Read_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockreader creates a new instance of Mockreader. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockreader(t interface {
	mock.TestingT
	Cleanup(func())
}) *Mockreader {
	mock := &Mockreader{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
