package tos

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	"github.com/smartystreets/goconvey/convey"
	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"

	"github.com/GBA-BI/tes-filer/pkg/consts"
	"github.com/GBA-BI/tes-filer/pkg/log"
	"github.com/GBA-BI/tes-filer/pkg/mock"
)

func TestTosTransput_UploadFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockFile1 := mock.NewMockFileInfo(ctrl)
	mockFile1.EXPECT().Size().Return(int64(100000000)).AnyTimes()
	mockFile2 := mock.NewMockFileInfo(ctrl)
	mockFile2.EXPECT().Size().Return(int64(10)).AnyTimes()

	tests := []struct {
		name      string
		local     string
		remote    string
		file      os.FileInfo
		expectErr bool
	}{
		{
			name:      "successfully upload file",
			local:     "/path/to/local",
			remote:    "s3://bucketName/objectName",
			file:      mockFile1,
			expectErr: false,
		},
		{
			name:      "failed to upload file",
			local:     "/path/to/local",
			remote:    "s3://bucketName/objectName",
			file:      mockFile1,
			expectErr: true,
		},
		{
			name:      "successfully upload file",
			local:     "/path/to/local",
			remote:    "s3://bucketName/objectName",
			file:      mockFile2,
			expectErr: false,
		},
		{
			name:      "failed to upload file",
			local:     "/path/to/local",
			remote:    "s3://bucketName/objectName",
			file:      mockFile2,
			expectErr: true,
		},
	}

	for _, tc := range tests {
		convey.Convey(tc.name, t, func() {
			tosTrans := &tosTransput{
				client:                            &tos.ClientV2{},
				logger:                            log.NewNopLogger(),
				uploadEventListenerAndRateLimiter: &uploadEventListenerAndRateLimiter{},
				partSize:                          consts.DefaultPartSize,
			}

			patch1 := gomonkey.ApplyMethod(reflect.TypeOf(tosTrans.client), "UploadFile", func(_ *tos.ClientV2, _ context.Context, _ *tos.UploadFileInput) (*tos.UploadFileOutput, error) {
				if tc.expectErr {
					return nil, fmt.Errorf("failed to upload file")
				}
				return &tos.UploadFileOutput{}, nil
			})
			defer patch1.Reset()

			patch2 := gomonkey.ApplyMethod(reflect.TypeOf(tosTrans.client), "PutObjectV2", func(_ *tos.ClientV2, _ context.Context, _ *tos.PutObjectV2Input) (*tos.PutObjectV2Output, error) {
				if tc.expectErr {
					return nil, fmt.Errorf("failed to upload file")
				}
				return &tos.PutObjectV2Output{}, nil
			})
			defer patch2.Reset()

			patch3 := gomonkey.ApplyFunc(os.Stat, func(string) (os.FileInfo, error) {
				return tc.file, nil
			})
			defer patch3.Reset()
			patch4 := gomonkey.ApplyFunc(os.Open, func(string) (*os.File, error) {
				return &os.File{}, nil
			})
			defer patch4.Reset()

			err := tosTrans.UploadFile(context.Background(), tc.local, tc.remote)
			if tc.expectErr {
				convey.So(err, convey.ShouldNotBeNil)
			} else {
				convey.So(err, convey.ShouldBeNil)
			}
		})
	}
}

func TestTosTransput_DownloadFile(t *testing.T) {
	tests := []struct {
		name      string
		local     string
		remote    string
		expectErr bool
	}{
		{
			name:      "successfully download file",
			local:     "/path/to/local",
			remote:    "tos://bucketName/objectName",
			expectErr: false,
		},
		{
			name:      "failed to download file",
			local:     "/path/to/local",
			remote:    "tos://bucketName/objectName",
			expectErr: true,
		},
	}

	for _, tc := range tests {
		convey.Convey(tc.name, t, func() {
			tosTrans := &tosTransput{
				client:                              &tos.ClientV2{},
				logger:                              log.NewNopLogger(),
				downloadEventListenerAndRateLimiter: &downloadEventListenerAndRateLimiter{},
			}

			patch1 := gomonkey.ApplyMethod(reflect.TypeOf(tosTrans.client), "DownloadFile", func(_ *tos.ClientV2, _ context.Context, _ *tos.DownloadFileInput) (*tos.DownloadFileOutput, error) {
				if tc.expectErr {
					return nil, fmt.Errorf("failed to download file")
				}
				return &tos.DownloadFileOutput{}, nil
			})
			defer patch1.Reset()

			patch2 := gomonkey.ApplyFunc(os.MkdirAll, func(_ string, _ os.FileMode) error {
				if tc.expectErr {
					return fmt.Errorf("failed to mkdir all")
				}
				return nil
			})
			defer patch2.Reset()

			err := tosTrans.DownloadFile(context.Background(), tc.local, tc.remote)
			if tc.expectErr {
				convey.So(err, convey.ShouldNotBeNil)
			} else {
				convey.So(err, convey.ShouldBeNil)
			}
		})
	}
}

func TestGetUploadPartSize(t *testing.T) {
	tests := []struct {
		name     string
		fileSize int64
		want     int64
		wantErr  bool
	}{
		{
			name:     "too large",
			fileSize: tosMaximumPartSize*tosMaximumPartNum + 1,
			want:     0,
			wantErr:  true,
		},
		{
			name:     "largest",
			fileSize: tosMaximumPartSize * tosMaximumPartNum,
			want:     tosMaximumPartSize,
			wantErr:  false,
		},
		{
			name:     "larger than default, divide exactly",
			fileSize: 1024 * 1024 * 1024 * tosMaximumPartNum,
			want:     1024 * 1024 * 1024,
			wantErr:  false,
		},
		{
			name:     "larger than default",
			fileSize: 1025 * 1024 * 1024 * tosMaximumPartNum,
			want:     2048 * 1024 * 1024,
			wantErr:  false,
		},
		{
			name:     "default, divide exactly",
			fileSize: consts.DefaultPartSize * tosMaximumPartNum,
			want:     consts.DefaultPartSize,
			wantErr:  false,
		},
		{
			name:     "default",
			fileSize: 300 * 1024 * 1024, // 300MiB
			want:     consts.DefaultPartSize,
			wantErr:  false,
		},
	}

	for _, tc := range tests {
		convey.Convey(tc.name, t, func() {
			partSize, err := getUploadPartSize(tc.fileSize, consts.DefaultPartSize)
			convey.So(err != nil, convey.ShouldEqual, tc.wantErr)
			convey.So(partSize, convey.ShouldEqual, tc.want)
		})
	}
}
