package main

import (
	"flag"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/stepanpopov/proxy_scanner/internal/api"
	"github.com/tarantool/go-tarantool/v2"
)

var (
	ttURI = flag.String("tt_uri", "localhost:3301", "tarantool uri")
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	r := gin.Default()

	conn, err := tarantool.Connect(*ttURI, tarantool.Opts{})
	if err != nil {
		panic(err)
	}

	if _, err := conn.Ping(); err != nil {
		panic(err)
	}

	r.GET("/requests", api.GetAll(conn))
	r.GET("/requests/*id", api.Get(conn))
	r.GET("/repeat/*id", api.Repeat(conn))
	// r.GET("/scan/*id", api.Get(conn))

	r.Run(":8080")
}

func init() {
	flag.Parse()
}
