package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
)

var sharedTransport = &http.Transport{
	MaxIdleConns:        10000,
	MaxIdleConnsPerHost: 10000,
	MaxConnsPerHost:     0,
	IdleConnTimeout:     90 * time.Second,
	DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
		dialer := &net.Dialer{}
		return dialer.DialContext(ctx, network, addr) // no TLS wrapping
	},
}

func main() {
	target := os.Getenv("TARGET_URL")
	if target == "" {
		log.Fatal("TARGET_URL not set")
	}
	StartServer(":8080", target)
	select {} // keep running
}

func StartServer(addr, target string) *http.Server {
	remote, err := url.Parse(target)
	if err != nil {
		log.Fatalf("Invalid TARGET_URL: %v", err)
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.Transport = sharedTransport

	server := &http.Server{
		Addr:    addr,
		Handler: proxy,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()
	return server
}
