package consts

type FileType string

const (
	FileTypeFile FileType = "FILE"
	FileTypeDir  FileType = "DIRECTORY"
)

type TransputMode string

const (
	TransputModeInputs  TransputMode = "INPUTS"
	TransputModeOutputs TransputMode = "OUTPUTS"
	TransputModeAll     TransputMode = "ALL"
)

type Scheme string

const (
	SchemeHTTP Scheme = "HTTP"
	SchemeFTP  Scheme = "FTP"
	SchemeS3   Scheme = "S3"
	SchemeTOS  Scheme = "TOS"
	SchemeDRS  Scheme = "DRS"
	SchemeFILE Scheme = "FILE"
)

const (
	OffloadTypePVC string = "pvc"
	OffloadTypeSQL string = "sql"
)

const DefaultFileMode = 0777

const S3Prefix = "s3://"

const (
	ErrCodeExceedAccountQPSLimit  = "ExceedAccountQPSLimit"
	ErrCodeExceedAccountRateLimit = "ExceedAccountRateLimit"
	ErrCodeExceedBucketQPSLimit   = "ExceedBucketQPSLimit"
	ErrCodeExceedBucketRateLimit  = "ExceedBucketRateLimit"
)

var ErrCodeRateLimitList = []string{ErrCodeExceedAccountQPSLimit, ErrCodeExceedAccountRateLimit, ErrCodeExceedBucketQPSLimit, ErrCodeExceedBucketRateLimit}

const CheckerTypeMD5 = "md5"

const (
	DefaultMinBandwidth = 1024 * 1024       // 1MB/s
	DefaultMaxBandwidth = 128 * 1024 * 1024 // 128MB/s = 1Gbps, consider this as no limit
)

const (
	DefaultRetryCount = 5
	DefaultPartSize   = 64 * 1024 * 1024 // 64MiB
)
