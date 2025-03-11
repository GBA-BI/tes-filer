package md5

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"

	"github.com/GBA-BI/tes-filer/pkg/checker"
)

func NewMD5Checker(checksum string) checker.Checker {
	return &MD5Checker{
		checksum: checksum,
	}
}

type MD5Checker struct {
	checksum string
}

func (m *MD5Checker) Check(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, fmt.Errorf("failed to open file %s %w", path, err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return false, fmt.Errorf("failed to build hash %w", err)
	}

	md5sum := fmt.Sprintf("%x", hash.Sum(nil))
	return md5sum == m.checksum, nil
}
