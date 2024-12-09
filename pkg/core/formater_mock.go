// Code generated by mockery v2.46.3. DO NOT EDIT.

//go:build !compile

package core

import mock "github.com/stretchr/testify/mock"

// MockFormater is an autogenerated mock type for the Formater type
type MockFormater struct {
	mock.Mock
}

type MockFormater_Expecter struct {
	mock *mock.Mock
}

func (_m *MockFormater) EXPECT() *MockFormater_Expecter {
	return &MockFormater_Expecter{mock: &_m.Mock}
}

// FormatForFile provides a mock function with given fields: msgType, msgData
func (_m *MockFormater) FormatForFile(msgType string, msgData string) (string, error) {
	ret := _m.Called(msgType, msgData)

	if len(ret) == 0 {
		panic("no return value specified for FormatForFile")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string) (string, error)); ok {
		return rf(msgType, msgData)
	}
	if rf, ok := ret.Get(0).(func(string, string) string); ok {
		r0 = rf(msgType, msgData)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(msgType, msgData)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockFormater_FormatForFile_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FormatForFile'
type MockFormater_FormatForFile_Call struct {
	*mock.Call
}

// FormatForFile is a helper method to define mock.On call
//   - msgType string
//   - msgData string
func (_e *MockFormater_Expecter) FormatForFile(msgType interface{}, msgData interface{}) *MockFormater_FormatForFile_Call {
	return &MockFormater_FormatForFile_Call{Call: _e.mock.On("FormatForFile", msgType, msgData)}
}

func (_c *MockFormater_FormatForFile_Call) Run(run func(msgType string, msgData string)) *MockFormater_FormatForFile_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *MockFormater_FormatForFile_Call) Return(_a0 string, _a1 error) *MockFormater_FormatForFile_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockFormater_FormatForFile_Call) RunAndReturn(run func(string, string) (string, error)) *MockFormater_FormatForFile_Call {
	_c.Call.Return(run)
	return _c
}

// FormatMessage provides a mock function with given fields: msgType, msgData
func (_m *MockFormater) FormatMessage(msgType string, msgData string) (string, error) {
	ret := _m.Called(msgType, msgData)

	if len(ret) == 0 {
		panic("no return value specified for FormatMessage")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(string, string) (string, error)); ok {
		return rf(msgType, msgData)
	}
	if rf, ok := ret.Get(0).(func(string, string) string); ok {
		r0 = rf(msgType, msgData)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(msgType, msgData)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MockFormater_FormatMessage_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'FormatMessage'
type MockFormater_FormatMessage_Call struct {
	*mock.Call
}

// FormatMessage is a helper method to define mock.On call
//   - msgType string
//   - msgData string
func (_e *MockFormater_Expecter) FormatMessage(msgType interface{}, msgData interface{}) *MockFormater_FormatMessage_Call {
	return &MockFormater_FormatMessage_Call{Call: _e.mock.On("FormatMessage", msgType, msgData)}
}

func (_c *MockFormater_FormatMessage_Call) Run(run func(msgType string, msgData string)) *MockFormater_FormatMessage_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *MockFormater_FormatMessage_Call) Return(_a0 string, _a1 error) *MockFormater_FormatMessage_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *MockFormater_FormatMessage_Call) RunAndReturn(run func(string, string) (string, error)) *MockFormater_FormatMessage_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockFormater creates a new instance of MockFormater. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockFormater(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockFormater {
	mock := &MockFormater{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
