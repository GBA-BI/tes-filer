package transput

type S3SDKConfig struct {
	S3Type        string `json:"s3_type" mapstructure:"s3_type"`
	Endpoint      string `json:"endpoint_url" mapstructure:"endpoint_url"`
	Region        string `json:"region" mapstructure:"region"`
	PartSize      int64  `json:"part_size" mapstructure:"part_size"`
	TaskNum       int64  `json:"task_num" mapstructure:"task_num"`
	EnableCRC     bool   `json:"enable_crc" mapstructure:"enable_crc"`
	MaxBandwidth  int64  `json:"max_band_width" mapstructure:"max_band_width"`
	MaxRetryCount int64  `json:"max_retry_count" mapstructure:"max_retry_count"`
}

type S3SecretConfig struct {
	AccessKey string `json:"aws_access_key_id" mapstructure:"aws_access_key_id"`
	SecretKey string `json:"aws_secret_access_key" mapstructure:"aws_secret_access_key"`
	CreToken  string `json:"aws_session_token" mapstructure:"aws_session_token"`
}

type S3ExpirationConfig struct {
	ExpiredTime string `json:"expiredTime" mapstructure:"expiredTime"`
}
