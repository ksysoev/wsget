// Code generated by mockery v2.46.3. DO NOT EDIT.

//go:build !compile

package core

import (
	io "io"

	formater "github.com/ksysoev/wsget/pkg/formater"

	mock "github.com/stretchr/testify/mock"
)

// MockExecutionContext is an autogenerated mock type for the ExecutionContext type
type MockExecutionContext struct {
	mock.Mock
}

type MockExecutionContext_Expecter struct {
	mock *mock.Mock
}

func (_m *MockExecutionContext) EXPECT() *MockExecutionContext_Expecter {
	return &MockExecutionContext_Expecter{mock: &_m.Mock}
}

// Connection provides a mock function with given fields:
func (_m *MockExecutionContext) Connection() ConnectionHandler {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Connection")
	}

	var r0 ConnectionHandler
	if rf, ok := ret.Get(0).(func() ConnectionHandler); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(ConnectionHandler)
		}
	}

	return r0
}

// MockExecutionContext_Connection_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Connection'
type MockExecutionContext_Connection_Call struct {
	*mock.Call
}

// Connection is a helper method to define mock.On call
func (_e *MockExecutionContext_Expecter) Connection() *MockExecutionContext_Connection_Call {
	return &MockExecutionContext_Connection_Call{Call: _e.mock.On("Connection")}
}

func (_c *MockExecutionContext_Connection_Call) Run(run func()) *MockExecutionContext_Connection_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockExecutionContext_Connection_Call) Return(_a0 ConnectionHandler) *MockExecutionContext_Connection_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockExecutionContext_Connection_Call) RunAndReturn(run func() ConnectionHandler) *MockExecutionContext_Connection_Call {
	_c.Call.Return(run)
	return _c
}

// Editor provides a mock function with given fields:
func (_m *MockExecutionContext) Editor() Editor {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Editor")
	}

	var r0 Editor
	if rf, ok := ret.Get(0).(func() Editor); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(Editor)
		}
	}

	return r0
}

// MockExecutionContext_Editor_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Editor'
type MockExecutionContext_Editor_Call struct {
	*mock.Call
}

// Editor is a helper method to define mock.On call
func (_e *MockExecutionContext_Expecter) Editor() *MockExecutionContext_Editor_Call {
	return &MockExecutionContext_Editor_Call{Call: _e.mock.On("Editor")}
}

func (_c *MockExecutionContext_Editor_Call) Run(run func()) *MockExecutionContext_Editor_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockExecutionContext_Editor_Call) Return(_a0 Editor) *MockExecutionContext_Editor_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockExecutionContext_Editor_Call) RunAndReturn(run func() Editor) *MockExecutionContext_Editor_Call {
	_c.Call.Return(run)
	return _c
}

// Factory provides a mock function with given fields:
func (_m *MockExecutionContext) Factory() CommandFactory {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Factory")
	}

	var r0 CommandFactory
	if rf, ok := ret.Get(0).(func() CommandFactory); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(CommandFactory)
		}
	}

	return r0
}

// MockExecutionContext_Factory_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Factory'
type MockExecutionContext_Factory_Call struct {
	*mock.Call
}

// Factory is a helper method to define mock.On call
func (_e *MockExecutionContext_Expecter) Factory() *MockExecutionContext_Factory_Call {
	return &MockExecutionContext_Factory_Call{Call: _e.mock.On("Factory")}
}

func (_c *MockExecutionContext_Factory_Call) Run(run func()) *MockExecutionContext_Factory_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockExecutionContext_Factory_Call) Return(_a0 CommandFactory) *MockExecutionContext_Factory_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockExecutionContext_Factory_Call) RunAndReturn(run func() CommandFactory) *MockExecutionContext_Factory_Call {
	_c.Call.Return(run)
	return _c
}

// Formater provides a mock function with given fields:
func (_m *MockExecutionContext) Formater() formater.Formater {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Formater")
	}

	var r0 formater.Formater
	if rf, ok := ret.Get(0).(func() formater.Formater); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(formater.Formater)
		}
	}

	return r0
}

// MockExecutionContext_Formater_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Formater'
type MockExecutionContext_Formater_Call struct {
	*mock.Call
}

// Formater is a helper method to define mock.On call
func (_e *MockExecutionContext_Expecter) Formater() *MockExecutionContext_Formater_Call {
	return &MockExecutionContext_Formater_Call{Call: _e.mock.On("Formater")}
}

func (_c *MockExecutionContext_Formater_Call) Run(run func()) *MockExecutionContext_Formater_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockExecutionContext_Formater_Call) Return(_a0 formater.Formater) *MockExecutionContext_Formater_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockExecutionContext_Formater_Call) RunAndReturn(run func() formater.Formater) *MockExecutionContext_Formater_Call {
	_c.Call.Return(run)
	return _c
}

// Input provides a mock function with given fields:
func (_m *MockExecutionContext) Input() <-chan KeyEvent {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Input")
	}

	var r0 <-chan KeyEvent
	if rf, ok := ret.Get(0).(func() <-chan KeyEvent); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(<-chan KeyEvent)
		}
	}

	return r0
}

// MockExecutionContext_Input_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Input'
type MockExecutionContext_Input_Call struct {
	*mock.Call
}

// Input is a helper method to define mock.On call
func (_e *MockExecutionContext_Expecter) Input() *MockExecutionContext_Input_Call {
	return &MockExecutionContext_Input_Call{Call: _e.mock.On("Input")}
}

func (_c *MockExecutionContext_Input_Call) Run(run func()) *MockExecutionContext_Input_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockExecutionContext_Input_Call) Return(_a0 <-chan KeyEvent) *MockExecutionContext_Input_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockExecutionContext_Input_Call) RunAndReturn(run func() <-chan KeyEvent) *MockExecutionContext_Input_Call {
	_c.Call.Return(run)
	return _c
}

// Output provides a mock function with given fields:
func (_m *MockExecutionContext) Output() io.Writer {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for Output")
	}

	var r0 io.Writer
	if rf, ok := ret.Get(0).(func() io.Writer); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.Writer)
		}
	}

	return r0
}

// MockExecutionContext_Output_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Output'
type MockExecutionContext_Output_Call struct {
	*mock.Call
}

// Output is a helper method to define mock.On call
func (_e *MockExecutionContext_Expecter) Output() *MockExecutionContext_Output_Call {
	return &MockExecutionContext_Output_Call{Call: _e.mock.On("Output")}
}

func (_c *MockExecutionContext_Output_Call) Run(run func()) *MockExecutionContext_Output_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockExecutionContext_Output_Call) Return(_a0 io.Writer) *MockExecutionContext_Output_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockExecutionContext_Output_Call) RunAndReturn(run func() io.Writer) *MockExecutionContext_Output_Call {
	_c.Call.Return(run)
	return _c
}

// OutputFile provides a mock function with given fields:
func (_m *MockExecutionContext) OutputFile() io.Writer {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for OutputFile")
	}

	var r0 io.Writer
	if rf, ok := ret.Get(0).(func() io.Writer); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(io.Writer)
		}
	}

	return r0
}

// MockExecutionContext_OutputFile_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'OutputFile'
type MockExecutionContext_OutputFile_Call struct {
	*mock.Call
}

// OutputFile is a helper method to define mock.On call
func (_e *MockExecutionContext_Expecter) OutputFile() *MockExecutionContext_OutputFile_Call {
	return &MockExecutionContext_OutputFile_Call{Call: _e.mock.On("OutputFile")}
}

func (_c *MockExecutionContext_OutputFile_Call) Run(run func()) *MockExecutionContext_OutputFile_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockExecutionContext_OutputFile_Call) Return(_a0 io.Writer) *MockExecutionContext_OutputFile_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockExecutionContext_OutputFile_Call) RunAndReturn(run func() io.Writer) *MockExecutionContext_OutputFile_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockExecutionContext creates a new instance of MockExecutionContext. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockExecutionContext(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockExecutionContext {
	mock := &MockExecutionContext{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
