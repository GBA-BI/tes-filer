package ftp

import (
	"context"
	"errors"
	"io"
	"os"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/jlaffaye/ftp"
	"github.com/smartystreets/goconvey/convey"
)

func TestFtpTransput_UploadFile(t *testing.T) {
	tests := []struct {
		name      string
		local     string
		remote    string
		osOpenErr error
		storErr   error
		expectErr bool
	}{

		{
			name:      "os.Open error",
			local:     "localfile",
			remote:    "remotefile",
			osOpenErr: errors.New("os open error"),
			storErr:   nil,
			expectErr: true,
		},
		{
			name:      "Stor error",
			local:     "localfile",
			remote:    "remotefile",
			osOpenErr: nil,
			storErr:   errors.New("stor error"),
			expectErr: true,
		}, {
			name:      "successful upload",
			local:     "localfile",
			remote:    "remotefile",
			osOpenErr: nil,
			storErr:   nil,
			expectErr: false,
		},
	}

	for _, tc := range tests {
		convey.Convey(tc.name, t, func() {
			outputs := []gomonkey.OutputCell{
				{Values: gomonkey.Params{nil, tc.osOpenErr}},
			}
			patch1 := gomonkey.ApplyFuncSeq(os.Open, outputs)
			defer patch1.Reset()

			ftpTrans := &ftpTransput{
				conn: &ftp.ServerConn{},
			}
			patch2 := gomonkey.ApplyMethod(reflect.TypeOf(ftpTrans.conn), "Stor", func(_ *ftp.ServerConn, _ string, _ io.Reader) error {
				return tc.storErr
			})
			defer patch2.Reset()

			err := ftpTrans.UploadFile(context.Background(), tc.local, tc.remote)
			if tc.expectErr {
				convey.So(err, convey.ShouldNotBeNil)
			} else {
				convey.So(err, convey.ShouldBeNil)
			}
		})
	}
}

func TestFtpTransput_DownloadFile(t *testing.T) {

	tests := []struct {
		name      string
		local     string
		remote    string
		mkdirErr  error
		retrErr   error
		createErr error
		copyErr   error
		expectErr bool
	}{
		{
			name:      "failed to mkdir",
			local:     "/invalid/path/to/localfile",
			remote:    "/path/to/remotefile",
			mkdirErr:  errors.New("failed to mkdir"),
			expectErr: true,
		},
		{
			name:      "connect error",
			local:     "/path/to/localfile",
			remote:    "/path/to/remotefile",
			mkdirErr:  nil,
			retrErr:   errors.New("connect error"),
			expectErr: true,
		},
		{
			name:      "create file error",
			local:     "/path/to/localfile",
			remote:    "/path/to/remotefile",
			mkdirErr:  nil,
			retrErr:   nil,
			createErr: errors.New("create file error"),
			expectErr: true,
		},
		{
			name:      "copy error",
			local:     "/path/to/localfile",
			remote:    "/path/to/remotefile",
			mkdirErr:  nil,
			retrErr:   nil,
			createErr: nil,
			copyErr:   errors.New("copy error"),
			expectErr: true,
		},
		{
			name:      "successful download",
			local:     "/path/to/localfile",
			remote:    "/path/to/remotefile",
			mkdirErr:  nil,
			retrErr:   nil,
			createErr: nil,
			copyErr:   nil,
			expectErr: false,
		},
	}

	for _, tc := range tests {
		convey.Convey(tc.name, t, func() {
			ftpTrans := &ftpTransput{
				conn: &ftp.ServerConn{},
			}

			patch1 := gomonkey.ApplyFunc(os.MkdirAll, func(path string, perm os.FileMode) error {
				return tc.mkdirErr
			})
			defer patch1.Reset()

			tempResp := &ftp.Response{}

			patch6 := gomonkey.ApplyMethod(reflect.TypeOf(tempResp), "Close", func(_ *ftp.Response) error {
				return nil
			})
			defer patch6.Reset()

			patch2 := gomonkey.ApplyMethod(reflect.TypeOf(ftpTrans.conn), "Retr", func(_ *ftp.ServerConn, path string) (*ftp.Response, error) {
				return tempResp, tc.retrErr
			})
			defer patch2.Reset()

			tempFile := &os.File{}

			patch5 := gomonkey.ApplyMethod(reflect.TypeOf(tempFile), "Close", func(_ *os.File) error {
				return nil
			})
			defer patch5.Reset()

			patch3 := gomonkey.ApplyFunc(os.Create, func(name string) (*os.File, error) {
				return tempFile, tc.createErr
			})
			defer patch3.Reset()

			patch4 := gomonkey.ApplyFunc(io.Copy, func(dst io.Writer, src io.Reader) (int64, error) {
				return 0, tc.copyErr
			})
			defer patch4.Reset()

			err := ftpTrans.DownloadFile(context.Background(), tc.local, tc.remote)

			if tc.expectErr {
				convey.So(err, convey.ShouldNotBeNil)
			} else {
				convey.So(err, convey.ShouldBeNil)
			}
		})
	}
}

func TestFtpTransput_DownloadDir(t *testing.T) {
	tests := []struct {
		name      string
		local     string
		remote    string
		expectErr bool
	}{
		{
			name:      "successfully download directory",
			local:     ".",
			remote:    "/path/to/remote",
			expectErr: false,
		},
		{
			name:      "failed to download directory",
			local:     ".",
			remote:    "/path/to/remote",
			expectErr: true,
		},
	}

	for _, tc := range tests {
		convey.Convey(tc.name, t, func() {
			ftpTrans := &ftpTransput{
				conn: &ftp.ServerConn{},
			}

			patch1 := gomonkey.ApplyMethod(reflect.TypeOf(ftpTrans.conn), "List", func(_ *ftp.ServerConn, _ string) ([]*ftp.Entry, error) {
				return []*ftp.Entry{
					{Type: ftp.EntryTypeFile, Name: "file1"},
				}, nil
			})
			defer patch1.Reset()

			patch2 := gomonkey.ApplyMethod(reflect.TypeOf(ftpTrans), "DownloadFile", func(_ *ftpTransput, _ context.Context, _ string, _ string) error {
				if tc.expectErr {
					return errors.New("download error")
				}
				return nil
			})
			defer patch2.Reset()

			err := ftpTrans.DownloadDir(context.Background(), tc.local, tc.remote)
			if tc.expectErr {
				convey.So(err, convey.ShouldNotBeNil)
			} else {
				convey.So(err, convey.ShouldBeNil)
			}
		})
	}
}
