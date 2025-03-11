package http

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/GBA-BI/tes-filer/pkg/consts"
	"github.com/GBA-BI/tes-filer/pkg/transput"
)

func NewHTTPTransput(cfg *Config) (transput.Transput, error) {
	if cfg == nil {
		return nil, fmt.Errorf("nil config of http transput")
	}
	return &httpTransput{
		client:  &http.Client{},
		headers: cfg.Headers,
	}, nil
}

type httpTransput struct {
	transput.DefaultTransput

	headers map[string]string

	client *http.Client
}

func (h *httpTransput) UploadDir(ctx context.Context, local, remote string) error {
	return transput.CommonUploadDir(ctx, local, remote, h)
}

func (h *httpTransput) UploadFile(ctx context.Context, local, remote string) error {
	fileData, err := os.ReadFile(local)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, remote, bytes.NewReader(fileData))
	if err != nil {
		return err
	}

	for k, v := range h.headers {
		req.Header.Set(k, v)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("upload file error with status code：%d", resp.StatusCode)
	}

	return nil
}

func (h *httpTransput) DownloadFile(ctx context.Context, local, remote string) error {
	basedir := filepath.Dir(local)
	if err := os.MkdirAll(basedir, os.FileMode(consts.DefaultFileMode)); err != nil {
		return fmt.Errorf("failed to mkdir: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, remote, nil)
	if err != nil {
		return err
	}

	for k, v := range h.headers {
		req.Header.Set(k, v)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download file error with status code: %d", resp.StatusCode)
	}

	out, err := os.OpenFile(local, os.O_CREATE|os.O_WRONLY, os.FileMode(consts.DefaultFileMode))
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
