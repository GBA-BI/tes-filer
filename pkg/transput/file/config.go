package file

type Config struct {
	HostBasePath      string `env:"HOST_BASE_PATH"`
	ContainerBasePath string `env:"CONTAINER_BASE_PATH"`
}
