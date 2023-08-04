package main

import (
	"io/fs"
	"time"

	"github.com/stretchr/testify/mock"
)

type fileHandlerMock struct {
	mock.Mock
}

func NewHandlerMock(builders ...func(builder *fileHandlerMock)) *fileHandlerMock {
	m := &fileHandlerMock{}

	for i := range builders {
		builders[i](m)
	}

	return m
}

func (h *fileHandlerMock) Readdir(n int) ([]fs.FileInfo, error) {
	args := h.Called(n)

	return args.Get(0).([]fs.FileInfo), args.Error(1)
}

func (h *fileHandlerMock) Close() error {
	args := h.Called()

	return args.Error(0)
}

type fileInfoMock struct {
	mock.Mock
}

func NewFileInfoMock(builders ...func(builder *fileInfoMock)) *fileInfoMock {
	m := &fileInfoMock{}

	for i := range builders {
		builders[i](m)
	}

	return m
}

func (h *fileInfoMock) Name() string {
	args := h.Called()

	return args.Get(0).(string)
}

func (h *fileInfoMock) Size() int64 {
	args := h.Called()

	return args.Get(0).(int64)
}

func (h *fileInfoMock) Mode() fs.FileMode {
	args := h.Called()

	return args.Get(0).(fs.FileMode)
}

func (h *fileInfoMock) ModTime() time.Time {
	args := h.Called()

	return args.Get(0).(time.Time)
}

func (h *fileInfoMock) IsDir() bool {
	args := h.Called()

	return args.Get(0).(bool)
}

func (h *fileInfoMock) Sys() any {
	args := h.Called()

	return args.Get(0).(any)
}
