package file

import (
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"github.com/GBA-BI/tes-filer/pkg/log"
)

func TestFileTransput_DownloadFile(t *testing.T) {
	tests := []struct {
		name      string
		local     string
		remote    string
		expectErr bool
	}{
		{
			name:      "successfully download file",
			local:     "/tmp/test",
			remote:    "file://remote/tmp/test",
			expectErr: false,
		},
		{
			name:      "failed to download file",
			local:     "/tmp/test",
			remote:    "invalid url",
			expectErr: true,
		},
	}

	for _, tc := range tests {
		convey.Convey(tc.name, t, func() {
			fileTrans := &fileTransput{
				hostBasePath:      "/tmp",
				containerBasePath: "/tmp",
				logger:            log.NewNopLogger(),
			}

			patch1 := gomonkey.ApplyFunc(symlink, func(src, dst string) error {
				if tc.expectErr {
					return fmt.Errorf("failed to symlink")
				}
				return nil
			})
			defer patch1.Reset()

			err := fileTrans.DownloadFile(context.Background(), tc.local, tc.remote)
			if tc.expectErr {
				convey.So(err, convey.ShouldNotBeNil)
			} else {
				convey.So(err, convey.ShouldBeNil)
			}
		})
	}
}

func TestFileTransput_DownloadDir(t *testing.T) {
	tests := []struct {
		name      string
		local     string
		remote    string
		expectErr bool
	}{
		{
			name:      "successfully download directory",
			local:     "/path/to/local",
			remote:    "path/path1",
			expectErr: false,
		},
		{
			name:      "failed to download directory",
			local:     "/path/to/local",
			remote:    "invalid url",
			expectErr: true,
		},
	}

	for _, tc := range tests {
		convey.Convey(tc.name, t, func() {
			fileTrans := &fileTransput{
				hostBasePath:      "./",
				containerBasePath: "/path/to/container/base",
				logger:            log.NewNopLogger(),
			}

			patch1 := gomonkey.ApplyFunc(symlink, func(src, dst string) error {
				if tc.expectErr {
					return fmt.Errorf("failed to symlink")
				}
				return nil
			})
			defer patch1.Reset()

			err := fileTrans.DownloadDir(context.Background(), tc.local, tc.remote)
			if tc.expectErr {
				convey.So(err, convey.ShouldNotBeNil)
			} else {
				convey.So(err, convey.ShouldBeNil)
			}
		})
	}
}

func TestFileTransput_UploadFile(t *testing.T) {
	tests := []struct {
		name      string
		local     string
		remote    string
		expectErr bool
	}{
		{
			name:      "successfully upload file",
			local:     "/tmp/test",
			remote:    "file://remote/tmp/test",
			expectErr: false,
		},
		{
			name:      "failed to upload file",
			local:     "/tmp/test",
			remote:    "invalid url",
			expectErr: true,
		},
	}

	for _, tc := range tests {
		convey.Convey(tc.name, t, func() {
			fileTrans := &fileTransput{
				hostBasePath:      "/tmp",
				containerBasePath: "/tmp",
				logger:            log.NewNopLogger(),
			}

			patch1 := gomonkey.ApplyFunc(os.Stat, func(string) (os.FileInfo, error) {
				if tc.expectErr {
					return nil, fmt.Errorf("failed to stat")
				}
				return nil, nil
			})
			defer patch1.Reset()

			tempFile := &os.File{}
			patch5 := gomonkey.ApplyMethod(reflect.TypeOf(tempFile), "Close", func(_ *os.File) error {
				return nil
			})
			defer patch5.Reset()

			patch6 := gomonkey.ApplyMethod(reflect.TypeOf(tempFile), "Sync", func(_ *os.File) error {
				return nil
			})
			defer patch6.Reset()

			patch2 := gomonkey.ApplyFunc(os.Open, func(string) (*os.File, error) {
				if tc.expectErr {
					return nil, fmt.Errorf("failed to open")
				}
				return tempFile, nil
			})
			defer patch2.Reset()

			patch3 := gomonkey.ApplyFunc(os.Create, func(string) (*os.File, error) {
				if tc.expectErr {
					return nil, fmt.Errorf("failed to create")
				}
				return tempFile, nil
			})
			defer patch3.Reset()

			patch4 := gomonkey.ApplyFunc(io.Copy, func(dst io.Writer, src io.Reader) (written int64, err error) {
				if tc.expectErr {
					return 0, fmt.Errorf("failed to copy")
				}
				return 0, nil
			})
			defer patch4.Reset()

			err := fileTrans.UploadFile(context.Background(), tc.local, tc.remote)
			if tc.expectErr {
				convey.So(err, convey.ShouldNotBeNil)
			} else {
				convey.So(err, convey.ShouldBeNil)
			}
		})
	}
}
