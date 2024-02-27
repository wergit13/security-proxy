package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"log"
	"os"
	"os/signal"
	"sc-proxy/pkg/api"
	"sc-proxy/pkg/api/handlers"
	"sc-proxy/pkg/proxy"
	"sc-proxy/pkg/repository"
	"sc-proxy/pkg/service"
	"syscall"
)

const (
	apiAddr   = ":8000"
	proxyAddr = ":8080"
	certsDir  = "certs"
	//certsDir = "../../certs"
)

func main() {
	ctx := context.Background()

	db, err := repository.NewPostgresDB(ctx, repository.PostgresConfig{
		Host:     "db_proxy",
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
	handlers := handlers.NewHandler(service)

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

	go func() {
		err = proxy.StartProxy(proxyAddr)
		if err != nil {
			log.Fatal("error when starting the proxy server")
		}
	}()

	srv := new(api.Server)

	go func() {
		if err = srv.Serve("8000", handlers.InitRoutes()); err != nil {
			log.Fatal("error occurred on server shutting down")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	if err = srv.Shutdown(context.Background()); err != nil {
		log.Fatal("error occured on server shutting down")
	}
}
