package jks_test

import (
	"context"
	"testing"

	"github.com/fhke/terraform-provider-jks/jks"
	"github.com/fhke/terraform-provider-jks/test/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test keystore with one server cert & two intermediate certs.
func TestKeystore(t *testing.T) {
	password := "test1234"
	origKey, origCrt := util.NewSelfSignedCertPEM(t)
	_, origInterCrt1 := util.NewSelfSignedCertPEM(t)
	_, origInterCrt2 := util.NewSelfSignedCertPEM(t)

	ksBuilder := jks.NewKeystoreBuilder()
	ksBuilder.AddCert(
		"cert",
		origCrt,
		origKey,
		origInterCrt1,
		origInterCrt2,
	)
	ksBuilder.SetPassword(password)
	keyStore, err := ksBuilder.Build()
	require.NoError(t, err, "It should build keystore")

	newKey, newCerts := util.ReadKeystore(context.TODO(), t, keyStore, password)
	require.NotEmpty(t, newKey, "Private key should not be empty")
	require.Len(t, newCerts, 3, "JKS should contain one cert")
	assert.Equal(t, origKey, newKey, "Private key should match")
	assert.Equal(t, origCrt, newCerts[0], "Server cert should match")
	assert.Equal(t, origInterCrt1, newCerts[1], "Intermediate cert 1 should match")
	assert.Equal(t, origInterCrt2, newCerts[2], "Intermediate cert 2 should match")
}

// Test keystore with one server cert & no intermediate certs.
func TestKeystoreSingle(t *testing.T) {
	password := "test4321"
	origKey, origCrt := util.NewSelfSignedCertPEM(t)

	ksBuilder := jks.NewKeystoreBuilder()
	ksBuilder.AddCert(
		"cert",
		origCrt,
		origKey,
	)
	ksBuilder.SetPassword(password)
	keyStore, err := ksBuilder.Build()
	require.NoError(t, err, "It should build keystore")

	newKey, newCerts := util.ReadKeystore(context.TODO(), t, keyStore, password)
	require.NotEmpty(t, newKey, "Private key should not be empty")
	require.Len(t, newCerts, 1, "JKS should contain one cert")
	assert.Equal(t, origKey, newKey, "Private key should match")
	assert.Equal(t, origCrt, newCerts[0], "Server cert should match")
}
