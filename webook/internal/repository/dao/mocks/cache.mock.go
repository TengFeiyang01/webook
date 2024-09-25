// Code generated by MockGen. DO NOT EDIT.
// Source: webook/internal/repository/dao/types.go
//
// Generated by this command:
//
//	mockgen -source=webook/internal/repository/dao/types.go -package=daomocks -destination=webook/internal/repository/dao/mocks/cache.mock.go
//

// Package daomocks is a generated GoMock package.
package daomocks

import (
	context "context"
	reflect "reflect"
	dao "webook/webook/internal/repository/dao"

	gomock "go.uber.org/mock/gomock"
)

// MockUserDAO is a mock of UserDAO interface.
type MockUserDAO struct {
	ctrl     *gomock.Controller
	recorder *MockUserDAOMockRecorder
}

// MockUserDAOMockRecorder is the mock recorder for MockUserDAO.
type MockUserDAOMockRecorder struct {
	mock *MockUserDAO
}

// NewMockUserDAO creates a new mock instance.
func NewMockUserDAO(ctrl *gomock.Controller) *MockUserDAO {
	mock := &MockUserDAO{ctrl: ctrl}
	mock.recorder = &MockUserDAOMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserDAO) EXPECT() *MockUserDAOMockRecorder {
	return m.recorder
}

// FindByEmail mocks base method.
func (m *MockUserDAO) FindByEmail(ctx context.Context, email string) (dao.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindByEmail", ctx, email)
	ret0, _ := ret[0].(dao.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindByEmail indicates an expected call of FindByEmail.
func (mr *MockUserDAOMockRecorder) FindByEmail(ctx, email any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindByEmail", reflect.TypeOf((*MockUserDAO)(nil).FindByEmail), ctx, email)
}

// FindById mocks base method.
func (m *MockUserDAO) FindById(ctx context.Context, id int64) (dao.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindById", ctx, id)
	ret0, _ := ret[0].(dao.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindById indicates an expected call of FindById.
func (mr *MockUserDAOMockRecorder) FindById(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindById", reflect.TypeOf((*MockUserDAO)(nil).FindById), ctx, id)
}

// FindByPhone mocks base method.
func (m *MockUserDAO) FindByPhone(ctx context.Context, phone string) (dao.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "FindByPhone", ctx, phone)
	ret0, _ := ret[0].(dao.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// FindByPhone indicates an expected call of FindByPhone.
func (mr *MockUserDAOMockRecorder) FindByPhone(ctx, phone any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "FindByPhone", reflect.TypeOf((*MockUserDAO)(nil).FindByPhone), ctx, phone)
}

// Insert mocks base method.
func (m *MockUserDAO) Insert(ctx context.Context, u dao.User) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Insert", ctx, u)
	ret0, _ := ret[0].(error)
	return ret0
}

// Insert indicates an expected call of Insert.
func (mr *MockUserDAOMockRecorder) Insert(ctx, u any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Insert", reflect.TypeOf((*MockUserDAO)(nil).Insert), ctx, u)
}

// UpdateById mocks base method.
func (m *MockUserDAO) UpdateById(ctx context.Context, entity dao.User) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateById", ctx, entity)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateById indicates an expected call of UpdateById.
func (mr *MockUserDAOMockRecorder) UpdateById(ctx, entity any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateById", reflect.TypeOf((*MockUserDAO)(nil).UpdateById), ctx, entity)
}
