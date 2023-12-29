package config

import (
	"errors"
	"net"
	"os"

	"github.com/pelletier/go-toml/v2"
)

const defaultConfig = `
wasmServerAddr 	= ":5000"
apiServerAddr 	= ":3000"
`

// Config holds the global configuration which is READONLY ofcourse.
var config Config

type Config struct {
	APIServerAddr  string
	WASMServerAddr string
}

func Parse(path string) error {
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		if err := os.WriteFile("ffaas.toml", []byte(defaultConfig), os.ModePerm); err != nil {
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

func makeURL(address string) string {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return ""
	}

	if host == "" {
		host = "0.0.0.0"
	}

	return "http://" + net.JoinHostPort(host, port)
}

func GetWasmUrl() string {
	return makeURL(config.WASMServerAddr)
}

func GetApiUrl() string {
	return makeURL(config.APIServerAddr)
}
