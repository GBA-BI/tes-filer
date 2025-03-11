package file

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/GBA-BI/tes-filer/pkg/consts"
	"github.com/GBA-BI/tes-filer/pkg/log"
	"github.com/GBA-BI/tes-filer/pkg/transput"
	utilspath "github.com/GBA-BI/tes-filer/pkg/utils/path"
)

type fileTransput struct {
	transput.DefaultTransput

	hostBasePath      string
	containerBasePath string
	logger            log.Logger
}

func NewFileTransput(cfg *Config, logger log.Logger) (transput.Transput, error) {
	if cfg == nil {
		return nil, fmt.Errorf("nil config of file transput")
	}
	return &fileTransput{
		hostBasePath:      cfg.HostBasePath,
		containerBasePath: cfg.ContainerBasePath,
		logger:            logger,
	}, nil
}

func (ft *fileTransput) DownloadFile(ctx context.Context, local, remote string) error {
	urlContainerPath, err := ft.getContainerPathFromURL(remote)
	if err != nil {
		return err
	}
	ft.logger.Infof("Symlink %s to %s", urlContainerPath, local)
	return symlink(urlContainerPath, local)
}

func (ft *fileTransput) DownloadDir(ctx context.Context, local, remote string) error {
	urlContainerPath, err := ft.getContainerPathFromURL(remote)
	if err != nil {
		return err
	}
	ft.logger.Infof("Symlink %s to %s", urlContainerPath, local)
	return symlink(urlContainerPath, local)
}

func (ft *fileTransput) UploadFile(ctx context.Context, local, remote string) error {
	urlContainerPath, err := ft.getContainerPathFromURL(remote)
	if err != nil {
		return err
	}
	ft.logger.Infof("Copying %s to %s", local, urlContainerPath)
	return copyFile(local, urlContainerPath)
}

func (ft *fileTransput) UploadDir(ctx context.Context, local, remote string) error {
	urlContainerPath, err := ft.getContainerPathFromURL(remote)
	if err != nil {
		return err
	}
	ft.logger.Infof("Copying %s to %s", local, urlContainerPath)
	return copyDir(local, urlContainerPath)
}

func (ft *fileTransput) getContainerPathFromURL(urlStr string) (string, error) {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return "", fmt.Errorf("failed to get path: %w", err)
	}
	// if mount-tos, no need to limit url relative to hostBasePath
	if strings.HasPrefix(parsedURL.Path, "/tos-data/") {
		return parsedURL.Path, nil
	}
	relPath, err := filepath.Rel(ft.hostBasePath, parsedURL.Path)
	if err != nil {
		return "", err
	}
	if strings.HasPrefix(relPath, "..") {
		return "", fmt.Errorf("'%s' is not a descendant of 'HOST_BASE_PATH' (%s)", relPath, ft.hostBasePath)
	}
	return filepath.Join(ft.containerBasePath, relPath), nil
}

func copyContent(src, dst string) error {
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		fileInfo, err := os.Stat(srcPath)
		if err != nil {
			return err
		}

		if fileInfo.IsDir() {
			err = copyContent(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			err = copyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}

	}

	return nil
}

func copyDir(src, dst string) error {
	exist, err := utilspath.FileExists(dst)
	if err != nil {
		return err
	}
	if !exist {
		os.MkdirAll(dst, consts.DefaultFileMode)
	}
	return copyContent(src, dst)
}

func copyFile(src, dst string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	dstInfo, err := os.Stat(dst)
	if !os.IsNotExist(err) { // no logs return
		if sameFile := os.SameFile(srcInfo, dstInfo); sameFile {
			return nil
		}
	}

	dstDir := filepath.Dir(dst)
	if _, err := os.Stat(dstDir); os.IsNotExist(err) {
		os.MkdirAll(dstDir, consts.DefaultFileMode)
	}
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	err = dstFile.Sync()
	if err != nil {
		return err
	}

	return nil
}

func symlink(src, dst string) error {
	// check src exist
	if _, err := os.Stat(src); err != nil {
		return err
	}

	dstDir := filepath.Dir(dst)
	if _, err := os.Stat(dstDir); os.IsNotExist(err) {
		os.MkdirAll(dstDir, consts.DefaultFileMode)
	}
	return os.Symlink(src, dst)
}
