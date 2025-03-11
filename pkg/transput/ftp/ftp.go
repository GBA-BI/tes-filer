package ftp

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/jlaffaye/ftp"

	"github.com/GBA-BI/tes-filer/pkg/consts"
	apperror "github.com/GBA-BI/tes-filer/pkg/error"
	"github.com/GBA-BI/tes-filer/pkg/transput"
)

func NewFTPTransput(cfg *Config) (transput.Transput, error) {
	if cfg == nil {
		return nil, apperror.NewInvalidArgumentError("FTPTransput", "Config")
	}

	conn, err := ftp.Dial(cfg.URL)
	if err != nil {
		return nil, err
	}

	err = conn.Login(cfg.AccessKey, cfg.SecretKey)
	if err != nil {
		return nil, err
	}

	return &ftpTransput{
		url:      cfg.URL,
		username: cfg.AccessKey,
		password: cfg.SecretKey,
		conn:     conn,
	}, nil
}

type ftpTransput struct {
	transput.DefaultTransput

	url      string
	username string
	password string
	conn     *ftp.ServerConn
}

func (t *ftpTransput) UploadDir(ctx context.Context, local, remote string) error {
	return transput.CommonUploadDir(ctx, local, remote, t)
}

func (t *ftpTransput) DownloadDir(ctx context.Context, local, remote string) error {
	entries, err := t.conn.List(remote)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(local, entry.Name)
		dstPath := filepath.Join(remote, entry.Name)

		if entry.Type == ftp.EntryTypeFolder {
			err = os.MkdirAll(srcPath, os.ModePerm)
			if err != nil {
				return err
			}

			err = t.DownloadDir(ctx, srcPath, dstPath)
			if err != nil {
				return err
			}
		} else if entry.Type == ftp.EntryTypeFile {
			err = t.DownloadFile(ctx, srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (t *ftpTransput) UploadFile(ctx context.Context, local, remote string) error {
	file, err := os.Open(local)
	if err != nil {
		return err
	}
	defer file.Close()

	err = t.conn.Stor(remote, file)
	if err != nil {
		return err
	}

	return nil
}

func (t *ftpTransput) DownloadFile(ctx context.Context, local, remote string) error {
	basedir := filepath.Dir(local)
	if err := os.MkdirAll(basedir, os.FileMode(consts.DefaultFileMode)); err != nil {
		return fmt.Errorf("failed to mkdir: %w", err)
	}

	resp, err := t.conn.Retr(remote)
	if err != nil {
		return fmt.Errorf("connect error: %w", err)
	}
	defer resp.Close()

	out, err := os.Create(local)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp)
	if err != nil {
		return fmt.Errorf("copy error:%w", err)
	}

	return nil
}
