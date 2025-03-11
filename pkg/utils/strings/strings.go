package strings

import (
	"fmt"
	"strings"
)

func Contains(arr []string, target string) bool {
	for _, value := range arr {
		if value == target {
			return true
		}
	}
	return false
}

func CheckDir(dir string) string {
	if !strings.HasSuffix(dir, "/") {
		return fmt.Sprintf("%s/", dir)
	}
	return dir
}

func IsDir(key string) bool {
	return strings.HasSuffix(key, "/")
}
