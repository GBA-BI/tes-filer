package ftp

type Config struct {
	URL       string `env:"FTP_URL"`
	AccessKey string `env:"FTP_ACCESS_KEY"`
	SecretKey string `env:"FTP_SECRET_KEY"`
}
