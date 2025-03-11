package transput

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
)

func TestCommonUploadDir(t *testing.T) {
	tests := []struct {
		name      string
		local     string
		remote    string
		expectErr bool
	}{
		{
			name:      "successfully upload directory",
			local:     ".",
			remote:    "remote",
			expectErr: false,
		},
		{
			name:      "failed to upload directory",
			local:     ".",
			remote:    "remote",
			expectErr: true,
		},
	}

	for _, tc := range tests {
		convey.Convey(tc.name, t, func() {
			transput := &DefaultTransput{}

			patch := gomonkey.ApplyMethod(reflect.TypeOf(transput), "UploadFile", func(_ *DefaultTransput, _ context.Context, _ string, _ string) error {
				if tc.expectErr {
					return errors.New("upload error")
				}
				return nil
			})
			defer patch.Reset()

			err := CommonUploadDir(context.Background(), tc.local, tc.remote, transput)
			if tc.expectErr {
				convey.So(err, convey.ShouldNotBeNil)
			} else {
				convey.So(err, convey.ShouldBeNil)
			}
		})
	}
}
