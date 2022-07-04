package util

import (
	"crypto/tls"
	"encoding/pem"
	"io/ioutil"
	"os"
)

func LoadCertificate(certString string) (tls.Certificate, error) {
	var certPEMBlock []byte
	certPEMBlock, err := GetCertBytes(certString, certPEMBlock)
	if err != nil {
		return tls.Certificate{}, err
	}
	var cert tls.Certificate
	var certDERBlock *pem.Block
	for {
		certDERBlock, certPEMBlock = pem.Decode(certPEMBlock)
		if certDERBlock == nil {
			break
		}
		if certDERBlock.Type == "CERTIFICATE" {
			cert.Certificate = append(cert.Certificate, certDERBlock.Bytes)
		}
	}
	return cert, nil
}

func GetCertBytes(certString string) ([]byte, error) {
	if _, err := os.Stat(certString); err != nil {
		certPEMBlock, err := ioutil.ReadFile(certString)
		return certPEMBlock, err
	} else {
		return []byte(certString), nil
	}
	return []byte(certString), nil
}
