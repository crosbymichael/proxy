package proxy

import (
	"io"
	"net"

	"github.com/BurntSushi/toml"
)

type Host struct {
	Backends        map[string]*Backend `toml:"backends"`
	Log             string              `toml:"log"`
	MaxListenErrors int                 `toml:"max_listen_errors"` // number of errors to accept before failing
	Docker          string              `toml:"docker"`
}

type Backend struct {
	Name             string
	Proto            string `toml:"proto"`
	ListenIP         net.IP `toml:"listen_ip"`
	ListenPort       int    `toml:"listen_port"`
	IP               net.IP `toml:"ip"`
	Port             int    `toml:"port"`
	MaxConcurrent    int    `toml:"max_concurrent"`
	ConnectionBuffer int    `toml:"connection_buffer"`
	Cert             string `toml:"cert"`
	Key              string `toml:"key"`
	CA               string `toml:"ca"`
	Container        string `toml:"container"`
}

func LoadConfig(r io.Reader) (*Host, error) {
	var config *Host
	if _, err := toml.DecodeReader(r, &config); err != nil {
		return nil, err
	}

	return config, nil
}
