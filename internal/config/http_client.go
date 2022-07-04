package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/threetoes/peeper/internal/util"
	"net/http"
	"net/url"
	"time"
)

type HttpConfig struct {
	ClientCertificate string `toml:"client_certificate"`
	ClientCAs         string `toml:"client_cas"`
	ProxyServer       string `toml:"proxy_server"`
	Timeout           int64  `toml:"timeout"`
}

func (h *HttpConfig) Client() (*http.Client, error) {
	transport := &http.Transport{}

	if h.ProxyServer != "" {
		proxyUrl, err := url.Parse(h.ProxyServer)
		if err != nil {
			return nil, err
		}
		transport.Proxy = http.ProxyURL(proxyUrl)
	}

	transport.TLSClientConfig = &tls.Config{}

	if h.ClientCertificate != "" {
		cert, err := util.LoadCertificate(h.ClientCertificate)
		if err != nil {
			return nil, err
		}
		transport.TLSClientConfig.Certificates = []tls.Certificate{cert}
	}

	if h.ClientCAs != "" {
		certPool := x509.NewCertPool()
		certBytes, err := util.GetCertBytes(h.ClientCAs)
		if err != nil {
			return nil, err
		}
		if !certPool.AppendCertsFromPEM(certBytes) {
			return nil, fmt.Errorf("could not read PEM format certificate")
		}

		transport.TLSClientConfig.ClientCAs = certPool
	}

	return &http.Client{
		Transport: transport,
		Timeout:   time.Duration(h.Timeout) * time.Second,
	}, nil
}
