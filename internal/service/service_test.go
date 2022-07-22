package service

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"github.com/stretchr/testify/assert"
	"github.com/threetoes/peeper/internal/config"
	"io/ioutil"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
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

func TestRegisterAndServe(t *testing.T) {
	svc := New(":9090")
	svc.RegisterEndpoint(&config.Endpoint{
		LocalPath:    "/testpath",
		RemotePath:   "http://localhost:9091/forwarded",
		LocalMethod:  "GET",
		RemoteMethod: "POST",
	})
	testSvc := httptest.NewUnstartedServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if !assert.Equal(t, http.MethodPost, request.Method) {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte("bad method in request"))
			return
		}
		if !assert.Equal(t, "/forwarded", request.URL.Path) {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte("bad path in request"))
			return
		}

		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("test success I guess"))
	}))

	listener, err := net.Listen("tcp", "localhost:9091")
	if err != nil {
		t.Errorf("couldn't listen: %v", err)
		return
	}
	testSvc.Listener.Close()
	testSvc.Listener = listener

	go func() {
		svc.Start()
	}()
	testSvc.Start()
	defer testSvc.Close()

	// give the server a couple of seconds to come up
	time.Sleep(2 * time.Second)
	client := &http.Client{}
	resp, err := client.Get("http://localhost:9090/testpath")
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	assert.Equal(t, "test success I guess", string(body))
	svc.Stop()
}

func TestBasicAuthEndpoint(t *testing.T) {
	svc := New(":9090")
	svc.RegisterEndpoint(&config.Endpoint{
		LocalPath:    "/testpath",
		RemotePath:   "http://localhost:9091/forwarded",
		LocalMethod:  "GET",
		RemoteMethod: "POST",
		BasicAuth: &config.BasicAuthConfig{
			Username: "username1",
			Password: "passw0rd",
		},
	})
	testSvc := httptest.NewUnstartedServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		username, password, ok := request.BasicAuth()
		if !assert.Equal(t, "test value", request.Header.Get("test-header")) {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte("didn't forward headers"))
			return
		}
		if !assert.True(t, ok) {
			writer.WriteHeader(http.StatusUnauthorized)
			writer.Write([]byte("basic auth returned an error"))
			return
		}
		assert.Equal(t, "username1", username)
		assert.Equal(t, "passw0rd", password)
		if !assert.Equal(t, http.MethodPost, request.Method) {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte("bad method in request"))
			return
		}
		if !assert.Equal(t, "/forwarded", request.URL.Path) {
			writer.WriteHeader(http.StatusBadRequest)
			writer.Write([]byte("bad path in request"))
			return
		}

		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte("test success I guess"))
	}))

	listener, err := net.Listen("tcp", "localhost:9091")
	if err != nil {
		t.Errorf("couldn't listen: %v", err)
		return
	}
	testSvc.Listener.Close()
	testSvc.Listener = listener

	go func() {
		svc.Start()
	}()
	testSvc.Start()
	defer testSvc.Close()

	// give the server a couple of seconds to come up
	time.Sleep(2 * time.Second)
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://localhost:9090/testpath", nil)
	if !assert.NoError(t, err) {
		t.FailNow()
	}

	// Make sure headers are forwarded correctly
	req.Header.Set("test-header", "test value")

	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	assert.Equal(t, "test success I guess", string(body))
	svc.Stop()
}

func TestMTLS(t *testing.T) {
	testSvr := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	}))
	defer testSvr.Close()
	l, err := net.Listen("tcp", "localhost:9091")
	if !assert.NoError(t, err) {
		return
	}
	testSvr.Listener.Close()
	testSvr.Listener = l

	ca, _, caKey, err := GenerateCA()
	if !assert.NoError(t, err) {
		return
	}
	clientCert, _, err := GenerateAndSignCertificate(ca, caKey)
	if !assert.NoError(t, err) {
		return
	}
	svrCert, _, err := GenerateAndSignCertificate(ca, caKey)
	if !assert.NoError(t, err) {
		return
	}
	block, _ := pem.Decode(svrCert)
	if !assert.NotNil(t, block) {
		return
	}

	testSvr.TLS.ClientCAs.AddCert(ca)
	testSvr.TLS.Certificates = append(testSvr.TLS.Certificates, block)

	conf := config.AppOptions{
		Network: &config.NetworkConfig{
			BindInterface: "0.0.0.0",
			BindPort:      9090,
		},
		Endpoints: map[string]*config.Endpoint{
			"test": {
				LocalPath:    "/test",
				RemotePath:   "https://localhost:9091/test",
				LocalMethod:  "GET",
				RemoteMethod: "GET",
				HttpConfig: &config.HttpConfig{
					ClientCertificate: string(clientCert),
					ClientCAs:         testSvr.TLS.ClientCAs,
				},
			},
		},
	}
}
