package application

import (
	"context"
	"fmt"

	"github.com/GBA-BI/tes-filer/internal/domain"
	apperror "github.com/GBA-BI/tes-filer/pkg/error"
)

type TransputCmd struct {
	path string
	mode string

	fileDirsRepo domain.Filer
}

func NewTransputCmd(c *Config, fileDirsRepo domain.Filer) (*TransputCmd, error) {
	if c == nil {
		return nil, apperror.NewInternalError(fmt.Errorf("nil config of application transput"))
	}
	return &TransputCmd{
		path:         c.Path,
		mode:         c.Mode,
		fileDirsRepo: fileDirsRepo,
	}, nil
}

func (t *TransputCmd) Transput(ctx context.Context) error {
	fileDirs, err := t.fileDirsRepo.BuildFromFile(ctx, t.path, t.mode)
	if err != nil {
		return err
	}

	if err := t.fileDirsRepo.Transput(ctx, fileDirs); err != nil {
		return err
	}

	return nil
}
