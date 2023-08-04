package main

import (
	"fmt"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_processor_Read(t *testing.T) {
	type fields struct {
		hidden           *bool
		deleteDuplicates *bool
		silent           *bool
		fileInfoBuilder  func(builder *fileInfoMock)
		openerBuilder    func([]fs.FileInfo) func(builder *fileHandlerMock)
	}
	type args struct {
		directory string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []*entry
		wantErr bool
	}{
		{
			name: "success reading empty dir",
			fields: fields{
				hidden:           nil,
				deleteDuplicates: nil,
				silent:           nil,
				fileInfoBuilder: func(builder *fileInfoMock) {
					builder.On("IsDir").Return(true)
					builder.On("Name").Return("someDir")
				},
				openerBuilder: func(mocks []fs.FileInfo) func(builder *fileHandlerMock) {
					return func(builder *fileHandlerMock) {
						builder.On("Readdir", 0).Return(mocks, nil).Once()
						builder.On("Readdir", 0).Return([]fs.FileInfo{}, nil).Once()
						builder.On("Close").Return(nil).Once()
						builder.On("Close").Return(nil).Once()
					}
				},
			},
			args:    args{directory: "./"},
			wantErr: false,
		},
		{
			name: "success reading dir with files",
			fields: fields{
				hidden:           nil,
				deleteDuplicates: nil,
				silent:           nil,
				fileInfoBuilder: func(builder *fileInfoMock) {
					builder.On("IsDir").Return(false)
					builder.On("Name").Return("someFile")
				},
				openerBuilder: func(mocks []fs.FileInfo) func(builder *fileHandlerMock) {
					return func(builder *fileHandlerMock) {
						builder.On("Readdir", 0).Return(mocks, nil).Once()
						builder.On("Close").Return(nil).Once()
					}
				},
			},
			args: args{directory: "./"},
			want: []*entry{
				&entry{
					FileInfo: &fileInfoMock{},
					fullPath: "./someFile",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileInfo := NewFileInfoMock(tt.fields.fileInfoBuilder)
			defer fileInfo.AssertExpectations(t)

			fileHandler := NewHandlerMock(tt.fields.openerBuilder([]fs.FileInfo{fileInfo}))
			defer fileHandler.AssertExpectations(t)

			p := &processor{
				hidden:           tt.fields.hidden,
				deleteDuplicates: tt.fields.deleteDuplicates,
				silent:           tt.fields.silent,
				opener: func(name string) (handlerIface, error) {
					return fileHandler, nil
				},
			}
			if err := p.Read(tt.args.directory); (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
			}

			for i := range tt.want {
				assert.Equal(t, tt.want[i].GetPath(), p.GetEntries()[i].GetPath())
			}
		})
	}
}

func Test_processor_compare(t *testing.T) {
	type fields struct {
		entriesBuilders    []func(builder *fileInfoMock)
		duplicatesBuilders []func(builder *fileInfoMock)
	}
	type args struct {
		eBuilder func(builder *fileInfoMock)
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "success",
			fields: fields{
				entriesBuilders: []func(builder *fileInfoMock){
					func(builder *fileInfoMock) {
						builder.On("Name").Return("someName")
					},
					func(builder *fileInfoMock) {
						builder.On("Name").Return("someOther")
					},
					func(builder *fileInfoMock) {
						builder.On("Name").Return("someAnother")
					},
					func(builder *fileInfoMock) {},
				},
				duplicatesBuilders: []func(builder *fileInfoMock){
					func(builder *fileInfoMock) {
						builder.On("Name").Return("someName")
					},
				},
			},
			args: args{
				eBuilder: func(builder *fileInfoMock) {
					builder.On("Name").Return("someName")
				},
			},
			wantErr: func(t assert.TestingT, err error, i ...interface{}) bool {
				return err == nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entriesMock := []*fileInfoMock{}
			for i := range tt.fields.entriesBuilders {
				entriesMock = append(entriesMock, NewFileInfoMock(tt.fields.entriesBuilders[i]))
			}

			entries := []*entry{}
			for i := range entriesMock {
				entries = append(entries, &entry{
					index:    i,
					FileInfo: entriesMock[i],
				})
			}
			duplicatesMock := []*fileInfoMock{}
			for i := range tt.fields.duplicatesBuilders {
				duplicatesMock = append(duplicatesMock, NewFileInfoMock(tt.fields.duplicatesBuilders[i]))
			}

			duplicates := []*entry{}
			for i := range duplicatesMock {
				duplicates = append(duplicates, &entry{
					index:    i,
					FileInfo: duplicatesMock[i],
				})
			}

			p := &processor{
				entries: entries,
			}

			eMock := NewFileInfoMock(tt.args.eBuilder)
			defer eMock.AssertExpectations(t)

			e := &entry{
				index:    len(entries) - 1,
				FileInfo: eMock,
			}

			tt.wantErr(t, p.compare(e), fmt.Sprintf("compare(%v)", e))

			for i := range duplicates {
				assert.Equal(t, duplicates[i].Name(), p.duplicates[i].Name())

			}

			for i := range entriesMock {
				entriesMock[i].AssertExpectations(t)
			}

			for i := range duplicatesMock {
				duplicatesMock[i].AssertExpectations(t)
			}
		})
	}
}
