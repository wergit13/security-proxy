package proxy

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"net/textproto"
	"sc-proxy/pkg/service"
	"strings"
)

var hopHeaders = []string{
	"Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te", // canonicalized version of "TE"
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
}

type httpProxy struct {
	Director  func(*http.Request)
	Transport http.RoundTripper
	Service   *service.Service
}

func httpDirector(r *http.Request) {
	r.URL.Host = r.Host
	r.URL.Scheme = "http"
}

func httpsDirector(r *http.Request) {
	r.URL.Host = r.Host
	r.URL.Scheme = "https"
}

func removeHopByHopHeaders(h http.Header) {
	for _, f := range h["Connection"] {
		for _, sf := range strings.Split(f, ",") {
			if sf = textproto.TrimString(sf); sf != "" {
				h.Del(sf)
			}
		}
	}

	for _, f := range hopHeaders {
		h.Del(f)
	}
}

func (p *httpProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	transport := p.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	ctx := r.Context()
	if ctx.Done() != nil {

	} else if cn, ok := w.(http.CloseNotifier); ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithCancel(ctx)
		defer cancel()
		notifyChan := cn.CloseNotify()
		go func() {
			select {
			case <-notifyChan:
				cancel()
			case <-ctx.Done():
			}
		}()
	}

	outr := r.Clone(ctx)
	p.Director(outr)
	removeHopByHopHeaders(outr.Header)
	rqBodyBytes, err := io.ReadAll(outr.Body)
	if err != nil {
		http.Error(w, "Error reading request", http.StatusInternalServerError)
		return
	}
	outr.Body.Close()
	if len(rqBodyBytes) != 0 {
		outr.Body = io.NopCloser(bytes.NewBuffer(rqBodyBytes))
	} else {
		outr.Body = http.NoBody
	}

	resp, err := transport.RoundTrip(outr)
	if err != nil {
		log.Println("ERROR", outr.Method, outr.Host, "\t", err)
		if len(rqBodyBytes) != 0 {
			outr.Body = io.NopCloser(bytes.NewBuffer(rqBodyBytes))
		} else {
			outr.Body = http.NoBody
		}
		p.Service.SavePair(ctx, outr, nil)
		http.Error(w, "Error sending proxy request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()
	outr.Body = io.NopCloser(bytes.NewBuffer(rqBodyBytes))
	p.Service.SavePair(ctx, outr, resp)
	log.Println(outr.Method, outr.Host, "\t", resp.Status)

	removeHopByHopHeaders(resp.Header)

	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		log.Println("ERROR: ", err)
	}
}
