package main

import (
	"flag"
	"log"
	"net/http"

	"github.com/stepanpopov/proxy_scanner/internal"
)

func main() {
	listenAddr := flag.String("listen_addr", ":8000", "host:port")

	server := &http.Server{
		Addr: *listenAddr,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodConnect {
				internal.HanldeHTTPS(w, r)
				return
			}

			internal.HanldeHTTP(w, r)
		}),
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}

}

func init() {
	flag.Parse()
}
