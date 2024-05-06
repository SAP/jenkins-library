//go:build unit
// +build unit

package http

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io"
	"log"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func GenerateSelfSignedCertificate(usages []x509.ExtKeyUsage) (pemKey, pemCert []byte) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		log.Fatalf("Failed to generate private key: %v", err)
	}

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("Failed to generate serial number: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"My Corp"},
		},
		DNSNames:  []string{"localhost"},
		NotBefore: time.Now(),
		NotAfter:  time.Now().Add(3 * time.Hour),

		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           usages,
		BasicConstraintsValid: true,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		log.Fatalf("Failed to create certificate: %v", err)
	}

	pemCert = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})
	if pemCert == nil {
		log.Fatal("Failed to encode certificate to PEM")
	}

	privBytes, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		log.Fatalf("Unable to marshal private key: %v", err)
	}
	pemKey = pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: privBytes})
	if pemKey == nil {
		log.Fatal("Failed to encode key to PEM")
	}

	return pemKey, pemCert
}

func GenerateSelfSignedServerAuthCertificate() (pemKey, pemCert []byte) {
	return GenerateSelfSignedCertificate([]x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth})
}

func GenerateSelfSignedClientAuthCertificate() (pemKey, pemCert []byte) {
	return GenerateSelfSignedCertificate([]x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth})
}

func TestCertificateLogon(t *testing.T) {
	testOkayString := "Okidoki"

	server := httptest.NewUnstartedServer(http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		rw.Write([]byte(testOkayString))
	}))

	clientPemKey, clientPemCert := GenerateSelfSignedClientAuthCertificate()

	// server
	clientCertPool := x509.NewCertPool()
	clientCertPool.AppendCertsFromPEM(clientPemCert)

	tlsConfig := tls.Config{
		MinVersion:               tls.VersionTLS13,
		PreferServerCipherSuites: true,
		ClientCAs:                clientCertPool,
		ClientAuth:               tls.RequireAndVerifyClientCert,
	}

	server.TLS = &tlsConfig
	server.StartTLS()
	defer server.Close()

	// client
	tlsKeyPair, err := tls.X509KeyPair(clientPemCert, clientPemKey)
	if err != nil {
		log.Fatal("Failed to create clients tls key pair")
	}

	t.Run("Success - Login with certificate", func(t *testing.T) {
		c := Client{}
		c.SetOptions(ClientOptions{
			TransportSkipVerification: true,
			MaxRetries:                1,
			Certificates:              []tls.Certificate{tlsKeyPair},
		})

		response, err := c.SendRequest("GET", server.URL, nil, nil, nil)
		assert.NoError(t, err, "Error occurred but none expected")
		content, err := io.ReadAll(response.Body)
		assert.Equal(t, testOkayString, string(content), "Returned content incorrect")
		response.Body.Close()
	})

	t.Run("Failure - Login without certificate", func(t *testing.T) {
		c := Client{}
		c.SetOptions(ClientOptions{
			TransportSkipVerification: true,
			MaxRetries:                1,
		})

		_, err := c.SendRequest("GET", server.URL, nil, nil, nil)
		assert.ErrorContains(t, err, "certificate required")
	})

	t.Run("Failure - Login with wrong certificate", func(t *testing.T) {
		otherClientPemKey, otherClientPemCert := GenerateSelfSignedClientAuthCertificate()

		otherTlsKeyPair, err := tls.X509KeyPair(otherClientPemCert, otherClientPemKey)
		if err != nil {
			log.Fatal("Failed to create clients tls key pair")
		}

		c := Client{}
		c.SetOptions(ClientOptions{
			TransportSkipVerification: true,
			MaxRetries:                1,
			Certificates:              []tls.Certificate{otherTlsKeyPair},
		})

		_, err = c.SendRequest("GET", server.URL, nil, nil, nil)
		assert.ErrorContains(t, err, "unknown certificate authority")
	})

	t.Run("SanityCheck", func(t *testing.T) {
		client := &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					// RootCAs:      certPool,
					InsecureSkipVerify: true,
					Certificates:       []tls.Certificate{tlsKeyPair},
				},
			},
		}

		response, err := client.Get(server.URL)
		assert.NoError(t, err, "Error occurred but none expected")
		content, err := io.ReadAll(response.Body)
		assert.Equal(t, testOkayString, string(content), "Returned content incorrect")
		response.Body.Close()
	})
}
