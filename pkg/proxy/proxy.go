package proxy

import (
	"crypto/tls"
	"errors"
	"log"
	"net"
	"net/http"
	"sync"
)

// Proxy is a forward proxy that substitutes its own certificate
// for incoming TLS connections in place of the upstream server's
// certificate.
type Proxy struct {
	// CA specifies the root CA for generating leaf certs for each incoming
	// TLS request.
	CA *tls.Certificate

	// TLSServerConfig specifies the tls.Config to use when generating leaf
	// cert using CA.
	TLSServerConfig *tls.Config

	// TLSClientConfig specifies the tls.Config to use when establishing
	// an upstream connection for proxying.
	TLSClientConfig *tls.Config
}

func (p *Proxy) StartProxy(addr string) error {
	return http.ListenAndServe(addr, p)
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == "CONNECT" {
		p.serveConnect(w, r)
		return
	}
	rp := &httpProxy{
		Director: httpDirector,
	}
	rp.ServeHTTP(w, r)
}

func (p *Proxy) serveConnect(w http.ResponseWriter, r *http.Request) {
	var (
		err   error
		sconn *tls.Conn
		name  = dnsName(r.Host)
	)

	if name == "" {
		log.Println("cannot determine cert name for " + r.Host)
		http.Error(w, "no upstream", http.StatusServiceUnavailable)
		return
	}

	provisionalCert, err := p.cert(name)
	if err != nil {
		log.Println("cert", err)
		http.Error(w, "no upstream", http.StatusServiceUnavailable)
		return
	}

	sConfig := new(tls.Config)
	if p.TLSServerConfig != nil {
		*sConfig = *p.TLSServerConfig
	}
	sConfig.Certificates = []tls.Certificate{*provisionalCert}
	sConfig.GetCertificate = func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
		cConfig := new(tls.Config)
		if p.TLSClientConfig != nil {
			*cConfig = *p.TLSClientConfig
		}
		cConfig.ServerName = hello.ServerName
		sconn, err = tls.Dial("tcp", r.Host, cConfig)
		if err != nil {
			log.Println("dial", r.Host, err)
			return nil, err
		}
		return p.cert(hello.ServerName)
	}

	cconn, err := handshake(w, sConfig)
	if err != nil {
		log.Println("handshake", r.Host, err)
		return
	}
	defer cconn.Close()
	if sconn == nil {
		log.Println("could not determine cert name for " + r.Host)
		return
	}
	defer sconn.Close()

	od := &oneShotDialer{c: sconn}
	rp := &httpProxy{
		Director:  httpsDirector,
		Transport: &http.Transport{DialTLS: od.Dial},
	}

	ch := make(chan int)
	wc := &onCloseConn{cconn, func() { ch <- 0 }}
	http.Serve(&oneShotListener{wc}, rp)
	<-ch
}

func (p *Proxy) cert(names ...string) (*tls.Certificate, error) {
	return genCert(p.CA, names)
}

var okHeader = []byte("HTTP/1.1 200 OK\r\n\r\n")

// handshake hijacks w's underlying net.Conn, responds to the CONNECT request
// and manually performs the TLS handshake. It returns the net.Conn or and
// error if any.
func handshake(w http.ResponseWriter, config *tls.Config) (net.Conn, error) {
	raw, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		http.Error(w, "no upstream", http.StatusServiceUnavailable)
		return nil, err
	}
	if _, err = raw.Write(okHeader); err != nil {
		raw.Close()
		return nil, err
	}
	conn := tls.Server(raw, config)
	err = conn.Handshake()
	if err != nil {
		conn.Close()
		raw.Close()
		return nil, err
	}
	return conn, nil
}

// dnsName returns the DNS name in addr, if any.
func dnsName(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return ""
	}
	return host
}

// A oneShotDialer implements net.Dialer whos Dial only returns a
// net.Conn as specified by c followed by an error for each subsequent Dial.
type oneShotDialer struct {
	c  net.Conn
	mu sync.Mutex
}

func (d *oneShotDialer) Dial(network, addr string) (net.Conn, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.c == nil {
		return nil, errors.New("closed")
	}
	c := d.c
	d.c = nil
	return c, nil
}

// A oneShotListener implements net.Listener whos Accept only returns a
// net.Conn as specified by c followed by an error for each subsequent Accept.
type oneShotListener struct {
	c net.Conn
}

func (l *oneShotListener) Accept() (net.Conn, error) {
	if l.c == nil {
		return nil, errors.New("closed")
	}
	c := l.c
	l.c = nil
	return c, nil
}

func (l *oneShotListener) Close() error {
	return nil
}

func (l *oneShotListener) Addr() net.Addr {
	return l.c.LocalAddr()
}

// A onCloseConn implements net.Conn and calls its f on Close.
type onCloseConn struct {
	net.Conn
	f func()
}

func (c *onCloseConn) Close() error {
	if c.f != nil {
		c.f()
		c.f = nil
	}
	return c.Conn.Close()
}
