package http

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
)

func TestHttpTransput_UploadFile(t *testing.T) {
	tests := []struct {
		name      string
		local     string
		remote    string
		status    int
		expectErr bool
	}{
		{
			name:      "successfully upload file",
			local:     "/path/to/local",
			remote:    "http://remote.com",
			status:    http.StatusOK,
			expectErr: false,
		},
		{
			name:      "failed to upload file",
			local:     "/path/to/local",
			remote:    "http://remote.com",
			status:    http.StatusBadRequest,
			expectErr: true,
		},
	}

	for _, tc := range tests {
		convey.Convey(tc.name, t, func() {
			httpTrans := &httpTransput{
				client: &http.Client{},
				headers: map[string]string{
					"Content-Type": "application/json",
				},
			}

			patch1 := gomonkey.ApplyMethod(reflect.TypeOf(httpTrans.client), "Do", func(_ *http.Client, _ *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: tc.status,
					Body:       io.NopCloser(bytes.NewBufferString("")),
				}, nil
			})
			defer patch1.Reset()

			patch2 := gomonkey.ApplyFunc(os.ReadFile, func(_ string) ([]byte, error) {
				return []byte{}, nil
			})
			defer patch2.Reset()

			err := httpTrans.UploadFile(context.Background(), tc.local, tc.remote)
			if tc.expectErr {
				convey.So(err, convey.ShouldNotBeNil)
			} else {
				convey.So(err, convey.ShouldBeNil)
			}
		})
	}
}

func TestHttpTransput_DownloadFile(t *testing.T) {
	localFileName := "http-transput-temp"
	tests := []struct {
		name      string
		local     string
		remote    string
		status    int
		expectErr bool
	}{
		{
			name:      "successfully download file",
			local:     localFileName,
			remote:    "http://remote.com",
			status:    http.StatusOK,
			expectErr: false,
		},
		{
			name:      "failed to download file",
			local:     localFileName,
			remote:    "http://remote.com",
			status:    http.StatusBadRequest,
			expectErr: true,
		},
	}

	for _, tc := range tests {
		convey.Convey(tc.name, t, func() {
			httpTrans := &httpTransput{
				client: &http.Client{},
				headers: map[string]string{
					"Content-Type": "application/json",
				},
			}

			patch1 := gomonkey.ApplyMethod(reflect.TypeOf(httpTrans.client), "Do", func(_ *http.Client, _ *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: tc.status,
					Body:       io.NopCloser(bytes.NewBufferString("")),
				}, nil
			})
			defer patch1.Reset()

			err := httpTrans.DownloadFile(context.Background(), tc.local, tc.remote)
			if tc.expectErr {
				convey.So(err, convey.ShouldNotBeNil)
			} else {
				convey.So(err, convey.ShouldBeNil)
			}
			os.Remove(tc.local)
		})
	}
}
