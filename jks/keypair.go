package jks

import (
	"crypto/x509"
	"fmt"

	"github.com/lwithers/minijks/jks"
)

// certChain generates a chain of certificates, starting with the server cert
func (k keyPair) certChain() ([]*x509.Certificate, error) {
	// prepend server cert to CA certs
	pemCerts := append([][]byte{k.cert}, k.caCerts...)

	certs := make([]*x509.Certificate, len(pemCerts))

	for i, pemCert := range pemCerts {
		// parse cert & add to slice
		crt, err := parseCertPEM(pemCert)
		if err != nil {
			return nil, fmt.Errorf("error parsing certificate %d: %w", i, err)
		}
		certs[i] = crt
	}

	return certs, nil
}

// privKey decodes the private key to a private key format
func (k keyPair) privKey() (any, error) {
	// parse key from PEM
	pemKey, err := decodePEM(k.key)
	if err != nil {
		return nil, err
	}

	// decode key
	switch typ := pemKey.Type; typ {
	case "PRIVATE KEY":
		return x509.ParsePKCS8PrivateKey(pemKey.Bytes)
	case "RSA PRIVATE KEY":
		return x509.ParsePKCS1PrivateKey(pemKey.Bytes)
	case "EC PRIVATE KEY":
		return x509.ParseECPrivateKey(pemKey.Bytes)
	default:
		return nil, fmt.Errorf("unknown key type: %s", typ)
	}
}

// toJKSKeypair converts keyPair to a *jks.Keypair
func (k keyPair) toJKSKeypair(alias string) (*jks.Keypair, error) {
	// get private key
	privKey, err := k.privKey()
	if err != nil {
		return nil, fmt.Errorf("error parsing private key: %w", err)
	}

	// get cert chain
	certs, err := k.certChain()
	if err != nil {
		return nil, fmt.Errorf("error parsing certificate chain: %w", err)
	}

	// Create keypair
	jksKp := &jks.Keypair{
		Alias:      alias,
		PrivateKey: privKey,
		CertChain:  make([]*jks.KeypairCert, len(certs)),
	}

	// add certs to keypair
	for i, cert := range certs {
		jksKp.CertChain[i] = &jks.KeypairCert{
			Cert: cert,
		}
	}

	return jksKp, nil
}
