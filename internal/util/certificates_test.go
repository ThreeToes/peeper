package util

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"github.com/stretchr/testify/assert"
	"math/big"
	"net"
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

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, caPrivKey)
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

func TestLoadCertificate(t *testing.T) {
	t.Run("load from bytes", func(t *testing.T) {
		_, caPem, _, err := GenerateCA()
		if !assert.NoError(t, err, "could not generate cert") {
			t.FailNow()
		}
		tlsCert, err := LoadCertificate(string(caPem))
		if !assert.NoError(t, err, "failed to load cert from bytes: %v", err) {
			t.FailNow()
		}
		cert, err := x509.ParseCertificate(tlsCert.Certificate[0])
		if !assert.NoError(t, err, "failed to parse cert: %v", err) {
			t.FailNow()
		}

		assert.Len(t, cert.Subject.Organization, 1)
		assert.Equal(t, CompanyName, cert.Subject.Organization[0])
	})

}
