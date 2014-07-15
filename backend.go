package proxy

import "net"

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
