package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"log"
	"sc-proxy/pkg/proxy"
	"sc-proxy/pkg/repository"
	"sc-proxy/pkg/service"
)

const (
	apiAddr   = ":8000"
	proxyAddr = ":8080"
	//certsDir  = "certs"
	certsDir = "../../certs"
)

func main() {
	ctx := context.Background()

	db, err := repository.NewPostgresDB(ctx, repository.PostgresConfig{
		Host:     "localhost",
		Port:     "5432",
		Username: "postgres",
		DBName:   "postgres",
		SSLMode:  "disable",
		Password: "1474",
	})
	if err != nil {
		log.Fatal("error when connecting to the database")
	}

	repos := repository.NewRepository(db)
	service := service.NewService(repos)

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
		Service: service,
	}
	proxy.StartProxy(proxyAddr)
}
