package config

import (
	"errors"
	"net"
	"os"

	"github.com/pelletier/go-toml/v2"
)

const defaultConfig = `
wasmClusterAddr		= "localhost:6666"
wasmServerAddr 		= "localhost:5000"
apiServerAddr 		= "localhost:3000"
storageDriver 		= "sqlite"
apiToken 			= ""
authorization		= false
`

// Config holds the global configuration which is READONLY.
var config Config

type Storage struct {
	Name     string
	User     string
	Password string
	Host     string
	Port     string
	SSLMode  string
}

type Config struct {
	APIServerAddr   string
	WASMServerAddr  string
	StorageDriver   string
	WASMClusterAddr string
	APIToken        string
	Authorization   bool

	Storage Storage
}

func Parse(path string) error {
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile("config.toml", []byte(defaultConfig), os.ModePerm); err != nil {
			return err
		}
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	err = toml.Unmarshal(b, &config)
	return err
}

func Get() Config {
	return config
}

// makeURL takes a host address and returns a http URL.
func makeURL(address string) string {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		host = address
		port = ""
	}
	if host == "" {
		host = "0.0.0.0"
	}
	if port == "" || port == "http" {
		port = "80"
	}
	return "http://" + net.JoinHostPort(host, port)
}

func GetWasmUrl() string {
	return makeURL(config.WASMServerAddr)
}

func GetApiUrl() string {
	return makeURL(config.APIServerAddr)
}
