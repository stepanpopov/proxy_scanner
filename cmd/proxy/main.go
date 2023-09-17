package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/stepanpopov/proxy_scanner/internal/proxy"

	"github.com/tarantool/go-tarantool/v2"
)

var (
	ttURI = flag.String("tt_uri", "localhost:3301", "tarantool uri")
	// ttUser     = flag.String("tt_user", "guest", "tarantool user")
	// ttPassword = flag.String("tt_pass", "secret-cluster-cookie", "tarantool password")

	listenAddr = flag.String("listen_addr", ":8000", "host:port")
	caCertFile = flag.String("ca_cert_file", "", "certificate .pem file for trusted CA")
	caKeyFile  = flag.String("ca_key_file", "", "key .pem file for trusted CA")
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	conn, err := tarantool.Connect(*ttURI, tarantool.Opts{})
	if err != nil {
		panic(err)
	}

	if _, err := conn.Ping(); err != nil {
		panic(err)
	}

	proxy, err := proxy.NewProxyHandler(*caCertFile, *caKeyFile, proxy.NewTarantoolProxy(conn))
	if err != nil {
		log.Fatal(err)
	}

	server := &http.Server{
		Addr:    *listenAddr,
		Handler: proxy,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func init() {
	flag.Parse()
}
