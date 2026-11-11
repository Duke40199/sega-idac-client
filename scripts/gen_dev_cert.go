//go:build ignore

// gen_dev_cert.go generates a self-signed TLS certificate for LOCAL DEVELOPMENT
// of the HTTP gateway. It is NOT for production use.
//
// Usage:
//   go run scripts/gen_dev_cert.go               # writes certs/dev.crt, certs/dev.key
//   go run scripts/gen_dev_cert.go -out certs    # custom output dir
//
// The cert is valid for localhost, 127.0.0.1 and ::1 for 825 days.
package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"flag"
	"log"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

func main() {
	outDir := flag.String("out", "certs", "output directory for dev.crt / dev.key")
	flag.Parse()

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		log.Fatalf("creating %s: %v", *outDir, err)
	}

	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("generating key: %v", err)
	}

	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		log.Fatalf("generating serial: %v", err)
	}

	now := time.Now()
	tmpl := x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{Organization: []string{"sega-idac-client dev"}, CommonName: "localhost"},
		NotBefore:             now.Add(-time.Hour),
		NotAfter:              now.Add(825 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
	}

	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &priv.PublicKey, priv)
	if err != nil {
		log.Fatalf("creating certificate: %v", err)
	}

	certPath := filepath.Join(*outDir, "dev.crt")
	certOut, err := os.Create(certPath)
	if err != nil {
		log.Fatalf("opening %s: %v", certPath, err)
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: der}); err != nil {
		log.Fatalf("writing cert: %v", err)
	}
	certOut.Close()

	keyBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		log.Fatalf("marshaling key: %v", err)
	}
	keyPath := filepath.Join(*outDir, "dev.key")
	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		log.Fatalf("opening %s: %v", keyPath, err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: keyBytes}); err != nil {
		log.Fatalf("writing key: %v", err)
	}
	keyOut.Close()

	log.Printf("✅ wrote %s and %s (self-signed, localhost/127.0.0.1/::1, dev only)", certPath, keyPath)
}
