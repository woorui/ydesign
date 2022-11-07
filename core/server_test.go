package internal

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"testing"
	"time"
)

const testAddr = "localhost:19999"

var tlsConf = &tls.Config{
	InsecureSkipVerify: true,
	NextProtos:         []string{"yomo-test"},
}

var serverTLSConf = generateTLSConfig()

func TestServer(t *testing.T) {
	ctx := context.Background()

	mux := NewYomoMux()

	server := NewServer(mux, serverTLSConf, nil)

	go func() {
		server.ListenAndServe(ctx, testAddr)
	}()

	time.AfterFunc(time.Second, func() {
		server.Close()
	})

	client := NewClient(ctx, testAddr, tlsConf, nil)

	fmt.Println(client)
}

func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"yomo-test"},
	}
}
