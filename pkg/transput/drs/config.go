package drs

type Config struct {
	InsecureDirDomain string `env:"INSECURE_DIR_DOMAIN"`
	AAIPassport       string `env:"AAI_PASSPORT"`
}
