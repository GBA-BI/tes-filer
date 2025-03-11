package md5

import (
	"crypto/md5"
	"fmt"
	"os"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

func TestMD5Checker(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fileContent := "Hello, world!"
	filePath := tempDir + "/testfile"
	err = os.WriteFile(filePath, []byte(fileContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	hash := md5.Sum([]byte(fileContent))
	md5sum := fmt.Sprintf("%x", hash)

	tests := []struct {
		name      string
		checksum  string
		path      string
		expected  bool
		expectErr bool
	}{
		{
			name:      "valid checksum",
			checksum:  md5sum,
			path:      filePath,
			expected:  true,
			expectErr: false,
		},
		{
			name:      "invalid checksum",
			checksum:  "invalidchecksum",
			path:      filePath,
			expected:  false,
			expectErr: false,
		},
	}

	for _, tc := range tests {
		convey.Convey(tc.name, t, func() {
			checker := NewMD5Checker(tc.checksum)
			result, err := checker.Check(tc.path)
			if tc.expectErr {
				convey.So(err, convey.ShouldNotBeNil)
			} else {
				convey.So(err, convey.ShouldBeNil)
				convey.So(result, convey.ShouldEqual, tc.expected)
			}
		})
	}
}
