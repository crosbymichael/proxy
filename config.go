package proxy

import (
	"io"
	"net"

	"github.com/BurntSushi/toml"
)

type Host struct {
	Backends map[string]*Backend `toml:"backends"`
	// Where to log output to
	Log string `toml:"log"`
	// Number of errors to accept before failing
	MaxListenErrors int `toml:"max_listen_errors"`
	// Docker api endpoint to start containers on
	Docker string `toml:"docker"`
	// Rlimit to set
	Rlimit uint64 `toml:"rlimit"`
}

type Backend struct {
	// Name of the backend populated by the internal code
	Name string
	// Protocol of the proxy
	Proto string `toml:"proto"`
	// Ip that the proxy binds to
	ListenIP net.IP `toml:"listen_ip"`
	// Port that the proxy binds to
	ListenPort int `toml:"listen_port"`
	// Ip of the backend
	IP net.IP `toml:"ip"`
	// Port of the backend
	Port int `toml:"port"`
	// Maximum concurrent connections
	MaxConcurrent int `toml:"max_concurrent"`
	// How many connections to buffer
	ConnectionBuffer int `toml:"connection_buffer"`

	// TLS client side certs
	Cert string `toml:"cert"`
	Key  string `toml:"key"`
	CA   string `toml:"ca"`

	// Docker container to start for incoming connections
	Container string `toml:"container"`
	// Seconds to stop a container on inactivity
	ContainerStopTimeout int `toml:"container_stop_timeout"`
}

func LoadConfig(r io.Reader) (*Host, error) {
	var config *Host
	if _, err := toml.DecodeReader(r, &config); err != nil {
		return nil, err
	}

	return config, nil
}
