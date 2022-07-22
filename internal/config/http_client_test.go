package config

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/big"
	"net"
	"net/http"
	"testing"
	"time"
)

const CompanyName = "Company, INC."
const Country = "US"
const Province = ""
const Locality = "San Francisco"
const StreetAddress = "Golden Gate Bridge"
const PostalCode = "94016"

// GenerateCA returns the CA cert, the CA key and any errors
func GenerateCA() (*x509.Certificate, []byte, []byte, error) {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Organization:  []string{CompanyName},
			Country:       []string{Country},
			Province:      []string{Province},
			Locality:      []string{Locality},
			StreetAddress: []string{StreetAddress},
			PostalCode:    []string{PostalCode},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, nil, err
	}
	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return nil, nil, nil, err
	}
	caPEM := new(bytes.Buffer)
	err = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	if err != nil {
		return nil, nil, nil, err
	}

	caPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	})
	if err != nil {
		return nil, nil, nil, err
	}
	return ca, caPEM.Bytes(), caPrivKeyPEM.Bytes(), nil
}

func GenerateAndSignCertificate(ca *x509.Certificate, caPrivKey []byte) ([]byte, []byte, error) {
	// I hate that I have to duplicate this, but oh well
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(2019),
		Subject: pkix.Name{
			Organization:  []string{CompanyName},
			Country:       []string{Country},
			Province:      []string{Province},
			Locality:      []string{Locality},
			StreetAddress: []string{StreetAddress},
			PostalCode:    []string{PostalCode},
		},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	block, _ := pem.Decode(caPrivKey)
	if block == nil {
		return nil, nil, fmt.Errorf("could not decode CA private key")
	}
	caKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, nil, err
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, err
	}

	certPEM := new(bytes.Buffer)
	err = pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		return nil, nil, err
	}

	certPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})
	if err != nil {
		return nil, nil, err
	}

	return certPEM.Bytes(), certPrivKeyPEM.Bytes(), nil
}

func TestClient(t *testing.T) {
	t.Run("no configs", func(t *testing.T) {
		httpConfig := &HttpConfig{}
		client, err := httpConfig.Client()
		assert.NoError(t, err)
		assert.NotNil(t, client)
	})
	t.Run("proxy set", func(t *testing.T) {
		httpConfig := &HttpConfig{
			ProxyServer: "http://localhost:8080",
		}
		client, err := httpConfig.Client()
		assert.NoError(t, err)
		if !assert.NotNil(t, client) {
			t.FailNow()
		}
		transport, ok := client.Transport.(*http.Transport)
		if !assert.True(t, ok) {
			t.FailNow()
		}
		assert.NotNil(t, transport.Proxy)
	})

	t.Run("mtls set", func(t *testing.T) {
		_, caPem, caKey, err := GenerateCA()
		if !assert.NoError(t, err) {
			return
		}
		cert := &x509.Certificate{
			DNSNames:       []string{"localhost"},
			EmailAddresses: []string{"admin@test.com"},
			IPAddresses:    []net.IP{net.IPv4(byte(127), byte(0), byte(0), byte(1))},
		}
		signedCert, _, err := GenerateAndSignCertificate(cert, caKey)
		if !assert.NoError(t, err) {
			return
		}
		httpConfig := &HttpConfig{
			ClientCAs:         string(caPem),
			ClientCertificate: string(signedCert),
		}
		client, err := httpConfig.Client()
		assert.NoError(t, err)
		if !assert.NotNil(t, client) {
			t.FailNow()
		}
		transport, ok := client.Transport.(*http.Transport)
		if !assert.True(t, ok) {
			t.FailNow()
		}
		assert.NotNil(t, transport.TLSClientConfig)
		assert.NotNil(t, transport.TLSClientConfig.ClientCAs)
		assert.NotNil(t, transport.TLSClientConfig.Certificates)
	})
}
