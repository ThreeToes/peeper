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
	t.Run("test long cert", func(t *testing.T) {
		// Generated cert that caused the function to flip out in testing
		_, err := LoadCertificate(`-----BEGIN CERTIFICATE-----
        	            	MIIFxzCCA6+gAwIBAgICB+MwDQYJKoZIhvcNAQELBQAwdTELMAkGA1UEBhMCVVMx
        	            	CTAHBgNVBAgTADEWMBQGA1UEBxMNU2FuIEZyYW5jaXNjbzEbMBkGA1UECRMSR29s
        	            	ZGVuIEdhdGUgQnJpZGdlMQ4wDAYDVQQREwU5NDAxNjEWMBQGA1UEChMNQ29tcGFu
        	            	eSwgSU5DLjAeFw0yMjA3MjIxMzMwMTdaFw0zMjA3MjIxMzMwMTdaMHUxCzAJBgNV
        	            	BAYTAlVTMQkwBwYDVQQIEwAxFjAUBgNVBAcTDVNhbiBGcmFuY2lzY28xGzAZBgNV
        	            	BAkTEkdvbGRlbiBHYXRlIEJyaWRnZTEOMAwGA1UEERMFOTQwMTYxFjAUBgNVBAoT
        	            	DUNvbXBhbnksIElOQy4wggIiMA0GCSqGSIb3DQEBAQUAA4ICDwAwggIKAoICAQC9
        	            	fqEB/JiAaRZsuiuQ2FV8ir7HKDHKo4eTrZ1P8o8ZFUYnQwY0HhuBVqt836sfcVq0
        	            	slDMhmp2n8E9oIAtrrJUhoPYDnEY1cdqKsvlyi2t6kM3gh0Y3GCx7Hm5Y66AeIJq
        	            	YyMmZ4buFs8KQbs7LMwMNr2ndmQ+JOc45+WLbOP3C+GJ6qDMLGGQWqXzS7Us4eVI
        	            	2ocYBOtiU59zS6dDqHKPje5qPw8m+sRhNrwNBfqw8anta0WO3auJxvkrluYWWG4s
        	            	REUf8eHI/WeeAfoOPwLR9p8D04nKV/X7EsXUxMnooE8/KREi+kxu6TOcqTISjeX2
        	            	3aQwli7jURG5+IKzcCV1iFd/tTuoWmiHwvOjofeem5j/cTQHXEJv1u+ikQjdMlNI
        	            	FfcITvbiYF9Zibwb8TiwTUG9mUKrGVStZgRN7InC/ExhV3UP26eD7KkfEo4ovGEi
        	            	g3wzYO0DFi4ElduXodCxuU0VLEZDa7Y85feP1F9ED3R0cIBDBgQgjKhsfAMgjCrY
        	            	6gIsj4PuMQus/p0LmIetv3tA2OlmDJXeiqyoTfD/4M57UK4QfZGWfB6i2OlGmh8t
        	            	xaoADJFaM5yHcK5Q5S084O/ieiPGVpIwPiUTpiqn22tQeFdcbwjWC4mpNxBaITS3
        	            	NSU6eVJd0SGv8aNrYXGHRm2rCpi9woO1rF2W5CO8gQIDAQABo2EwXzAOBgNVHQ8B
        	            	Af8EBAMCAoQwHQYDVR0lBBYwFAYIKwYBBQUHAwIGCCsGAQUFBwMBMA8GA1UdEwEB
        	            	/wQFMAMBAf8wHQYDVR0OBBYEFP5+AVabVN09l4skrp+OWMusuD9GMA0GCSqGSIb3
        	            	DQEBCwUAA4ICAQB7/w0xijapvetrtb1Xy9p42PrGbUIWCpo170o89tXh3KRKtR2F
        	            	B+MsBRTr2IHKvStw0KKoo0vLsTX+npc22Cd25HnNZQ/CeJPHVnka76hkaBg0+yIH
        	            	jfMtO7zLJAsnxeC+wDSLa0CifIPVN/marbdDYZxOfgyND/k0wvg+5K93YGuuyyOE
        	            	Nu0XMsLFGdSvbTyepQzv2QNn1bMisxI9NEOBpIHgyieFSHErur+ZheSfIB+1syIU
        	            	JC1uQ2/0Xr6b+v2aqmCoMuNGHxoHylosc15fOwlN3Lt8bZiSm0i80KkyBslo5qKy
        	            	St4OZ25epkYGbn9c/XC+bb7vrD89eL4gprALUycVMlQlY9Cd7gs1S0eGr9YeScAg
        	            	MyPBlP9B6i1yn41WUEjcITno81A7eEoLOWdELwiG/lQs061GrSU6+kCwzdnR496t
        	            	0TekmatQ9DZCQ/fKWkz+Tey6Hq/vJjJI5AcFDczof3yMJO4Uav0hqEZf0Y43sIxq
        	            	xV0HQZu3r3ThzXySVJQ0f0hFRLJ7S/qyRwgkWt5jK8Ak5dtzA7C8PXsXvCqpRsvv
        	            	uQunecDO4by34gffzlG7jixuImTAqa/YJgZYFn1xw4jClCgtsO48h+1+iI7dyOAM
        	            	JmZyhoddRtrU3siS4y1PaIY/kvp6P/Xw6SGeXo5NVsBXOHP4jf62E1/qlg==
        	            	-----END CERTIFICATE-----`)
		assert.NoError(t, err)
	})
}
