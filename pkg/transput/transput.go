package transput

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	utilspath "github.com/GBA-BI/tes-filer/pkg/utils/path"
	utilsstrings "github.com/GBA-BI/tes-filer/pkg/utils/strings"
)

type Transput interface {
	UploadDir(ctx context.Context, local, remote string) error
	DownloadDir(ctx context.Context, local, remote string) error
	UploadFile(ctx context.Context, local, remote string) error
	DownloadFile(ctx context.Context, local, remote string) error
}

type DefaultTransput struct{}

func (d *DefaultTransput) DownloadDir(ctx context.Context, local, remote string) error {
	panic("not implemented")
}

func (d *DefaultTransput) UploadFile(ctx context.Context, local, remote string) error {
	panic("not implemented")
}

func (d *DefaultTransput) DownloadFile(ctx context.Context, local, remote string) error {
	panic("not implemented")
}

func (d *DefaultTransput) UploadDir(ctx context.Context, local, remote string) error {
	panic("not implemented")
}

func CommonUploadDir(ctx context.Context, local, remote string, transput Transput) error {
	return filepath.Walk(local, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		fileInfo, err := os.Lstat(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				// skip the deleting file
				return nil
			}
			return err
		}

		// dir
		if fileInfo.IsDir() {
			return nil
		}

		// symlink
		if fileInfo.Mode()&os.ModeSymlink != 0 {
			realPath, err := utilspath.GetRealPathOfLink(path)
			if err != nil {
				return err
			}

			realFileInfo, err := os.Stat(realPath)
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					// skip invalid symlink
					return nil
				}
				return err
			}

			dstPath := fmt.Sprintf("%s%s", utilsstrings.CheckDir(remote), filepath.Base(path))
			if realFileInfo.IsDir() {
				return CommonUploadDir(ctx, realPath, dstPath, transput)
			} else {
				return transput.UploadFile(ctx, realPath, dstPath)
			}
		}

		// file
		relativePath, err := filepath.Rel(local, path)
		if err != nil {
			return err
		}

		dstPath := fmt.Sprintf("%s%s", utilsstrings.CheckDir(remote), relativePath)
		return transput.UploadFile(ctx, path, dstPath)
	})
}
