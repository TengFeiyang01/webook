// Code generated by MockGen. DO NOT EDIT.
// Source: ./types.go
//
// Generated by this command:
//
//	mockgen -source=./types.go -package=loggermocks -destination=./mocks/logger.mock.go LoggerV1
//

// Package loggermocks is a generated GoMock package.
package loggermocks

import (
	logger "github.com/TengFeiyang01/webook/webook/pkg/logger"
	reflect "reflect"

	gomock "go.uber.org/mock/gomock"
)

// MockLogger is a mock of Logger interface.
type MockLogger struct {
	ctrl     *gomock.Controller
	recorder *MockLoggerMockRecorder
}

// MockLoggerMockRecorder is the mock recorder for MockLogger.
type MockLoggerMockRecorder struct {
	mock *MockLogger
}

// NewMockLogger creates a new mock instance.
func NewMockLogger(ctrl *gomock.Controller) *MockLogger {
	mock := &MockLogger{ctrl: ctrl}
	mock.recorder = &MockLoggerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLogger) EXPECT() *MockLoggerMockRecorder {
	return m.recorder
}

// Debug mocks base method.
func (m *MockLogger) Debug(msg string, args ...any) {
	m.ctrl.T.Helper()
	varargs := []any{msg}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Debug", varargs...)
}

// Debug indicates an expected call of Debug.
func (mr *MockLoggerMockRecorder) Debug(msg any, args ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{msg}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Debug", reflect.TypeOf((*MockLogger)(nil).Debug), varargs...)
}

// Error mocks base method.
func (m *MockLogger) Error(msg string, args ...any) {
	m.ctrl.T.Helper()
	varargs := []any{msg}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Error", varargs...)
}

// Error indicates an expected call of Error.
func (mr *MockLoggerMockRecorder) Error(msg any, args ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{msg}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Error", reflect.TypeOf((*MockLogger)(nil).Error), varargs...)
}

// Info mocks base method.
func (m *MockLogger) Info(msg string, args ...any) {
	m.ctrl.T.Helper()
	varargs := []any{msg}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Info", varargs...)
}

// Info indicates an expected call of Info.
func (mr *MockLoggerMockRecorder) Info(msg any, args ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{msg}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Info", reflect.TypeOf((*MockLogger)(nil).Info), varargs...)
}

// Warn mocks base method.
func (m *MockLogger) Warn(msg string, args ...any) {
	m.ctrl.T.Helper()
	varargs := []any{msg}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Warn", varargs...)
}

// Warn indicates an expected call of Warn.
func (mr *MockLoggerMockRecorder) Warn(msg any, args ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{msg}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Warn", reflect.TypeOf((*MockLogger)(nil).Warn), varargs...)
}

// MockLoggerV1 is a mock of LoggerV1 interface.
type MockLoggerV1 struct {
	ctrl     *gomock.Controller
	recorder *MockLoggerV1MockRecorder
}

// MockLoggerV1MockRecorder is the mock recorder for MockLoggerV1.
type MockLoggerV1MockRecorder struct {
	mock *MockLoggerV1
}

// NewMockLoggerV1 creates a new mock instance.
func NewMockLoggerV1(ctrl *gomock.Controller) *MockLoggerV1 {
	mock := &MockLoggerV1{ctrl: ctrl}
	mock.recorder = &MockLoggerV1MockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLoggerV1) EXPECT() *MockLoggerV1MockRecorder {
	return m.recorder
}

// Debug mocks base method.
func (m *MockLoggerV1) Debug(msg string, args ...logger.Field) {
	m.ctrl.T.Helper()
	varargs := []any{msg}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Debug", varargs...)
}

// Debug indicates an expected call of Debug.
func (mr *MockLoggerV1MockRecorder) Debug(msg any, args ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{msg}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Debug", reflect.TypeOf((*MockLoggerV1)(nil).Debug), varargs...)
}

// Error mocks base method.
func (m *MockLoggerV1) Error(msg string, args ...logger.Field) {
	m.ctrl.T.Helper()
	varargs := []any{msg}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Error", varargs...)
}

// Error indicates an expected call of Error.
func (mr *MockLoggerV1MockRecorder) Error(msg any, args ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{msg}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Error", reflect.TypeOf((*MockLoggerV1)(nil).Error), varargs...)
}

// Info mocks base method.
func (m *MockLoggerV1) Info(msg string, args ...logger.Field) {
	m.ctrl.T.Helper()
	varargs := []any{msg}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Info", varargs...)
}

// Info indicates an expected call of Info.
func (mr *MockLoggerV1MockRecorder) Info(msg any, args ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{msg}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Info", reflect.TypeOf((*MockLoggerV1)(nil).Info), varargs...)
}

// Warn mocks base method.
func (m *MockLoggerV1) Warn(msg string, args ...logger.Field) {
	m.ctrl.T.Helper()
	varargs := []any{msg}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Warn", varargs...)
}

// Warn indicates an expected call of Warn.
func (mr *MockLoggerV1MockRecorder) Warn(msg any, args ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{msg}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Warn", reflect.TypeOf((*MockLoggerV1)(nil).Warn), varargs...)
}

// With mocks base method.
func (m *MockLoggerV1) With(args ...logger.Field) logger.LoggerV1 {
	m.ctrl.T.Helper()
	varargs := []any{}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "With", varargs...)
	ret0, _ := ret[0].(logger.LoggerV1)
	return ret0
}

// With indicates an expected call of With.
func (mr *MockLoggerV1MockRecorder) With(args ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "With", reflect.TypeOf((*MockLoggerV1)(nil).With), args...)
}

// MockLoggerV2 is a mock of LoggerV2 interface.
type MockLoggerV2 struct {
	ctrl     *gomock.Controller
	recorder *MockLoggerV2MockRecorder
}

// MockLoggerV2MockRecorder is the mock recorder for MockLoggerV2.
type MockLoggerV2MockRecorder struct {
	mock *MockLoggerV2
}

// NewMockLoggerV2 creates a new mock instance.
func NewMockLoggerV2(ctrl *gomock.Controller) *MockLoggerV2 {
	mock := &MockLoggerV2{ctrl: ctrl}
	mock.recorder = &MockLoggerV2MockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLoggerV2) EXPECT() *MockLoggerV2MockRecorder {
	return m.recorder
}

// Debug mocks base method.
func (m *MockLoggerV2) Debug(msg string, args ...any) {
	m.ctrl.T.Helper()
	varargs := []any{msg}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Debug", varargs...)
}

// Debug indicates an expected call of Debug.
func (mr *MockLoggerV2MockRecorder) Debug(msg any, args ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{msg}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Debug", reflect.TypeOf((*MockLoggerV2)(nil).Debug), varargs...)
}

// Error mocks base method.
func (m *MockLoggerV2) Error(msg string, args ...any) {
	m.ctrl.T.Helper()
	varargs := []any{msg}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Error", varargs...)
}

// Error indicates an expected call of Error.
func (mr *MockLoggerV2MockRecorder) Error(msg any, args ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{msg}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Error", reflect.TypeOf((*MockLoggerV2)(nil).Error), varargs...)
}

// Info mocks base method.
func (m *MockLoggerV2) Info(msg string, args ...any) {
	m.ctrl.T.Helper()
	varargs := []any{msg}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Info", varargs...)
}

// Info indicates an expected call of Info.
func (mr *MockLoggerV2MockRecorder) Info(msg any, args ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{msg}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Info", reflect.TypeOf((*MockLoggerV2)(nil).Info), varargs...)
}

// Warn mocks base method.
func (m *MockLoggerV2) Warn(msg string, args ...any) {
	m.ctrl.T.Helper()
	varargs := []any{msg}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	m.ctrl.Call(m, "Warn", varargs...)
}

// Warn indicates an expected call of Warn.
func (mr *MockLoggerV2MockRecorder) Warn(msg any, args ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{msg}, args...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Warn", reflect.TypeOf((*MockLoggerV2)(nil).Warn), varargs...)
}

// With mocks base method.
func (m *MockLoggerV2) With(args ...logger.Field) logger.LoggerV2 {
	m.ctrl.T.Helper()
	varargs := []any{}
	for _, a := range args {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "With", varargs...)
	ret0, _ := ret[0].(logger.LoggerV2)
	return ret0
}

// With indicates an expected call of With.
func (mr *MockLoggerV2MockRecorder) With(args ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "With", reflect.TypeOf((*MockLoggerV2)(nil).With), args...)
}
