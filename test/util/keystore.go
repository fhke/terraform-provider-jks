package util

import (
	"context"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/docker/docker/client"
	"github.com/stretchr/testify/require"
)

// ReadKeystore reads a JKS keystore to PEM format.
// Returns private key, and slice of certificates.
func ReadKeystore(ctx context.Context, t *testing.T, keyStore []byte, password string) ([]byte, [][]byte) {
	const (
		jksFile = "keystore.jks" // file containing JKS keystore
		p12File = "keystore.p12" // file containing pkcs#12 store
		pemFile = "certs.pem"    // file containing pem certs
	)

	// create docker client
	cli, err := client.NewClientWithOpts(client.FromEnv)
	require.NoError(t, err, "It should create Docker client")

	// get temp working dir
	tmpDir := t.TempDir()

	// clean files at end of function
	defer mustCleanFiles(t, tmpDir, jksFile, p12File, pemFile)

	// write keystore to file
	require.NoError(
		t,
		os.WriteFile(filepath.Join(tmpDir, jksFile), keyStore, 0640),
		"It should write keystore to file",
	)

	// convert keystore to pkcs#12
	runContainer(
		ctx,
		t,
		cli,
		envOr(EnvKeytoolImage, DefaultKeytoolImage),
		tmpDir,
		"keytool",
		"-importkeystore",
		"-srckeystore", jksFile,
		"-destkeystore", p12File,
		"-deststoretype", "pkcs12",
		"-srcstorepass", password,
		"-deststorepass", password,
		"-noprompt",
	)

	// convert pkcs#12 store to pem
	runContainer(
		ctx,
		t,
		cli,
		envOr(EnvOpenSSLImage, DefaultOpenSSLImage),
		tmpDir,
		"openssl",
		"pkcs12",
		"-nodes",
		"-in", p12File,
		"-out", pemFile,
		"-passin", fmt.Sprintf("pass:%s", password),
	)

	// read PEM file
	pemData, err := os.ReadFile(filepath.Join(tmpDir, pemFile))
	require.NoError(t, err, "It should read PEM file")

	// extract private key & certs
	var (
		privKey []byte
		certs   = make([][]byte, 0)
		pemBl   *pem.Block
	)
	for pemBl, pemData = pem.Decode(pemData); pemBl != nil; pemBl, pemData = pem.Decode(pemData) {
		switch typ := pemBl.Type; typ {
		case "PRIVATE KEY":
			privKey = pk8ToTraditional(ctx, t, cli, tmpDir, pem.EncodeToMemory(pemBl))
		case "CERTIFICATE":
			certs = append(certs, pem.EncodeToMemory(pemBl))
		default:
			t.Fatalf("Unknown PEM type: %s", typ)
		}
	}

	return privKey, certs
}

// convert pkcs8 format to traditional format.
func pk8ToTraditional(ctx context.Context, t *testing.T, cli *client.Client, wd string, data []byte) []byte {
	var (
		inFile  = "in.pem"
		outFile = "out.pem"
	)

	defer mustCleanFiles(t, wd, inFile, outFile)

	// write to file
	err := os.WriteFile(filepath.Join(wd, inFile), data, 0644)
	require.NoError(t, err, "It should write to file")

	runContainer(
		ctx,
		t,
		cli,
		envOr(EnvOpenSSLImage, DefaultOpenSSLImage),
		wd,
		"openssl",
		"pkcs8",
		"-inform", "pem",
		"-in", inFile,
		"-out", outFile,
		"-traditional",
		"-nocrypt",
	)

	// read file
	out, err := os.ReadFile(filepath.Join(wd, outFile))
	require.NoError(t, err, "It should read file")

	return out
}
