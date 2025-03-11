package s3

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"

	"github.com/GBA-BI/tes-filer/pkg/transput"
	"github.com/GBA-BI/tes-filer/pkg/viper"
)

func NewCustomProvider(credentialFilePath, expirationFilePath string) credentials.Provider {
	return &customProvider{
		credentialFilePath: credentialFilePath,
		expirationFilePath: expirationFilePath,
	}
}

type customProvider struct {
	credentialFilePath string
	expirationFilePath string

	expirationTime time.Time
}

func (p *customProvider) Retrieve() (credentials.Value, error) {
	sConfig := &transput.S3SecretConfig{}
	if err := viper.SetConfigFromFileINI(p.credentialFilePath, "", sConfig); err != nil {
		return credentials.Value{}, err
	}

	data, err := os.ReadFile(p.expirationFilePath)
	if err != nil {
		return credentials.Value{}, fmt.Errorf("failed to read expirationFilePath: %w", err)
	}
	eConfig := &transput.S3ExpirationConfig{
		ExpiredTime: string(data),
	}

	expirationTime, err := time.Parse(time.RFC3339, eConfig.ExpiredTime)
	if err != nil {
		return credentials.Value{}, fmt.Errorf("parse experation time error: %w", err)
	}

	p.expirationTime = expirationTime
	newCredentials := credentials.Value{
		AccessKeyID:     sConfig.AccessKey,
		SecretAccessKey: sConfig.SecretKey,
		SessionToken:    sConfig.CreToken,
	}

	return newCredentials, nil
}

func (p *customProvider) IsExpired() bool {
	return p.expirationTime.Before(time.Now())
}
