// Code generated by MockGen. DO NOT EDIT.
// Source: ./art.go
//
// Generated by this command:
//
//	mockgen -source=./art.go -destination=./mocks/service.mock.go -package=svcmocks ArticleService
//

// Package svcmocks is a generated GoMock package.
package svcmocks

import (
	context "context"
	reflect "reflect"
	time "time"
	domain "webook/webook/article/domain"

	gomock "go.uber.org/mock/gomock"
)

// MockArticleService is a mock of ArticleService interface.
type MockArticleService struct {
	ctrl     *gomock.Controller
	recorder *MockArticleServiceMockRecorder
}

// MockArticleServiceMockRecorder is the mock recorder for MockArticleService.
type MockArticleServiceMockRecorder struct {
	mock *MockArticleService
}

// NewMockArticleService creates a new mock instance.
func NewMockArticleService(ctrl *gomock.Controller) *MockArticleService {
	mock := &MockArticleService{ctrl: ctrl}
	mock.recorder = &MockArticleServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockArticleService) EXPECT() *MockArticleServiceMockRecorder {
	return m.recorder
}

// GetById mocks base method.
func (m *MockArticleService) GetById(ctx context.Context, id int64) (domain.Article, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetById", ctx, id)
	ret0, _ := ret[0].(domain.Article)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetById indicates an expected call of GetById.
func (mr *MockArticleServiceMockRecorder) GetById(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetById", reflect.TypeOf((*MockArticleService)(nil).GetById), ctx, id)
}

// GetPublishedById mocks base method.
func (m *MockArticleService) GetPublishedById(ctx context.Context, id, uid int64) (domain.Article, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPublishedById", ctx, id, uid)
	ret0, _ := ret[0].(domain.Article)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPublishedById indicates an expected call of GetPublishedById.
func (mr *MockArticleServiceMockRecorder) GetPublishedById(ctx, id, uid any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPublishedById", reflect.TypeOf((*MockArticleService)(nil).GetPublishedById), ctx, id, uid)
}

// List mocks base method.
func (m *MockArticleService) List(ctx context.Context, id int64, offset, limit int) ([]domain.Article, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "List", ctx, id, offset, limit)
	ret0, _ := ret[0].([]domain.Article)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// List indicates an expected call of List.
func (mr *MockArticleServiceMockRecorder) List(ctx, id, offset, limit any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "List", reflect.TypeOf((*MockArticleService)(nil).List), ctx, id, offset, limit)
}

// ListPub mocks base method.
func (m *MockArticleService) ListPub(ctx context.Context, start time.Time, offset, limit int) ([]domain.Article, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListPub", ctx, start, offset, limit)
	ret0, _ := ret[0].([]domain.Article)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListPub indicates an expected call of ListPub.
func (mr *MockArticleServiceMockRecorder) ListPub(ctx, start, offset, limit any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListPub", reflect.TypeOf((*MockArticleService)(nil).ListPub), ctx, start, offset, limit)
}

// Publish mocks base method.
func (m *MockArticleService) Publish(ctx context.Context, art domain.Article) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Publish", ctx, art)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Publish indicates an expected call of Publish.
func (mr *MockArticleServiceMockRecorder) Publish(ctx, art any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Publish", reflect.TypeOf((*MockArticleService)(nil).Publish), ctx, art)
}

// PublishV1 mocks base method.
func (m *MockArticleService) PublishV1(ctx context.Context, art domain.Article) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PublishV1", ctx, art)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// PublishV1 indicates an expected call of PublishV1.
func (mr *MockArticleServiceMockRecorder) PublishV1(ctx, art any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PublishV1", reflect.TypeOf((*MockArticleService)(nil).PublishV1), ctx, art)
}

// Save mocks base method.
func (m *MockArticleService) Save(ctx context.Context, art domain.Article) (int64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Save", ctx, art)
	ret0, _ := ret[0].(int64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Save indicates an expected call of Save.
func (mr *MockArticleServiceMockRecorder) Save(ctx, art any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Save", reflect.TypeOf((*MockArticleService)(nil).Save), ctx, art)
}

// WithDraw mocks base method.
func (m *MockArticleService) WithDraw(ctx context.Context, art domain.Article) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithDraw", ctx, art)
	ret0, _ := ret[0].(error)
	return ret0
}

// WithDraw indicates an expected call of WithDraw.
func (mr *MockArticleServiceMockRecorder) WithDraw(ctx, art any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithDraw", reflect.TypeOf((*MockArticleService)(nil).WithDraw), ctx, art)
}
