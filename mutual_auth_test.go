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
	// certData, err := os.ReadFile("ca.crt")
	// if err != nil {
	// 	t.Fatalf("Failed to read CA certificate: %v", err)
	// }
	// cert, err := x509.ParseCertificate(certData)
	// if err != nil {
	// 	t.Fatalf("Failed to parse CA certificate: %v", err)
	// }
	// log.Printf("CA Certificate Subject: %v", cert.Subject)
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

func TestCreateServerCert(t *testing.T) {
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
	s := Server{
		ServerOpts: ServerOpts{},
	}
	_, _ = s.CreateServerCert(caKey, caCert, pkix.Name{CommonName: "Peer"})

	// Verify files created
	if _, err := os.Stat(s.crtPath); err != nil {
		t.Error("Certificate file not created")
	}
	if _, err := os.Stat(s.keyPath); err != nil {
		t.Error("Key file not created")
	}
	os.Remove(s.crtPath)
	os.Remove(s.keyPath)
}

func TestCreateCACert(t *testing.T) {
	caKeyPath := "testca.key"
	caCertPath := "testca.crt"
	detail := pkix.Name{
		Country:      []string{"US"},
		Province:     []string{"California"},
		Locality:     []string{"TestingLocation"},
		Organization: []string{"TestOrg"},
		CommonName:   "TestCA",
	}
	createCACert(caKeyPath, caCertPath, detail)
	if _, err := os.Stat(caKeyPath); err != nil {
		t.Errorf("Expected CA key file to exist, got error: %v", err)
	}
	if _, err := os.Stat(caCertPath); err != nil {
		t.Errorf("Expected CA cert file to exist, got error: %v", err)
	}
	os.Remove(caKeyPath)
	os.Remove(caCertPath)
}
