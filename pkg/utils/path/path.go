package path

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func GetRealPathOfLink(path string) (string, error) {
	linkPath, err := os.Readlink(path)
	if err != nil {
		return "", err
	}

	realPath := linkPath
	if !filepath.IsAbs(linkPath) {
		realPath = filepath.Join(filepath.Dir(path), linkPath)
	}

	realPath, err = filepath.Abs(realPath)
	if err != nil {
		return "", err
	}
	return realPath, nil
}

func ParseURL(rawURL string) (string, string, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return "", "", err
	}
	host := parsedURL.Host
	prefix := strings.TrimPrefix(parsedURL.Path, "/")
	return host, prefix, nil
}

func FileExists(filename string) (bool, error) {
	_, err := os.Stat(filename)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
