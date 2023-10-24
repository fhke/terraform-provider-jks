package jks

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
)

// parse PEM data to a certificate.
func parseCertPEM(data []byte) (*x509.Certificate, error) {
	bl, err := decodePEM(data)
	if err != nil {
		return nil, err
	}
	return x509.ParseCertificate(bl.Bytes)
}

// decode PEM data to a pem block.
func decodePEM(data []byte) (*pem.Block, error) {
	bl, _ := pem.Decode(data)
	if bl == nil {
		return nil, errors.New("error decoding data from PEM")
	}
	return bl, nil
}
