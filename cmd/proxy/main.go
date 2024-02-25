package main

import (
	"crypto/tls"
	"crypto/x509"
	"sc-proxy/pkg/proxy"
)

const (
	apiAddr   = ":8000"
	proxyAddr = ":8080"
	certsDir  = "certs"
)

func main() {
	ca, _ := tls.LoadX509KeyPair(certsDir+"/ca.crt", certsDir+"/ca.key")
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
	proxy.StartProxy(proxyAddr)
}
