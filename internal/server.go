package internal

import (
	"io"
	"net/http"
)

type ForwardProxyTransport struct {
	http.Transport
}

func (t *ForwardProxyTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.URL.Host = r.Host
	r.Header.Del("Proxy-Connection")

	return t.Transport.RoundTrip(r)
}

func copyHeaders(source, dest http.Header) {
	for headerKey, headerValList := range source {
		for _, headerVal := range headerValList {
			dest.Set(headerKey, headerVal)
		}
	}
}

func HanldeHTTP(w http.ResponseWriter, r *http.Request) {

	client := &http.Client{Transport: &ForwardProxyTransport{}}

	resp, err := client.Do(r)
	if err != nil {
		http.Error(w, "failed to proxy", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
	copyHeaders(w.Header(), resp.Header)
}

func HanldeHTTPS(w http.ResponseWriter, r *http.Request) {

}
