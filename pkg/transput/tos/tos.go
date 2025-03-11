package tos

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"

	"github.com/GBA-BI/tes-filer/pkg/consts"
	"github.com/GBA-BI/tes-filer/pkg/log"
	"github.com/GBA-BI/tes-filer/pkg/transput"
	utilspath "github.com/GBA-BI/tes-filer/pkg/utils/path"
	utilsstrings "github.com/GBA-BI/tes-filer/pkg/utils/strings"
	"github.com/GBA-BI/tes-filer/pkg/viper"
)

const (
	tosMaximumPartNum  = 10000
	tosMaximumPartSize = 5 * 1024 * 1024 * 1024 // 5GiB
)

type tosTransput struct {
	transput.DefaultTransput
	client                              *tos.ClientV2
	uploadEventListenerAndRateLimiter   *uploadEventListenerAndRateLimiter
	downloadEventListenerAndRateLimiter *downloadEventListenerAndRateLimiter

	partSize int64
	taskNum  int64

	logger log.Logger
}

func NewTOSTransput(cfg *Config, userInfo *url.Userinfo, logger log.Logger) (transput.Transput, error) {
	if cfg == nil {
		return nil, fmt.Errorf("nil config of tos transput")
	}

	fCredentials, err := getCre(cfg, userInfo)

	if err != nil {
		return nil, fmt.Errorf("failed to init credential: %w", err)
	}

	maxRetryCount := consts.DefaultRetryCount
	if cfg.MaxRetryCount > 0 {
		maxRetryCount = int(cfg.MaxRetryCount)
	}
	var partSize int64 = consts.DefaultPartSize
	if cfg.PartSize > 0 {
		partSize = cfg.PartSize
	}
	client, err := tos.NewClientV2(cfg.Endpoint,
		tos.WithRegion(cfg.Region),
		tos.WithCredentials(fCredentials),
		tos.WithEnableCRC(cfg.EnableCRC),
		tos.WithMaxRetryCount(maxRetryCount))
	if err != nil {
		return nil, fmt.Errorf("init tos client failed: %w", err)
	}

	return &tosTransput{
		client:                              client,
		partSize:                            partSize,
		taskNum:                             cfg.TaskNum,
		logger:                              logger,
		uploadEventListenerAndRateLimiter:   newUploadEventListenerAndRateLimiter(cfg.MaxBandwidth, cfg.MaxBandwidth, logger),
		downloadEventListenerAndRateLimiter: newDownloadEventListenerAndRateLimiter(cfg.MaxBandwidth, cfg.MaxBandwidth, logger),
	}, nil
}

func getCre(cfg *Config, userInfo *url.Userinfo) (tos.Credentials, error) {
	var fCredentials tos.Credentials
	var err error

	if userInfo != nil {
		accessKey := userInfo.Username()
		if secretKey, ok := userInfo.Password(); ok {
			fCredentials = tos.NewStaticCredentials(accessKey, secretKey)
			return fCredentials, nil
		}
	}

	var needRefreshCre = false
	if cfg.ExpirationFilePath != "" {
		exist, err := utilspath.FileExists(cfg.ExpirationFilePath)
		if err != nil {
			return nil, err
		}
		if exist {
			needRefreshCre = true
		}
	}

	if needRefreshCre {
		fCredentials, err = tos.NewFederationCredentials(NewFederationToken(cfg.CredentialFilePath, cfg.ExpirationFilePath))
	} else {
		sConfig := &transput.S3SecretConfig{}
		if err := viper.SetConfigFromFileINI(cfg.CredentialFilePath, "", sConfig); err != nil {
			return nil, err
		}
		fCredentials = tos.NewStaticCredentials(sConfig.AccessKey, sConfig.SecretKey)
	}
	return fCredentials, err
}

func (t *tosTransput) UploadDir(ctx context.Context, local, remote string) error {
	return transput.CommonUploadDir(ctx, local, remote, t)
}

func (t *tosTransput) DownloadDir(ctx context.Context, local, remote string) error {
	local = utilsstrings.CheckDir(local)
	remote = utilsstrings.CheckDir(remote)
	bucket, remotePath, err := utilspath.ParseURL(remote)
	if err != nil {
		return fmt.Errorf("failed to parse url %w", err)
	}
	truncated := true
	continuationToken := ""
	subFileList := make([]string, 0)
	subDirList := make([]string, 0)
	for truncated {
		output, err := t.client.ListObjectsType2(ctx, &tos.ListObjectsType2Input{
			Bucket:            bucket,
			ContinuationToken: continuationToken,
			Delimiter:         "/",
			Prefix:            remotePath,
		})
		if err != nil {
			return fmt.Errorf("failed to list file of tos: %w", err)
		}
		for _, prefix := range output.CommonPrefixes {
			subDirList = append(subDirList, filepath.Base(prefix.Prefix))
		}
		for _, obj := range output.Contents {
			if !utilsstrings.IsDir(obj.Key) {
				subFileList = append(subFileList, filepath.Base(obj.Key))
			}
		}
		truncated = output.IsTruncated
		continuationToken = output.NextContinuationToken
	}

	for _, obj := range subFileList {
		remoteObj := fmt.Sprintf("%s%s", remote, obj)
		localObj := fmt.Sprintf("%s%s", local, obj)
		if err := t.DownloadFile(ctx, localObj, remoteObj); err != nil {
			return err
		}
	}

	for _, prefix := range subDirList {
		remotePath := fmt.Sprintf("%s%s", remote, prefix)
		localPath := fmt.Sprintf("%s%s", local, prefix)
		if err := t.DownloadDir(ctx, localPath, remotePath); err != nil {
			return err
		}
	}

	return nil
}

func (t *tosTransput) UploadFile(ctx context.Context, local, remote string) error {
	bucket, object, err := utilspath.ParseURL(remote)
	if err != nil {
		return fmt.Errorf("failed to parse url of tos while uploading file: %w", err)
	}

	stat, err := os.Stat(local)
	if err != nil {
		return fmt.Errorf("failed to stat file of path %s: %w", local, err)
	}

	var uploadErr error

	fileSize := stat.Size()
	for {
		if fileSize < t.partSize {
			// do not slice to be compatible with cromwell
			t.logger.Debugf("file %s size is %d, less than partSize %d, no need to do multipart", local, fileSize, t.partSize)
			fileReader, err := os.Open(local)
			if err != nil {
				return fmt.Errorf("unable to open file %s: %w", local, err)
			}
			_, uploadErr = t.client.PutObjectV2(ctx, &tos.PutObjectV2Input{
				PutObjectBasicInput: tos.PutObjectBasicInput{
					Bucket:      bucket,
					Key:         object,
					RateLimiter: t.uploadEventListenerAndRateLimiter,
				},
				Content: fileReader,
			})
		} else {
			partSize, err := getUploadPartSize(fileSize, t.partSize)
			if err != nil {
				return err
			}
			_, uploadErr = t.client.UploadFile(ctx, &tos.UploadFileInput{
				CreateMultipartUploadV2Input: tos.CreateMultipartUploadV2Input{
					Bucket: bucket,
					Key:    object,
				},
				FilePath:            local,
				PartSize:            partSize,
				TaskNum:             int(t.taskNum),
				EnableCheckpoint:    true,
				UploadEventListener: t.uploadEventListenerAndRateLimiter,
				RateLimiter:         t.uploadEventListenerAndRateLimiter,
			})
		}

		if uploadErr == nil {
			return nil
		}

		if !t.handleUploadRateLimitError(uploadErr) {
			return fmt.Errorf("failed to upload file to tos: %w", uploadErr)
		}
	}
}

func getUploadPartSize(fileSize, defaultPartSize int64) (int64, error) {
	// ceil divide
	minimumPartSize := (fileSize-1)/tosMaximumPartNum + 1
	if minimumPartSize > tosMaximumPartSize {
		return 0, fmt.Errorf("fileSize too large, fileSize: %d, maximumPartSize: %d, maximumPartNum: %d", fileSize, tosMaximumPartSize, tosMaximumPartNum)
	}
	for {
		if minimumPartSize <= defaultPartSize {
			return defaultPartSize, nil
		}
		defaultPartSize *= 2
		if defaultPartSize > tosMaximumPartSize {
			return tosMaximumPartSize, nil
		}
	}
}

func (t *tosTransput) DownloadFile(ctx context.Context, local, remote string) error {
	basedir := filepath.Dir(local)
	if err := os.MkdirAll(basedir, os.FileMode(consts.DefaultFileMode)); err != nil {
		return fmt.Errorf("failed to mkdir: %w", err)
	}
	bucket, object, err := utilspath.ParseURL(remote)
	if err != nil {
		return fmt.Errorf("failed to parse bucket and object from url: %w", err)
	}

	for {
		_, downloadErr := t.client.DownloadFile(ctx, &tos.DownloadFileInput{
			HeadObjectV2Input: tos.HeadObjectV2Input{
				Bucket: bucket,
				Key:    object,
			},
			FilePath:              local,
			PartSize:              t.partSize,
			TaskNum:               int(t.taskNum),
			EnableCheckpoint:      true,
			DownloadEventListener: t.downloadEventListenerAndRateLimiter,
			RateLimiter:           t.downloadEventListenerAndRateLimiter,
		})

		if downloadErr == nil {
			return nil
		}

		if !t.handleDownloadRateLimitError(downloadErr) {
			return fmt.Errorf("failed to download file from tos: %w", downloadErr)
		}
	}
}

func (t *tosTransput) handleUploadRateLimitError(err error) bool {
	if err == nil {
		return true
	}

	// adapt to PutObjectV2
	if isRateLimitError(err) {
		return true
	}

	return t.uploadEventListenerAndRateLimiter.onlyOccurRateLimitErr()
}

func (t *tosTransput) handleDownloadRateLimitError(err error) bool {
	if err == nil {
		return true
	}

	// adapt to HeadObject in DownloadFile
	if isRateLimitError(err) {
		return true
	}

	return t.downloadEventListenerAndRateLimiter.onlyOccurRateLimitErr()
}

func isRateLimitError(err error) bool {
	// the err may be TosServerError or UnexpectedStatusCodeError, so we should check
	// statusCode, not code in TosServerError
	return tos.StatusCode(err) == http.StatusTooManyRequests
}
