package s3

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"golang.org/x/time/rate"

	"github.com/GBA-BI/tes-filer/pkg/consts"
	"github.com/GBA-BI/tes-filer/pkg/transput"
	utilspath "github.com/GBA-BI/tes-filer/pkg/utils/path"
	utilsstrings "github.com/GBA-BI/tes-filer/pkg/utils/strings"
	"github.com/GBA-BI/tes-filer/pkg/viper"
)

const batchSize = 1000

type s3Transput struct {
	transput.DefaultTransput

	client     s3iface.S3API
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
	limiter    *rate.Limiter

	bandwidth        int64
	limitErrCount    int64
	lastLimitErrTime time.Time
}

func NewS3Transput(cfg *Config, userInfo *url.Userinfo) (transput.Transput, error) {
	if cfg == nil {
		return nil, fmt.Errorf("nil s3 transput config")
	}

	maxBandwidth := cfg.MaxBandwidth
	if cfg.MaxBandwidth == 0 {
		maxBandwidth = consts.DefaultMaxBandwidth
	}
	sharedLimiter := rate.NewLimiter(rate.Limit(maxBandwidth), int(maxBandwidth))

	// Create an HTTP client with the rate limiter.
	httpClient := &http.Client{
		Transport: &rateLimitingTransport{
			upLimiter:   sharedLimiter,
			downLimiter: sharedLimiter,
			transport:   http.DefaultTransport,
		},
	}

	cre, err := getCre(cfg, userInfo)
	if err != nil {
		return nil, err
	}

	maxRetryCount := consts.DefaultRetryCount
	if cfg.MaxRetryCount > 0 {
		maxRetryCount = int(cfg.MaxRetryCount)
	}
	var partSize int64 = consts.DefaultPartSize
	if cfg.PartSize > 0 {
		partSize = cfg.PartSize
	}
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(cfg.Region),
		Endpoint:    aws.String(cfg.Endpoint),
		Credentials: cre,
		MaxRetries:  aws.Int(maxRetryCount),
		HTTPClient:  httpClient,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create s3 transput: %w", err)
	}

	return &s3Transput{
		uploader: s3manager.NewUploader(sess, func(u *s3manager.Uploader) {
			u.PartSize = partSize
		}),
		downloader: s3manager.NewDownloader(sess, func(u *s3manager.Downloader) {
			u.PartSize = partSize
		}),
		client:        s3.New(sess),
		limiter:       sharedLimiter,
		bandwidth:     cfg.MaxBandwidth,
		limitErrCount: 0,
	}, nil
}

func getCre(cfg *Config, userInfo *url.Userinfo) (*credentials.Credentials, error) {
	if userInfo != nil {
		accessKey := userInfo.Username()
		if secretKey, ok := userInfo.Password(); ok {
			return credentials.NewStaticCredentials(accessKey, secretKey, ""), nil
		}
	}

	var needRefreshCre bool = false
	if cfg.ExpirationFilePath != "" {
		exist, err := utilspath.FileExists(cfg.ExpirationFilePath)
		if err != nil {
			return nil, err
		}
		if exist {
			needRefreshCre = true
		}
	}

	var cre *credentials.Credentials
	if needRefreshCre {
		cre = credentials.NewCredentials(NewCustomProvider(cfg.CredentialFilePath, cfg.ExpirationFilePath))
	} else {
		sConfig := &transput.S3SecretConfig{}
		if err := viper.SetConfigFromFileINI(cfg.CredentialFilePath, "", sConfig); err != nil {
			return nil, err
		}
		cre = credentials.NewStaticCredentials(sConfig.AccessKey, sConfig.SecretKey, sConfig.CreToken)
	}
	return cre, nil
}

func (t *s3Transput) DownloadDir(ctx context.Context, local, remote string) error {
	bucketName, objectPrefix, err := utilspath.ParseURL(remote)
	if err != nil {
		return err
	}
	objects, err := t.listObjects(ctx, bucketName, &objectPrefix, false)
	if err != nil {
		return err
	}
	for _, obj := range objects {
		pureObj := strings.TrimPrefix(obj, objectPrefix)
		filePath := path.Join(local, pureObj)
		remotePath := fmt.Sprintf("%s%s", consts.S3Prefix, path.Join(bucketName, obj))
		fileDir := path.Dir(filePath)
		if err := os.MkdirAll(fileDir, os.FileMode(consts.DefaultFileMode)); err != nil {
			return fmt.Errorf("failed to mkdir: %w", err)
		}
		if err := t.DownloadFile(ctx, filePath, remotePath); err != nil {
			return err
		}
	}
	return nil
}

func (t *s3Transput) UploadFile(ctx context.Context, local, remote string) error {
	// 5 minutes do not trigger rate limit error, restore
	if time.Since(t.lastLimitErrTime).Minutes() > 5 {
		t.updateRateLimit(t.bandwidth)
	}
	bucketName, objectName, err := utilspath.ParseURL(remote)
	if err != nil {
		return err
	}

	fileReader, err := os.Open(local)
	if err != nil {
		return fmt.Errorf("unable to open file, %w", err)
	}

	for {
		_, uploadErr := t.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
			Bucket: aws.String(bucketName),
			Key:    aws.String(objectName),
			Body:   fileReader,
		})

		if uploadErr == nil {
			return nil
		}

		if !t.handleRateLimitError(uploadErr) {
			return fmt.Errorf("failed to upload file from s3: %w", uploadErr)
		}
	}

}

func (t *s3Transput) UploadDir(ctx context.Context, local, remote string) error {
	return transput.CommonUploadDir(ctx, local, remote, t)
}

func (t *s3Transput) DownloadFile(ctx context.Context, local, remote string) error {
	// 5 minutes do not trigger rate limit error, restore
	if time.Since(t.lastLimitErrTime).Minutes() > 5 {
		t.updateRateLimit(t.bandwidth)
	}
	basedir := filepath.Dir(local)
	if err := os.MkdirAll(basedir, os.FileMode(consts.DefaultFileMode)); err != nil {
		return fmt.Errorf("failed to mkdir: %w", err)
	}
	bucketName, objectName, err := utilspath.ParseURL(remote)
	if err != nil {
		return err
	}

	file, err := os.Create(local)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	for {
		_, downloadErr := t.downloader.DownloadWithContext(ctx, file, &s3.GetObjectInput{
			Bucket: &bucketName,
			Key:    &objectName,
		})

		if downloadErr == nil {
			return nil
		}

		if !t.handleRateLimitError(downloadErr) {
			return fmt.Errorf("failed to download file from s3: %w", downloadErr)
		}
	}
}

func (t *s3Transput) listObjects(ctx context.Context, bucketName string, prefix *string, withDir bool) ([]string, error) {
	var res []string
	if err := t.listObjectAndForeach(ctx, bucketName, prefix, func(object *s3.Object) error {
		if withDir || !utilsstrings.IsDir(*object.Key) {
			res = append(res, *object.Key)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return res, nil
}

func (t *s3Transput) listObjectAndForeach(ctx context.Context, bucketName string, prefix *string, fn func(*s3.Object) error) error {
	listInput := &s3.ListObjectsInput{
		Bucket:  aws.String(bucketName),
		Prefix:  prefix,
		MaxKeys: aws.Int64(batchSize),
	}
	for {
		listOutput, err := t.client.ListObjectsWithContext(ctx, listInput)
		if err != nil {
			return fmt.Errorf("failed to ListObjects of bucket %s: %w", bucketName, err)
		}
		for _, object := range listOutput.Contents {
			if object != nil {
				if err = fn(object); err != nil {
					return fmt.Errorf("deal s3 object %s fail: %w", *object.Key, err)
				}
			}
		}
		if listOutput.IsTruncated == nil || !*listOutput.IsTruncated {
			break
		}
		listInput.Marker = listOutput.NextMarker
	}
	return nil
}

func (t *s3Transput) updateRateLimit(newRate int64) {
	if newRate > consts.DefaultMinBandwidth {
		t.limiter.SetLimit(rate.Limit(newRate))
	}
}

func (t *s3Transput) handleRateLimitError(err error) bool {
	var awsErr awserr.Error
	ok := errors.As(err, &awsErr)
	if !ok {
		return false
	}

	if utilsstrings.Contains(consts.ErrCodeRateLimitList, awsErr.Code()) {
		t.limitErrCount++
		t.lastLimitErrTime = time.Now()
		t.updateRateLimit(t.bandwidth / (t.limitErrCount + 1))
		return true
	}

	return false
}
