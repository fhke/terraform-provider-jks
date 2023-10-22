package jks

import (
	"fmt"

	"github.com/lwithers/minijks/jks"
)

// NewKeystoreBuilder creates a new KeyStoreBuilder
func NewKeystoreBuilder() *KeystoreBuilder {
	return &KeystoreBuilder{
		keyPairs: make(map[string]keyPair),
	}
}

/*
AddCert adds a certificate and private key to the key store.
If an alias is reused, this overwrites the previous cert.

Parameters:

	`alias`   - Alias for cert/key pair
	`cert`    - Certificate, in X.509 PEM format
	`key`     - Private key, in PEM format
	`caCerts` - Optional intermediate certificate authorities to add to keypair, in X.509 PEM format
*/
func (k *KeystoreBuilder) AddCert(alias string, cert []byte, key []byte, caCerts ...[]byte) {
	k.keyPairs[alias] = keyPair{
		key:     key,
		cert:    cert,
		caCerts: caCerts,
	}
}

// SetPassword sets the keystore password.
func (k *KeystoreBuilder) SetPassword(password string) {
	k.password = password
}

// Build constructs the keystore from the builder contents
func (k *KeystoreBuilder) Build() ([]byte, error) {
	// Validate builder contents
	if err := k.validate(); err != nil {
		return nil, fmt.Errorf("validation error: %w", err)
	}

	// Convert internal certs & keys to key pairs
	keyPairs, err := k.genKeyPairs()
	if err != nil {
		return nil, fmt.Errorf("error generating key pairs: %w", err)
	}

	// Create key store
	ks := &jks.Keystore{
		Keypairs: keyPairs,
	}

	// pack the keystore
	ksByt, err := ks.Pack(&jks.Options{
		Password: k.password,
	})
	if err != nil {
		return nil, fmt.Errorf("error converting keystore to JKS: %w", err)
	}

	return ksByt, nil
}

func (k *KeystoreBuilder) genKeyPairs() ([]*jks.Keypair, error) {
	kps := make([]*jks.Keypair, 0, len(k.keyPairs))

	// Add certs
	for alias, kp := range k.keyPairs {
		// Generate key pair
		jksKp, err := kp.toJKSKeypair(alias)
		if err != nil {
			return nil, fmt.Errorf("error generating key pair for certificate %q: %w", alias, err)
		}

		// add keypair to keystore
		kps = append(kps, jksKp)
	}

	return kps, nil
}

// validate validates the contents of the builder
func (k *KeystoreBuilder) validate() error {
	if k.password == "" {
		return ErrNoPassword
	}

	for alias, kp := range k.keyPairs {
		if alias == "" {
			return ErrInvalidAlias
		}
		if len(kp.cert) == 0 {
			return fmt.Errorf("certificate is empty for alias %q", alias)
		}
		if len(kp.key) == 0 {
			return fmt.Errorf("key is empty for alias %q", alias)
		}
		for i, caCert := range kp.caCerts {
			if len(caCert) == 0 {
				return fmt.Errorf("CA certificate %d for alias %q is empty", i, alias)
			}
		}
	}

	return nil
}
