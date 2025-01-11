package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
	"encoding/pem"
	"log"
	"math/big"
	"os"
	"time"
)

// Create a certificate pool from the certificate authority
// pass nil in caCertPool to create a new one
func create_certPool(caCertPath string, caCertPool *x509.CertPool) *x509.CertPool {
	// Load CA cert
	caCert, err := os.ReadFile(caCertPath)
	if err != nil {
		log.Printf("Error loading certificate: %s", err)
	}

	if caCertPool == nil {
		caCertPool = x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
	} else {
		caCertPool.AppendCertsFromPEM(caCert)
	}

	return caCertPool
}

// Verify the server certificate
func (s *Server) VerifyServerCert(caCertPool *x509.CertPool) error {
	cert, err := tls.LoadX509KeyPair(s.crtPath, s.keyPath)
	if err != nil {
		log.Printf("Error loading certificate: %s", err)
		return err
	}
	opts := x509.VerifyOptions{
		Roots: caCertPool,
	}
	for _, cert := range cert.Certificate {
		x509Cert, err := x509.ParseCertificate(cert)
		if err != nil {
			log.Printf("Error parsing certificate: %s", err)
			return err
		}
		if _, err := x509Cert.Verify(opts); err != nil {
			log.Printf("Error verifying certificate: %s", err)
			return err
		}
	}
	return nil
}

// creates a certificate for peer/server of system
//
//	detail: pkix.Name{
//		Country:      []string{"US"},
//		Province:     []string{"California"},
//		Locality:     []string{"YourCompany"},
//
//	}
func (s *Server) CreateServerCert(caKey *rsa.PrivateKey, caCert *x509.Certificate, detail pkix.Name) (*big.Int, pkix.Name) {
	serverKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Fatalf("Failed to generate server private key: %v", err)
	}
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		log.Fatalf("Failed to generate serial number: %v", err)
	}
	hexserialNumber := hex.EncodeToString(serialNumber.Bytes())
	s.ServerOpts.crtPath = hexserialNumber + ".crt"
	s.ServerOpts.keyPath = hexserialNumber + ".key"
	// Create server cert template
	serverCertTemplate := x509.Certificate{
		SerialNumber: serialNumber,
		Issuer:       caCert.Subject,
		Subject:      detail,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0), // valid for 1 year, TODO: What about after expiration?
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}

	serverCertDER, err := x509.CreateCertificate(rand.Reader, &serverCertTemplate, caCert, &serverKey.PublicKey, caKey)
	if err != nil {
		log.Fatalf("Failed to create server certificate: %v", err)
	}
	serverKeyFile, err := os.Create(s.keyPath)
	if err != nil {
		log.Fatalf("Failed to open server key file: %v", err)
	}
	defer serverKeyFile.Close()

	pem.Encode(serverKeyFile, &pem.Block{Type: "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverKey)})
	serverCertFile, err := os.Create(s.crtPath)
	if err != nil {
		log.Fatalf("Failed to open server cert file: %v", err)
	}
	defer serverCertFile.Close()

	pem.Encode(serverCertFile, &pem.Block{Type: "CERTIFICATE", Bytes: serverCertDER})

	return serverCertTemplate.SerialNumber, serverCertTemplate.Subject
}

// Create a CA certificate and save files at provided paths
//
//	detail: pkix.Name{
//		Country:      []string{"US"},
//		Province:     []string{"California"},
//		Locality:     []string{"YourCompany"},
//		Organization: []string{"YourCA"},
//		CommonName:   "YourCA",
//	}
//
// Still considering, is it need to be automated or manual -> creation of CA certificate
func createCACert(caKeyPath, caCertPath string, detail pkix.Name) {
	caKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		log.Printf("Failed to generate CA private key: %v", err)
	}
	caCertTemplate := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               detail,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0), // Valid for 1 year
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}

	// Create CA certificate
	caCertDER, err := x509.CreateCertificate(rand.Reader, caCertTemplate, caCertTemplate, &caKey.PublicKey, caKey)
	if err != nil {
		log.Printf("Failed to create CA certificate: %v", err)
	}
	caKeyFile, err := os.Create(caKeyPath)
	if err != nil {
		log.Printf("Failed to open CA key file: %v", err)
	}
	defer caKeyFile.Close()
	pem.Encode(caKeyFile, &pem.Block{Type: "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caKey)})
	caCertFile, err := os.Create(caCertPath)
	if err != nil {
		log.Printf("Failed to open CA cert file: %v", err)
	}
	defer caCertFile.Close()

	pem.Encode(caCertFile, &pem.Block{Type: "CERTIFICATE", Bytes: caCertDER})

}
