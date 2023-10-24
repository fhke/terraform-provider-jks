package jks

type (
	// KeystoreBuilder provides a builder interface to generate JKS keystores.
	KeystoreBuilder struct {
		// keyPairs maps keypair aliases to keyPair.
		keyPairs map[string]keyPair
		// password is the keystore password.
		password string
	}

	// keyPair represents a certificate to add to the keystore.
	keyPair struct {
		// Private key in PEM format
		key []byte
		// Server cert in X.509 PEM format
		cert []byte
		// Optional slice of intermediate certs, in X.509 PEM format
		caCerts [][]byte
	}
)
