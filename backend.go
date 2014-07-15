package proxy

import "net"

type Backend struct {
	Name             string `json:"name,omitempty"`
	Proto            string `json:"proto,omitempty"`
	BindIP           net.IP `json:"bind_ip,omitempty"`
	BindPort         int    `json:"bind_port,omitempty"`
	IP               net.IP `json:"backend_ip,omitempty"`
	Port             int    `json:"backend_port,omitempty"`
	MaxConcurrent    int    `json:"max_concurrent,omitempty"`
	ConnectionBuffer int    `json:"connection_buffer,omitempty"`

	// TLS client side certs
	Cert string `json:"cert,omitempty"`
	Key  string `json:"key,omitempty"`
	CA   string `json:"ca,omitempty"`

	// Docker container to start for incoming connections
	Container string `json:"container,omitempty"`
	// Seconds to stop a container on inactivity
	ContainerStopTimeout int `json:"container_stop_timeout,omitempty"`
}
