package repo

import (
	"bufio"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/GBA-BI/tes-filer/internal/domain"
	"github.com/GBA-BI/tes-filer/pkg/consts"
	apperror "github.com/GBA-BI/tes-filer/pkg/error"
	"github.com/GBA-BI/tes-filer/pkg/log"
	"github.com/GBA-BI/tes-filer/pkg/transput"
	"github.com/GBA-BI/tes-filer/pkg/transput/drs"
	"github.com/GBA-BI/tes-filer/pkg/transput/file"
	"github.com/GBA-BI/tes-filer/pkg/transput/ftp"
	"github.com/GBA-BI/tes-filer/pkg/transput/http"
	"github.com/GBA-BI/tes-filer/pkg/transput/s3"
	"github.com/GBA-BI/tes-filer/pkg/transput/tos"
	utilspath "github.com/GBA-BI/tes-filer/pkg/utils/path"
	"github.com/GBA-BI/tes-filer/pkg/utils/retry"
	"github.com/GBA-BI/tes-filer/pkg/viper"
)

func NewFilerRepo(cfg *Config, logger log.Logger) (domain.Filer, error) {
	if cfg == nil {
		return nil, apperror.NewInvalidArgumentError("repo.Config", "")
	}
	return &filerRepo{
		transputFactory: newTransputFactory(cfg, logger),

		offloadType: cfg.OffloadType,
		logger:      logger,
		isMountTOS:  strings.ToLower(cfg.IsMountTOS) == "true",
	}, nil
}

type filerRepo struct {
	transputFactory *transputFactory

	offloadType string

	logger log.Logger

	isMountTOS bool
}

func (r *filerRepo) BuildFromFile(ctx context.Context, path string, mode string) (*domain.FileDirs, error) {
	inputsStr, outputsStr, err := r.readFile(path)
	if err != nil {
		return nil, err
	}
	createList := &struct {
		Inputs  []*domain.CreateFileDirParam `json:"inputs"`
		Outputs []*domain.CreateFileDirParam `json:"outputs"`
	}{}

	inputFileDirList := make([]*domain.FileDir, 0)
	outputFileDirList := make([]*domain.FileDir, 0)
	fileDirFactory := domain.NewFileDirFactory()
	// todo may be not gen input output both each time
	if len(inputsStr) > 0 {
		if err := json.Unmarshal([]byte(inputsStr), createList); err != nil {
			return nil, apperror.NewInternalError(err)
		}
		for _, param := range createList.Inputs {
			tempParam := param
			tempFileDir, err := fileDirFactory.New(tempParam)
			if err != nil {
				return nil, err
			}
			inputFileDirList = append(inputFileDirList, tempFileDir)
		}
	}

	if len(outputsStr) > 0 {
		if err := json.Unmarshal([]byte(outputsStr), createList); err != nil {
			return nil, apperror.NewInternalError(err)
		}
		for _, param := range createList.Outputs {
			tempParam := param
			tempFileDir, err := fileDirFactory.New(tempParam)
			if err != nil {
				return nil, err
			}
			outputFileDirList = append(outputFileDirList, tempFileDir)
		}
	}

	fileDirsFactory := domain.NewFileDirsFactory()
	return fileDirsFactory.New(inputFileDirList, outputFileDirList, mode)
}

func (r *filerRepo) Transput(ctx context.Context, fileDirs *domain.FileDirs) error {
	startTime := time.Now()
	switch fileDirs.Mode {
	case consts.TransputModeOutputs:
		return r.upload(ctx, fileDirs)
	case consts.TransputModeInputs:
		return r.download(ctx, fileDirs)
	case consts.TransputModeAll:
		// do download and then upload
		if err := r.download(ctx, fileDirs); err != nil {
			return err
		}
		if err := r.upload(ctx, fileDirs); err != nil {
			return err
		}

	}
	endTime := time.Now()
	duration := endTime.Sub(startTime)
	r.logger.Debugf("task time-consuming: %s", duration.String())
	r.logger.Infof("finish task %s", fileDirs.Mode)
	return nil
}

func (r *filerRepo) upload(ctx context.Context, fileDirs *domain.FileDirs) error {
	if fileDirs == nil || len(fileDirs.Outputs) == 0 {
		r.logger.Infof("parsed filedirs is empty, no need to upload")
		return nil
	}

	for _, fileDir := range fileDirs.Outputs {
		if err := retry.MountTOSRetry(r.logger, r.isMountTOS, func() error {
			return r.uploadFileDir(ctx, fileDir)
		}); err != nil {
			return err
		}
	}

	return nil
}

func (r *filerRepo) uploadFileDir(ctx context.Context, fileDir *domain.FileDir) error {
	if r.checkFinished(fileDir, string(consts.TransputModeOutputs)) {
		r.logger.Infof("already upload file %s", fileDir.Path)
		return nil
	}
	// if file not exist , log warning and skip
	_, err := os.Stat(fileDir.Path)
	if os.IsNotExist(err) {
		r.logger.Warnf("upload %s: %s not exist, just skip", fileDir.Typ, fileDir.Path)
		return nil
	}
	if err != nil {
		return apperror.NewInternalError(err)
	}
	transput, err := r.transputFactory.NewTransput(fileDir.Scheme, fileDir.UserInfo)
	if err != nil {
		return err
	}
	if fileDir.Typ == consts.FileTypeDir {
		r.logger.Infof("start uploading dir %s to url %s ", fileDir.Path, fileDir.URLForLog())
		if err := transput.UploadDir(ctx, fileDir.Path, fileDir.URL); err != nil {
			return apperror.NewInternalError(err)
		}
		r.logger.Infof("finish uploading dir %s to url %s", fileDir.Path, fileDir.URLForLog())
	}
	if fileDir.Typ == consts.FileTypeFile {
		r.logger.Infof("start uploading file %s to url %s", fileDir.Path, fileDir.URLForLog())
		if err := transput.UploadFile(ctx, fileDir.Path, fileDir.URL); err != nil {
			return apperror.NewInternalError(err)
		}
		r.logger.Infof("finish uploading file %s to url %s", fileDir.Path, fileDir.URLForLog())
	}
	r.setFinished(fileDir, string(consts.TransputModeOutputs))
	return nil
}

func (r *filerRepo) download(ctx context.Context, fileDirs *domain.FileDirs) error {
	if fileDirs == nil || len(fileDirs.Inputs) == 0 {
		r.logger.Infof("parsed filedirs is empty, no need to download")
		return nil
	}

	for _, fileDir := range fileDirs.Inputs {
		if err := retry.MountTOSRetry(r.logger, r.isMountTOS, func() error {
			return r.downloadFileDir(ctx, fileDir)
		}); err != nil {
			return err
		}
	}
	return nil
}

func (r *filerRepo) downloadFileDir(ctx context.Context, fileDir *domain.FileDir) error {
	if r.checkFinished(fileDir, string(consts.TransputModeInputs)) {
		r.logger.Infof("already download file %s", fileDir.Path)
		return nil
	}
	transput, err := r.transputFactory.NewTransput(fileDir.Scheme, fileDir.UserInfo)
	if err != nil {
		return err
	}
	if fileDir.Typ == consts.FileTypeDir {
		r.logger.Infof("start downloading dir %s from url %s", fileDir.Path, fileDir.URLForLog())
		if err := transput.DownloadDir(ctx, fileDir.Path, fileDir.URL); err != nil {
			return apperror.NewInternalError(err)
		}
		r.logger.Infof("finish downloading dir %s from url %s", fileDir.Path, fileDir.URLForLog())
	}
	if fileDir.Typ == consts.FileTypeFile {
		r.logger.Infof("start downloading file %s from url %s", fileDir.Path, fileDir.URLForLog())
		if err := transput.DownloadFile(ctx, fileDir.Path, fileDir.URL); err != nil {
			return apperror.NewInternalError(err)
		}
		r.logger.Infof("finish downloading file %s from url %s", fileDir.Path, fileDir.URLForLog())
	}
	r.setFinished(fileDir, string(consts.TransputModeInputs))
	return nil
}

func (r *filerRepo) readFile(path string) (string, string, error) {
	exist, err := utilspath.FileExists(path)
	if err != nil {
		return "", "", apperror.NewInternalError(err)
	}
	if !exist {
		return "", "", apperror.NewNotFoundError("File", path)
	}

	file, err := os.Open(path)
	if err != nil {
		return "", "", apperror.NewInternalError(err)
	}
	defer file.Close()

	var inputsStr, outputsStr string

	scanner := bufio.NewScanner(file)
	// The OffloadThreshold is 100KiB. Due to escape and other characters,
	// the maximum line length should be less than 200KiB.
	// Example:
	// {"inputs":[{"":{}}]} -> task-inputs="{\"inputs\":[{\"\":{}}]}"
	// Set the maxTokenSize of scanner as 200KiB, because default 64KiB
	// is too small.
	scanner.Buffer(make([]byte, 4096), 204800)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "=", 2)
		name, annoWithQuote := parts[0], parts[1]

		var anno string
		if err := json.Unmarshal([]byte(annoWithQuote), &anno); err != nil {
			return "", "", apperror.NewInternalError(err)
		}

		switch name {
		case "task-inputs":
			inputsStr = anno
		case "task-outputs":
			outputsStr = anno
		case "task-inputs-ref":
			if r.offloadType == consts.OffloadTypePVC {
				inputsStr, err = r.getFromPVC(anno)
				if err != nil {
					return "", "", err
				}
			}
		case "task-outputs-ref":
			if r.offloadType == consts.OffloadTypePVC {
				outputsStr, err = r.getFromPVC(anno)
				if err != nil {
					return "", "", err
				}
			}
		}
	}
	return inputsStr, outputsStr, nil
}

func (r *filerRepo) getFromPVC(ref string) (string, error) {
	exist, err := utilspath.FileExists(ref)
	if err != nil {
		return "{}", apperror.NewInternalError(err)
	}
	if !exist {
		return "{}", nil
	}

	content, err := os.ReadFile(ref)
	if err != nil {
		return "", apperror.NewInternalError(err)
	}

	return string(content), nil
}

func (r *filerRepo) setFinished(fileDir *domain.FileDir, mode string) {
	finishPath := genFinishPath(fileDir.Path, string(fileDir.Scheme), fileDir.URL, mode)
	if _, err := os.Create(finishPath); err != nil {
		r.logger.Debugf("Unable to create finish file: %v", err)
	}
}

func (r *filerRepo) checkFinished(fileDir *domain.FileDir, mode string) bool {
	finishPath := genFinishPath(fileDir.Path, string(fileDir.Scheme), fileDir.URL, mode)
	if _, err := os.Stat(finishPath); err != nil {
		if !os.IsNotExist(err) {
			r.logger.Debugf("Unable to check finishPath %s: %v", finishPath, err)
		}
		return false
	}
	return true
}

func newTransputFactory(cfg *Config, logger log.Logger) *transputFactory {
	return &transputFactory{
		s3ConfigPath:         cfg.S3ConfigPath,
		expirationConfigPath: cfg.ExpirationConfigPath,
		s3SecretPath:         cfg.S3SecretPath,
		transputMap:          sync.Map{},
		logger:               logger,
	}
}

type transputFactory struct {
	s3ConfigPath         string
	expirationConfigPath string
	s3SecretPath         string
	transputMap          sync.Map
	logger               log.Logger
}

func (t *transputFactory) NewTransput(schema consts.Scheme, userInfo *url.Userinfo) (transput.Transput, error) {
	if trans, ok := t.transputMap.Load(schema); ok {
		return trans.(transput.Transput), nil
	}
	var newTrans transput.Transput
	var err error

	switch schema {
	case consts.SchemeHTTP:
		cfg := &http.Config{}
		newTrans, err = http.NewHTTPTransput(cfg)
	case consts.SchemeDRS:
		cfg := &drs.Config{}
		viper.SetConfigFromEnv(cfg)
		newTrans, err = drs.NewDRSTransput(cfg, t.logger)
	case consts.SchemeFTP:
		cfg := &ftp.Config{}
		viper.SetConfigFromEnv(cfg)
		newTrans, err = ftp.NewFTPTransput(cfg)
	case consts.SchemeFILE:
		cfg := &file.Config{}
		viper.SetConfigFromEnv(cfg)
		newTrans, err = file.NewFileTransput(cfg, t.logger)
	case consts.SchemeS3:
		s3SDKConfig := &transput.S3SDKConfig{}
		if err := viper.SetConfigFromFileINI(t.s3ConfigPath, "", s3SDKConfig); err != nil {
			return nil, apperror.NewInternalError(err)
		}

		if strings.ToLower(s3SDKConfig.S3Type) == strings.ToLower(string(consts.SchemeTOS)) {
			cfg := &tos.Config{
				CredentialFilePath: t.s3SecretPath,
				ExpirationFilePath: t.expirationConfigPath,

				S3SDKConfig: *s3SDKConfig,
			}
			newTrans, err = tos.NewTOSTransput(cfg, userInfo, t.logger)
		} else {
			cfg := &s3.Config{
				CredentialFilePath: t.s3SecretPath,
				ExpirationFilePath: t.expirationConfigPath,

				S3SDKConfig: *s3SDKConfig,
			}
			newTrans, err = s3.NewS3Transput(cfg, userInfo)
		}
	default:
		return nil, apperror.NewInvalidArgumentError("transput.Scheme", string(schema))
	}
	if err != nil {
		return nil, apperror.NewInternalError(err)
	}
	t.transputMap.Store(schema, newTrans)
	return newTrans, nil
}

func genFinishPath(path, scheme, url, mode string) string {
	md5hash := md5.Sum([]byte(path + url))
	md5hashStr := hex.EncodeToString(md5hash[:])
	finishName := fmt.Sprintf(".%s-%s-%s.finish", mode, scheme, md5hashStr)
	return filepath.Join(filepath.Dir(path), finishName)
}
