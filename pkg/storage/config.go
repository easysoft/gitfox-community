package storage

type DriverType string

const (
	DriverFilesystem DriverType = "filesystem"
)

type DriverConfig struct {
	Driver     DriverType
	Parameters map[string]interface{}
}

type Config struct {
	Local DriverConfig
	S3    DriverConfig
}
