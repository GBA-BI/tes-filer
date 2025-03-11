package drs

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	"github.com/smartystreets/goconvey/convey"

	"github.com/GBA-BI/tes-filer/pkg/checker"
	"github.com/GBA-BI/tes-filer/pkg/log"
	"github.com/GBA-BI/tes-filer/pkg/mock"
	"github.com/GBA-BI/tes-filer/pkg/transput"
	transputhttp "github.com/GBA-BI/tes-filer/pkg/transput/http"
)

func TestDrsTransput_getAccessURL(t *testing.T) {
	tests := []struct {
		name         string
		accessMethod AccessMethod
		hostName     string
		objectID     string
		expectErr    bool
	}{
		{
			name: "successfully get access URL",
			accessMethod: AccessMethod{
				AccessURL: AccessURL{URL: "http://remote.com"},
				AccessID:  "accessID1",
			},
			hostName:  "host1",
			objectID:  "objectID1",
			expectErr: false,
		},
		{
			name: "failed to get access URL",
			accessMethod: AccessMethod{
				AccessURL: AccessURL{URL: ""},
				AccessID:  "accessID2",
			},
			hostName:  "host2",
			objectID:  "objectID2",
			expectErr: true,
		},
	}

	for _, tc := range tests {
		convey.Convey(tc.name, t, func() {
			drsTrans := &drsTransput{
				insecureDirDomain: "insecureDirDomain",
				aaiPassport:       "aaiPassport",
				logger:            log.NewNopLogger(),
			}

			patch := gomonkey.ApplyFunc(http.Get, func(string) (*http.Response, error) {
				if tc.expectErr {
					return nil, fmt.Errorf("failed to get")
				}
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(`{"access_url": {"url": "http://remote.com"}}`)),
				}, nil
			})
			defer patch.Reset()

			_, err := drsTrans.getAccessURL(tc.accessMethod, tc.hostName, tc.objectID)
			if tc.expectErr {
				convey.So(err, convey.ShouldNotBeNil)
			} else {
				convey.So(err, convey.ShouldBeNil)
			}
		})
	}
}

func TestDrsTransput_pickAvailableTransputAndDownload(t *testing.T) {
	tests := []struct {
		name          string
		accessMethods []AccessMethod
		hostName      string
		objectID      string
		local         string
		expectErr     bool
	}{
		{
			name: "successfully pick available transput and download",
			accessMethods: []AccessMethod{
				{
					Type:      "https",
					AccessURL: AccessURL{URL: "http://remote.com"},
					AccessID:  "accessID1",
				},
			},
			hostName:  "host1",
			objectID:  "objectID1",
			local:     "/path/to/local",
			expectErr: false,
		},
		{
			name: "failed to pick available transput and download",
			accessMethods: []AccessMethod{
				{
					Type:      "ftp",
					AccessURL: AccessURL{URL: ""},
					AccessID:  "accessID2",
				},
			},
			hostName:  "host2",
			objectID:  "objectID2",
			local:     "/path/to/local",
			expectErr: true,
		},
	}

	for _, tc := range tests {
		convey.Convey(tc.name, t, func() {
			drsTrans := &drsTransput{
				insecureDirDomain: "insecureDirDomain",
				logger:            log.NewNopLogger(),
			}

			patch1 := gomonkey.ApplyPrivateMethod(reflect.TypeOf(drsTrans), "getAccessURL", func(_ *drsTransput, _ AccessMethod, _ string, _ string) (AccessURL, error) {
				if tc.expectErr {
					return AccessURL{}, fmt.Errorf("failed to get access url")
				}
				return AccessURL{URL: "http://remote.com"}, nil
			})
			defer patch1.Reset()

			defaultTransput := &transput.DefaultTransput{}

			patch2 := gomonkey.ApplyMethod(reflect.TypeOf(defaultTransput), "DownloadFile", func(_ *transput.DefaultTransput, _ context.Context, _ string, _ string) error {
				return nil
			})

			defer patch2.Reset()
			patch3 := gomonkey.ApplyFunc(transputhttp.NewHTTPTransput, func(_ *transputhttp.Config) (transput.Transput, error) {
				if tc.expectErr {
					return nil, fmt.Errorf("failed to new http transput")
				}
				return defaultTransput, nil
			})
			defer patch3.Reset()

			err := drsTrans.pickAvailableTransputAndDownload(context.Background(), tc.accessMethods, tc.hostName, tc.objectID, tc.local)
			if tc.expectErr {
				convey.So(err, convey.ShouldNotBeNil)
			} else {
				convey.So(err, convey.ShouldBeNil)
			}
		})
	}
}

func TestDrsTransput_DownloadFile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockChecker := mock.NewMockChecker(ctrl)
	mockChecker.EXPECT().Check(gomock.Any()).Return(true, nil)

	mockFile := mock.NewMockFileInfo(ctrl)
	mockFile.EXPECT().Size().Return(int64(100))

	tests := []struct {
		name      string
		local     string
		remote    string
		expectErr bool
	}{
		{
			name:      "successfully download file",
			local:     "/path/to/local",
			remote:    "http://remote.com/objectID1",
			expectErr: false,
		},
		{
			name:      "failed to download file",
			local:     "/path/to/local",
			remote:    "http://remote.com/objectID2",
			expectErr: true,
		},
	}

	for _, tc := range tests {
		convey.Convey(tc.name, t, func() {
			drsTrans := &drsTransput{
				insecureDirDomain: "insecureDirDomain",
				client:            &http.Client{},
				logger:            log.NewNopLogger(),
			}

			patch1 := gomonkey.ApplyFunc(http.NewRequest, func(_ string, _ string, _ io.Reader) (*http.Request, error) {
				if tc.expectErr {
					return nil, fmt.Errorf("failed to create new request")
				}
				return &http.Request{}, nil
			})
			defer patch1.Reset()

			patch2 := gomonkey.ApplyMethod(reflect.TypeOf(drsTrans.client), "Do", func(_ *http.Client, _ *http.Request) (*http.Response, error) {
				if tc.expectErr {
					return nil, fmt.Errorf("failed to do request")
				}
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(strings.NewReader(`{"access_methods": [{"type": "https", "access_url": {"url": "http://remote.com"}}], "size": 100, "checksums": [{"type": "md5", "checksum": "1a79a4d60de6718e8e5b326e338ae533"}]}`)),
				}, nil
			})
			defer patch2.Reset()

			patch3 := gomonkey.ApplyPrivateMethod(reflect.TypeOf(drsTrans), "pickAvailableChecker", func(_ *drsTransput, _ []Checksum) checker.Checker {
				if tc.expectErr {
					return nil
				}
				return mockChecker
			})
			defer patch3.Reset()

			patch4 := gomonkey.ApplyPrivateMethod(reflect.TypeOf(drsTrans), "pickAvailableTransputAndDownload", func(_ *drsTransput, _ context.Context, _ []AccessMethod, _ string, _ string, _ string) error {
				if tc.expectErr {
					return fmt.Errorf("failed to pick available transput and download")
				}
				return nil
			})
			defer patch4.Reset()

			patch5 := gomonkey.ApplyFunc(os.Stat, func(_ string) (os.FileInfo, error) {
				if tc.expectErr {
					return nil, fmt.Errorf("failed to stat")
				}
				return mockFile, nil
			})
			defer patch5.Reset()

			err := drsTrans.DownloadFile(context.Background(), tc.local, tc.remote)
			if tc.expectErr {
				convey.So(err, convey.ShouldNotBeNil)
			} else {
				convey.So(err, convey.ShouldBeNil)
			}
		})
	}
}
