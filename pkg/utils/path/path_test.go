package path

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestParseURL(t *testing.T) {
	tests := []struct {
		name         string
		rawURL       string
		expectedHost string
		expectedPath string
		expectErr    bool
	}{
		{
			name:         "valid URL",
			rawURL:       "http://example.com/path",
			expectedHost: "example.com",
			expectedPath: "path",
			expectErr:    false,
		},
		{
			name:         "valid URL without path",
			rawURL:       "http://example.com",
			expectedHost: "example.com",
			expectedPath: "",
			expectErr:    false,
		},
		{
			name:         "invalid URL",
			rawURL:       "://example.com/path",
			expectedHost: "",
			expectedPath: "",
			expectErr:    true,
		},
		{
			name:         "S3 validate URL",
			rawURL:       "s3://bioos-dev-wch4ejbdeig44addonemg/aasa1",
			expectedHost: "bioos-dev-wch4ejbdeig44addonemg",
			expectedPath: "aasa1",
			expectErr:    false,
		},
	}

	for _, tc := range tests {
		convey.Convey(tc.name, t, func() {
			host, path, err := ParseURL(tc.rawURL)
			if tc.expectErr {
				convey.So(err, convey.ShouldNotBeNil)
			} else {
				convey.So(err, convey.ShouldBeNil)
				convey.So(host, convey.ShouldEqual, tc.expectedHost)
				convey.So(path, convey.ShouldEqual, tc.expectedPath)
			}
		})
	}
}

func TestGetRealPathOfLink(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	realFile := filepath.Join(tempDir, "realfile")
	err = os.WriteFile(realFile, []byte("content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create real file: %v", err)
	}

	linkFile := filepath.Join(tempDir, "linkfile")
	err = os.Symlink(realFile, linkFile)
	if err != nil {
		t.Fatalf("Failed to create symbolic link: %v", err)
	}

	tests := []struct {
		name      string
		path      string
		expected  string
		expectErr bool
	}{
		{
			name:      "valid symbolic link",
			path:      linkFile,
			expected:  realFile,
			expectErr: false,
		},
		{
			name:      "invalid symbolic link",
			path:      "invalidpath",
			expected:  "",
			expectErr: true,
		},
	}

	for _, tc := range tests {
		convey.Convey(tc.name, t, func() {
			realPath, err := GetRealPathOfLink(tc.path)
			if tc.expectErr {
				convey.So(err, convey.ShouldNotBeNil)
			} else {
				convey.So(err, convey.ShouldBeNil)
				convey.So(realPath, convey.ShouldEqual, tc.expected)
			}
		})
	}
}
