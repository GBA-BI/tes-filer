package tos

import (
	"fmt"
	"os"
	"time"

	"github.com/volcengine/ve-tos-golang-sdk/v2/tos"

	"github.com/GBA-BI/tes-filer/pkg/transput"
	"github.com/GBA-BI/tes-filer/pkg/viper"
)

func NewFederationToken(credentialFilePath, expirationFilePath string) tos.FederationTokenProvider {
	return &federationTokenProvider{
		credentialFilePath: credentialFilePath,
		expirationFilePath: expirationFilePath,
	}
}

type federationTokenProvider struct {
	credentialFilePath string
	expirationFilePath string
}

func (f *federationTokenProvider) FederationToken() (*tos.FederationToken, error) {
	// load
	sConfig := &transput.S3SecretConfig{}
	if err := viper.SetConfigFromFileINI(f.credentialFilePath, "", sConfig); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(f.expirationFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read expirationFilePath: %w", err)
	}
	eConfig := &transput.S3ExpirationConfig{
		ExpiredTime: string(data),
	}

	expirationTime, err := time.Parse(time.RFC3339, eConfig.ExpiredTime)
	if err != nil {
		return nil, fmt.Errorf("parse experation time error: %w", err)
	}

	return &tos.FederationToken{
		Credential: tos.Credential{
			AccessKeyID:     sConfig.AccessKey,
			AccessKeySecret: sConfig.SecretKey,
			SecurityToken:   sConfig.CreToken,
		},
		Expiration: expirationTime,
	}, nil
}
