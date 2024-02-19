package main

import (
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

var headersToDel = []string{
	"Connection",
	"Proxy-Connection",
	"Keep-Alive",
	"Proxy-Authenticate",
	"Proxy-Authorization",
	"Te", // canonicalized version of "TE"
	"Trailers",
	"Transfer-Encoding",
	"Upgrade",
	"Accept-Encoding",
}

var transport = http.DefaultTransport

func toRelative(abs *url.URL) (rel *url.URL) {
	rel = abs
	rel.Host = ""
	return rel
}

func handleRequest(w http.ResponseWriter, r *http.Request) {
	//replace url with relative

	a, err := httputil.DumpRequestOut(r, false)
	myString := string(a)
	print(myString)
	r.URL = toRelative(r.URL)
	//del headers
	b, err := httputil.DumpRequestOut(r, false)
	myString2 := string(b)
	print(myString2)
	for _, header := range headersToDel {
		r.Header.Del(header)
	}

	// Send the proxy request using the custom transport
	resp, err := transport.RoundTrip(r)
	if err != nil {
		http.Error(w, "Error sending proxy request", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Copy the headers from the proxy response to the original response
	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}

	// Set the status code of the original response to the status code of the proxy response
	w.WriteHeader(resp.StatusCode)

	// Copy the body of the proxy response to the original response
	io.Copy(w, resp.Body)
}

func main() {
	// Create a new HTTP server with the handleRequest function as the handler
	server := http.Server{
		Addr:    ":8080",
		Handler: http.HandlerFunc(handleRequest),
	}

	// Start the server and log any errors
	log.Println("Starting proxy server on :8080")
	err := server.ListenAndServe()
	if err != nil {
		log.Fatal("Error starting proxy server: ", err)
	}
}
