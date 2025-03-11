package tos

import "github.com/GBA-BI/tes-filer/pkg/transput"

type Config struct {
	CredentialFilePath string
	ExpirationFilePath string

	transput.S3SDKConfig
}
