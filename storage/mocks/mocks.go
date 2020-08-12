// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/whitekid/go-todo/storage (interfaces: Interface)

// Package mock_storage is a generated GoMock package.
package mock_storage

import (
	gomock "github.com/golang/mock/gomock"
	types "github.com/whitekid/go-todo/storage/types"
	reflect "reflect"
)

// MockInterface is a mock of Interface interface
type MockInterface struct {
	ctrl     *gomock.Controller
	recorder *MockInterfaceMockRecorder
}

// MockInterfaceMockRecorder is the mock recorder for MockInterface
type MockInterfaceMockRecorder struct {
	mock *MockInterface
}

// NewMockInterface creates a new mock instance
func NewMockInterface(ctrl *gomock.Controller) *MockInterface {
	mock := &MockInterface{ctrl: ctrl}
	mock.recorder = &MockInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockInterface) EXPECT() *MockInterfaceMockRecorder {
	return m.recorder
}

// Close mocks base method
func (m *MockInterface) Close() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Close")
}

// Close indicates an expected call of Close
func (mr *MockInterfaceMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockInterface)(nil).Close))
}

// TodoService mocks base method
func (m *MockInterface) TodoService() types.TodoService {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TodoService")
	ret0, _ := ret[0].(types.TodoService)
	return ret0
}

// TodoService indicates an expected call of TodoService
func (mr *MockInterfaceMockRecorder) TodoService() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TodoService", reflect.TypeOf((*MockInterface)(nil).TodoService))
}

// TokenService mocks base method
func (m *MockInterface) TokenService() types.TokenService {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TokenService")
	ret0, _ := ret[0].(types.TokenService)
	return ret0
}

// TokenService indicates an expected call of TokenService
func (mr *MockInterfaceMockRecorder) TokenService() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TokenService", reflect.TypeOf((*MockInterface)(nil).TokenService))
}

// UserService mocks base method
func (m *MockInterface) UserService() types.UserService {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UserService")
	ret0, _ := ret[0].(types.UserService)
	return ret0
}

// UserService indicates an expected call of UserService
func (mr *MockInterfaceMockRecorder) UserService() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UserService", reflect.TypeOf((*MockInterface)(nil).UserService))
}