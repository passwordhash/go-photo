// Code generated by MockGen. DO NOT EDIT.
// Source: service.go

// Package mock_service is a generated GoMock package.
package mock_service

import (
	context "context"
	model "go-photo/internal/model"
	multipart "mime/multipart"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
)

// MockUserService is a mock of UserService interface.
type MockUserService struct {
	ctrl     *gomock.Controller
	recorder *MockUserServiceMockRecorder
}

// MockUserServiceMockRecorder is the mock recorder for MockUserService.
type MockUserServiceMockRecorder struct {
	mock *MockUserService
}

// NewMockUserService creates a new mock instance.
func NewMockUserService(ctrl *gomock.Controller) *MockUserService {
	mock := &MockUserService{ctrl: ctrl}
	mock.recorder = &MockUserServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockUserService) EXPECT() *MockUserServiceMockRecorder {
	return m.recorder
}

// Get mocks base method.
func (m *MockUserService) Get(ctx context.Context, uuid string) (model.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", ctx, uuid)
	ret0, _ := ret[0].(model.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockUserServiceMockRecorder) Get(ctx, uuid interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockUserService)(nil).Get), ctx, uuid)
}

// GetAll mocks base method.
func (m *MockUserService) GetAll(ctx context.Context) ([]model.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAll", ctx)
	ret0, _ := ret[0].([]model.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAll indicates an expected call of GetAll.
func (mr *MockUserServiceMockRecorder) GetAll(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAll", reflect.TypeOf((*MockUserService)(nil).GetAll), ctx)
}

// MockPhotoService is a mock of PhotoService interface.
type MockPhotoService struct {
	ctrl     *gomock.Controller
	recorder *MockPhotoServiceMockRecorder
}

// MockPhotoServiceMockRecorder is the mock recorder for MockPhotoService.
type MockPhotoServiceMockRecorder struct {
	mock *MockPhotoService
}

// NewMockPhotoService creates a new mock instance.
func NewMockPhotoService(ctrl *gomock.Controller) *MockPhotoService {
	mock := &MockPhotoService{ctrl: ctrl}
	mock.recorder = &MockPhotoServiceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPhotoService) EXPECT() *MockPhotoServiceMockRecorder {
	return m.recorder
}

// GetPhotoVersions mocks base method.
func (m *MockPhotoService) GetPhotoVersions(ctx context.Context, photoID int) ([]model.PhotoVersion, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPhotoVersions", ctx, photoID)
	ret0, _ := ret[0].([]model.PhotoVersion)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPhotoVersions indicates an expected call of GetPhotoVersions.
func (mr *MockPhotoServiceMockRecorder) GetPhotoVersions(ctx, photoID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPhotoVersions", reflect.TypeOf((*MockPhotoService)(nil).GetPhotoVersions), ctx, photoID)
}

// UploadPhoto mocks base method.
func (m *MockPhotoService) UploadPhoto(ctx context.Context, userUUID string, file multipart.File, photoName string) (int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UploadPhoto", ctx, userUUID, file, photoName)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// UploadPhoto indicates an expected call of UploadPhoto.
func (mr *MockPhotoServiceMockRecorder) UploadPhoto(ctx, userUUID, file, photoName interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UploadPhoto", reflect.TypeOf((*MockPhotoService)(nil).UploadPhoto), ctx, userUUID, file, photoName)
}
