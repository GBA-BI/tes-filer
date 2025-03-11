package domain

import (
	"context"
	"net/url"
	"strings"

	"github.com/jinzhu/copier"

	"github.com/GBA-BI/tes-filer/pkg/consts"
	apperror "github.com/GBA-BI/tes-filer/pkg/error"
)

type FileDir struct {
	Name        string
	Description string
	URL         string
	Path        string

	Typ      consts.FileType
	Scheme   consts.Scheme
	UserInfo *url.Userinfo
}

type FileDirs struct {
	Mode consts.TransputMode

	Inputs  []*FileDir
	Outputs []*FileDir
}

func (f *FileDir) SetTyp(typ string) error {
	switch strings.ToLower(typ) {
	case "file":
		f.Typ = consts.FileTypeFile
	case "directory":
		f.Typ = consts.FileTypeDir
	default:
		return apperror.NewInvalidArgumentError("FileDir.Typ", typ)
	}
	return nil
}

func (f *FileDir) Complete() error {
	parsedURL, err := url.Parse(f.URL)
	if err != nil {
		return apperror.NewInvalidArgumentError("FileDir.URL", f.URL)
	}

	switch strings.ToLower(parsedURL.Scheme) {
	case "s3":
		f.Scheme = consts.SchemeS3
	case "drs":
		f.Scheme = consts.SchemeDRS
	case "http", "https":
		f.Scheme = consts.SchemeHTTP
	case "tos":
		f.Scheme = consts.SchemeTOS
	case "ftp":
		f.Scheme = consts.SchemeFTP
	case "file", "":
		f.Scheme = consts.SchemeFILE
	default:
		return apperror.NewInvalidArgumentError("FileDir.Scheme", parsedURL.Scheme)
	}

	f.UserInfo = parsedURL.User
	return nil
}

func (f *FileDirs) SetMode(mode string) error {
	switch strings.ToLower(mode) {
	case "inputs":
		f.Mode = consts.TransputModeInputs
	case "outputs":
		f.Mode = consts.TransputModeOutputs
	case "all":
		f.Mode = consts.TransputModeAll
	default:
		return apperror.NewInvalidArgumentError("FileDirs.Mode", mode)
	}
	return nil
}

func (f *FileDir) URLForLog() string {
	if f.UserInfo == nil {
		return f.URL
	}
	parsedURL, _ := url.Parse(f.URL)
	parsedURL.User = nil
	return parsedURL.String()
}

// Factory // hackable
type CreateFileDirParam struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Path        string `json:"path"`
	Typ         string `json:"type"`
}

type FileDirFactory interface {
	New(param *CreateFileDirParam) (*FileDir, error)
}

type fileDirFactoryImpl struct{}

func NewFileDirFactory() FileDirFactory {
	return &fileDirFactoryImpl{}
}

func (i *fileDirFactoryImpl) New(param *CreateFileDirParam) (*FileDir, error) {
	fileDir := &FileDir{}
	if err := copier.Copy(fileDir, param); err != nil {
		return nil, apperror.NewInternalError(err)
	}
	if err := fileDir.SetTyp(param.Typ); err != nil {
		return nil, err
	}
	if err := fileDir.Complete(); err != nil {
		return nil, err
	}
	return fileDir, nil
}

type FileDirsFactory interface {
	New(inputs []*FileDir, outputs []*FileDir, mode string) (*FileDirs, error)
}

func NewFileDirsFactory() FileDirsFactory {
	return &fileDirsFactoryImpl{}
}

type fileDirsFactoryImpl struct{}

func (i *fileDirsFactoryImpl) New(inputs []*FileDir, outputs []*FileDir, mode string) (*FileDirs, error) {
	fileDirs := &FileDirs{
		Inputs:  inputs,
		Outputs: outputs,
	}
	if err := fileDirs.SetMode(mode); err != nil {
		return fileDirs, nil
	}
	return fileDirs, nil
}

// repo
type Filer interface {
	BuildFromFile(ctx context.Context, path string, mode string) (*FileDirs, error)
	Transput(ctx context.Context, fileDirs *FileDirs) error
}
