package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"math/big"
	"os"
	"testing"
	"time"
)

/* tests here are hard-coded can be tested with valid paths only, need to be automated or
script provided to create required files*/

func TestCreateCertPool(t *testing.T) {
	// Replace with a valid path or mock
	certPool := create_certPool("ca.crt", nil)
	if certPool == nil {
		t.Error("Expected certPool to be non-nil")
	}
}

func TestVerifyServerCert(t *testing.T) {
	// Provide valid file paths
	s := Server{
		ServerOpts: ServerOpts{ // mayBe use pointer to avoid copying in main code
			crtPath: "serverA.crt",
			keyPath: "serverA.key",
		},
	}
	certPool := create_certPool("ca.crt", nil)
	err := s.VerifyServerCert(certPool)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestCreatePeerCert(t *testing.T) {
	caKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	caTemplate := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"TestCA"},
		},
		IsCA:                  true,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}
	caDER, _ := x509.CreateCertificate(rand.Reader, caTemplate, caTemplate, &caKey.PublicKey, caKey)
	caCert, _ := x509.ParseCertificate(caDER)

	_, _ = CreatePeerCert(caKey, caCert, pkix.Name{CommonName: "Peer"}, "peerCertTest")

	// Verify files created
	if _, err := os.Stat("peerCertTest.crt"); err != nil {
		t.Error("Certificate file not created")
	}
	if _, err := os.Stat("peerCertTest.key"); err != nil {
		t.Error("Key file not created")
	}
}
