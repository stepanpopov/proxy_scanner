package proxy

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"syscall"
)

func NewProxyHandler(caCertFile, caKeyFile string, tarantool *TarantoolProxy) (*ProxyHandler, error) {
	caCert, caKey, err := loadX509KeyPair(caCertFile, caKeyFile)
	if err != nil {
		return nil, err
	}

	return &ProxyHandler{
		caCert:    caCert,
		caKey:     caKey,
		tarantool: tarantool,
	}, nil
}

type ProxyHandler struct {
	caCert    *x509.Certificate
	caKey     any
	tarantool *TarantoolProxy
}

func (p *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		p.handleHTTPS(w, r)
		return
	}

	p.handleHTTP(w, r)
}

type ForwardProxyTransport struct {
	http.Transport
}

func (t *ForwardProxyTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Del("Proxy-Connection")

	return t.Transport.RoundTrip(r)
}

type ForwardProxyClient struct {
	http.Client
}

var CheckRedirectDisabler = func(req *http.Request, via []*http.Request) error {
	return http.ErrUseLastResponse
}

func (p ProxyHandler) handleHTTP(w http.ResponseWriter, r *http.Request) {
	if b, err := httputil.DumpRequest(r, true); err == nil {
		log.Printf("incoming request:\n%s\n", string(b))
	}

	r.RequestURI = ""

	client := http.Client{
		Transport:     &ForwardProxyTransport{},
		CheckRedirect: CheckRedirectDisabler,
	}

	copyReq := *r
	resp, err := client.Do(r)
	if err != nil {
		log.Println("client err", err)
		http.Error(w, "failed to proxy %v", http.StatusBadRequest)
		return
	}

	if b, err := httputil.DumpResponse(resp, false); err == nil {
		log.Printf("target response:\n%s\n", string(b))
	}

	hj, ok := w.(http.Hijacker)
	if !ok {
		log.Fatal("http server doesn't support hijacking connection")
	}

	clientConn, _, err := hj.Hijack()
	if err != nil {
		log.Fatal("http hijacking failed")
	}
	defer clientConn.Close()

	copyResp := *resp
	if err := resp.Write(clientConn); err != nil {
		log.Println("error writing response back:", err)
	}

	if err := p.tarantool.insertReqResp(parseRequest(copyReq), parseResponce(copyResp)); err != nil {
		log.Println("failed to insert to tarantool: ", err)
	}
}

func (p ProxyHandler) handleHTTPS(w http.ResponseWriter, proxyReq *http.Request) {
	log.Printf("CONNECT requested to %v (from %v)", proxyReq.Host, proxyReq.RemoteAddr)

	// "Hijack" the client connection to get a TCP (or TLS) socket we can read
	// and write arbitrary data to/from.
	hj, ok := w.(http.Hijacker)
	if !ok {
		log.Fatal("http server doesn't support hijacking connection")
	}

	clientConn, _, err := hj.Hijack()
	if err != nil {
		log.Fatal("http hijacking failed")
	}
	defer clientConn.Close()

	// proxyReq.Host will hold the CONNECT target host, which will typically have
	// a port - e.g. example.org:443
	// To generate a fake certificate for example.org, we have to first split off
	// the host from the port.
	host, _, err := net.SplitHostPort(proxyReq.Host)
	if err != nil {
		log.Fatal("error splitting host/port:", err)
	}

	// Create a fake TLS certificate for the target host, signed by our CA. The
	// certificate will be valid for 10 days - this number can be changed.
	pemCert, pemKey := createCert([]string{host}, p.caCert, p.caKey, 240)
	tlsCert, err := tls.X509KeyPair(pemCert, pemKey)
	if err != nil {
		log.Fatal(err)
	}

	// Send an HTTP OK response back to the client; this initiates the CONNECT
	// tunnel. From this point on the client will assume it's connected directly
	// to the target.
	if _, err := clientConn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n")); err != nil {
		log.Fatal("error writing status to client:", err)
	}

	// Configure a new TLS server, pointing it at the client connection, using
	// our certificate. This server will now pretend being the target.
	tlsConfig := &tls.Config{
		PreferServerCipherSuites: true,
		CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
		MinVersion:               tls.VersionTLS13,
		Certificates:             []tls.Certificate{tlsCert},
	}

	tlsConn := tls.Server(clientConn, tlsConfig)
	defer tlsConn.Close()

	// Create a buffered reader for the client connection; this is required to
	// use http package functions with this connection.
	connReader := bufio.NewReader(tlsConn)

	// Run the proxy in a loop until the client closes the connection.
	for {
		// Read an HTTP request from the client; the request is sent over TLS that
		// connReader is configured to serve. The read will run a TLS handshake in
		// the first invocation (we could also call tlsConn.Handshake explicitly
		// before the loop, but this isn't necessary).
		// Note that while the client believes it's talking across an encrypted
		// channel with the target, the proxy gets these requests in "plain text"
		// because of the MITM setup.
		r, err := http.ReadRequest(connReader)
		if err == io.EOF {
			break
		} else if errors.Is(err, syscall.ECONNRESET) {
			log.Print("This is connection reset by peer error")
			break
		} else if err != nil {
			log.Fatal(proxyReq, err)
		}

		// We can dump the request; log it, modify it...
		if b, err := httputil.DumpRequest(r, false); err == nil {
			log.Printf("incoming request:\n%s\n", string(b))
		}

		// Take the original request and changes its destination to be forwarded
		// to the target server.
		changeRequestToTarget(r, proxyReq.Host)

		// Send the request to the target server and log the response.
		client := http.Client{
			// Transport: &ForwardProxyTransport{},
			// CheckRedirect: CheckRedirectDisabler,
		}

		copyReq := *r
		resp, err := client.Do(r)
		if err != nil {
			log.Println("error sending request to target:", err)
			break
		}
		if b, err := httputil.DumpResponse(resp, false); err == nil {
			log.Printf("target response:\n%s\n", string(b))
		}
		defer resp.Body.Close()

		copyResp := *resp
		// Send the target server's response back to the client.
		if err := resp.Write(tlsConn); err != nil {
			log.Println("error writing response back:", err)
		}

		if err := p.tarantool.insertReqResp(parseRequest(copyReq), parseResponce(copyResp)); err != nil {
			log.Println("failed to insert to tarantool: ", err)
		}
	}
}
