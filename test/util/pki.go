package util

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Create a new dummy self signed cert.
// Returns private key & cert, in PEM format.
func NewSelfSignedCertPEM(t *testing.T) ([]byte, []byte) {
	// generate private key
	priv, err := rsa.GenerateKey(rand.Reader, 4096)
	require.NoError(t, err, "It should generate key")

	// certificate template
	crtTmpl := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Foo org"},
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(time.Hour),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	// create cert in DER format
	derBytes, err := x509.CreateCertificate(rand.Reader, &crtTmpl, &crtTmpl, publicKey(priv), priv)
	require.NoError(t, err, "It should create certificate")

	return pem.EncodeToMemory(pemBlockForKey(priv)),
		pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
}

// create public key for private key.
func publicKey(priv interface{}) interface{} {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey
	case *ecdsa.PrivateKey:
		return &k.PublicKey
	default:
		return nil
	}
}

// generate pem block for private key.
func pemBlockForKey(priv interface{}) *pem.Block {
	switch k := priv.(type) {
	case *rsa.PrivateKey:
		return &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(k)}
	case *ecdsa.PrivateKey:
		b, err := x509.MarshalECPrivateKey(k)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to marshal ECDSA private key: %v", err)
			os.Exit(2)
		}
		return &pem.Block{Type: "EC PRIVATE KEY", Bytes: b}
	default:
		return nil
	}
}
