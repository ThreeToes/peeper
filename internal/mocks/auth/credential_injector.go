// Code generated by MockGen. DO NOT EDIT.
// Source: ./credential_injector.go

// Package mock_auth is a generated GoMock package.
package mock_auth

import (
	http "net/http"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockCredentialInjector is a mock of CredentialInjector interface.
type MockCredentialInjector struct {
	ctrl     *gomock.Controller
	recorder *MockCredentialInjectorMockRecorder
}

// MockCredentialInjectorMockRecorder is the mock recorder for MockCredentialInjector.
type MockCredentialInjectorMockRecorder struct {
	mock *MockCredentialInjector
}

// NewMockCredentialInjector creates a new mock instance.
func NewMockCredentialInjector(ctrl *gomock.Controller) *MockCredentialInjector {
	mock := &MockCredentialInjector{ctrl: ctrl}
	mock.recorder = &MockCredentialInjectorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCredentialInjector) EXPECT() *MockCredentialInjectorMockRecorder {
	return m.recorder
}

// InjectCredentials mocks base method.
func (m *MockCredentialInjector) InjectCredentials(req *http.Request) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "InjectCredentials", req)
	ret0, _ := ret[0].(error)
	return ret0
}

// InjectCredentials indicates an expected call of InjectCredentials.
func (mr *MockCredentialInjectorMockRecorder) InjectCredentials(req interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "InjectCredentials", reflect.TypeOf((*MockCredentialInjector)(nil).InjectCredentials), req)
}