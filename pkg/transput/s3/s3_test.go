package s3

import (
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/smartystreets/goconvey/convey"
)

func TestS3Transput_UploadFile(t *testing.T) {
	tests := []struct {
		name      string
		local     string
		remote    string
		expectErr bool
	}{
		{
			name:      "successfully upload file",
			local:     "/path/to/local",
			remote:    "s3://bucketName/objectName",
			expectErr: false,
		},
		{
			name:      "failed to upload file",
			local:     "/path/to/local",
			remote:    "s3://bucketName/objectName",
			expectErr: true,
		},
	}

	for _, tc := range tests {
		convey.Convey(tc.name, t, func() {
			s3Trans := &s3Transput{
				uploader: &s3manager.Uploader{},
			}

			patch1 := gomonkey.ApplyFunc(os.Open, func(_ string) (*os.File, error) {
				if tc.expectErr {
					return nil, fmt.Errorf("failed to open file")
				}
				return &os.File{}, nil
			})
			defer patch1.Reset()

			patch2 := gomonkey.ApplyMethod(reflect.TypeOf(*s3Trans.uploader), "UploadWithContext", func(_ s3manager.Uploader, _ context.Context, _ *s3manager.UploadInput, _ ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
				if tc.expectErr {
					return nil, fmt.Errorf("failed to upload with context")
				}
				return &s3manager.UploadOutput{}, nil
			})
			defer patch2.Reset()

			err := s3Trans.UploadFile(context.Background(), tc.local, tc.remote)
			if tc.expectErr {
				convey.So(err, convey.ShouldNotBeNil)
			} else {
				convey.So(err, convey.ShouldBeNil)
			}
		})
	}
}

func TestS3Transput_DownloadFile(t *testing.T) {
	tests := []struct {
		name      string
		local     string
		remote    string
		expectErr bool
	}{
		{
			name:      "successfully download file",
			local:     "/path/to/local",
			remote:    "s3://bucketName/objectName",
			expectErr: false,
		},
		{
			name:      "failed to download file",
			local:     "/path/to/local",
			remote:    "s3://bucketName/objectName",
			expectErr: true,
		},
	}

	for _, tc := range tests {
		convey.Convey(tc.name, t, func() {
			s3Trans := &s3Transput{
				downloader: &s3manager.Downloader{},
			}

			patch1 := gomonkey.ApplyFunc(os.MkdirAll, func(_ string, _ os.FileMode) error {
				if tc.expectErr {
					return fmt.Errorf("failed to mkdir all")
				}
				return nil
			})
			defer patch1.Reset()

			patch2 := gomonkey.ApplyFunc(os.Create, func(_ string) (*os.File, error) {
				if tc.expectErr {
					return nil, fmt.Errorf("failed to create file")
				}
				return &os.File{}, nil
			})
			defer patch2.Reset()

			patch3 := gomonkey.ApplyMethod(reflect.TypeOf(*s3Trans.downloader), "DownloadWithContext", func(_ s3manager.Downloader, _ context.Context, _ io.WriterAt, _ *s3.GetObjectInput, _ ...func(*s3manager.Downloader)) (int64, error) {
				if tc.expectErr {
					return 0, fmt.Errorf("failed to download with context")
				}
				return 0, nil
			})
			defer patch3.Reset()

			err := s3Trans.DownloadFile(context.Background(), tc.local, tc.remote)
			if tc.expectErr {
				convey.So(err, convey.ShouldNotBeNil)
			} else {
				convey.So(err, convey.ShouldBeNil)
			}
		})
	}
}

func TestS3Transput_DownloadDir(t *testing.T) {
	tests := []struct {
		name      string
		local     string
		remote    string
		expectErr bool
	}{
		{
			name:      "successfully download directory",
			local:     "/path/to/local",
			remote:    "s3://bucketName/objectPrefix",
			expectErr: false,
		},
		{
			name:      "failed to download directory",
			local:     "/path/to/local",
			remote:    "s3://bucketName/objectPrefix",
			expectErr: true,
		},
	}

	for _, tc := range tests {
		convey.Convey(tc.name, t, func() {
			s3Trans := &s3Transput{
				client: &s3.S3{},
			}

			patch1 := gomonkey.ApplyMethod(reflect.TypeOf(s3Trans.client), "ListObjectsWithContext", func(_ *s3.S3, _ context.Context, _ *s3.ListObjectsInput, _ ...request.Option) (*s3.ListObjectsOutput, error) {
				if tc.expectErr {
					return nil, fmt.Errorf("failed to list objects")
				}
				return &s3.ListObjectsOutput{}, nil
			})
			defer patch1.Reset()

			patch2 := gomonkey.ApplyFunc(os.MkdirAll, func(_ string, _ os.FileMode) error {
				if tc.expectErr {
					return fmt.Errorf("failed to mkdir all")
				}
				return nil
			})
			defer patch2.Reset()

			patch3 := gomonkey.ApplyMethod(reflect.TypeOf(s3Trans), "DownloadFile", func(_ *s3Transput, _ context.Context, _ string, _ string) error {
				if tc.expectErr {
					return fmt.Errorf("failed to download file")
				}
				return nil
			})
			defer patch3.Reset()

			err := s3Trans.DownloadDir(context.Background(), tc.local, tc.remote)
			if tc.expectErr {
				convey.So(err, convey.ShouldNotBeNil)
			} else {
				convey.So(err, convey.ShouldBeNil)
			}
		})
	}
}
