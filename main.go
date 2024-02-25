package main

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"net/http"
	"os"
	"sc-proxy/proxy"
)

const (
	addr     = ":8080"
	certsDir = "certs"
)

var hostname, _ = os.Hostname()

func genCA() (cert tls.Certificate, err error) {
	certPEM, keyPEM, err := proxy.GenCA(hostname)
	if err != nil {
		return tls.Certificate{}, err
	}
	cert, err = tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return tls.Certificate{}, err
	}
	cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
	return cert, err
}

func main() {
	ca, _ := tls.LoadX509KeyPair("certs/ca.crt", "certs/ca.key")
	ca.Leaf, _ = x509.ParseCertificate(ca.Certificate[0])

	proxy := &proxy.Proxy{
		CA: &ca,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		TLSServerConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	log.Fatal(http.ListenAndServe(":8080", proxy))
}
