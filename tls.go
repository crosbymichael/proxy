package proxy

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
)

func createTLSConfig(cert, key, ca string) (*tls.Config, error) {
	loadedCertificate, err := tls.LoadX509KeyPair(cert, key)
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		NextProtos:   []string{"http/1.1"},
		Certificates: []tls.Certificate{loadedCertificate},
	}

	if ca != "" {
		pool := x509.NewCertPool()

		data, err := ioutil.ReadFile(ca)
		if err != nil {
			return nil, err
		}

		if !pool.AppendCertsFromPEM(data) {
			return nil, fmt.Errorf("unable to load ca from file %s", ca)
		}

		config.ClientAuth = tls.RequireAndVerifyClientCert
		config.ClientCAs = pool
	}

	return config, nil
}
