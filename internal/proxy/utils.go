package proxy

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func copyHeaders(source, dest http.Header) {
	for headerKey, headerValList := range source {
		for _, headerVal := range headerValList {
			dest.Set(headerKey, headerVal)
		}
	}
}

func changeRequestToTarget(req *http.Request, targetHost string) {
	targetUrl := addrToUrl(targetHost)
	targetUrl.Path = req.URL.Path
	targetUrl.RawQuery = req.URL.RawQuery
	req.URL = targetUrl

	req.RequestURI = ""
}

func addrToUrl(addr string) *url.URL {
	if !strings.HasPrefix(addr, "https") {
		addr = "https://" + addr
	}
	u, err := url.Parse(addr)
	if err != nil {
		log.Fatal(err)
	}
	return u
}

func jsonDecode(b []byte) (any, error) {
	var decoded any
	if err := json.Unmarshal(b, &decoded); err != nil {
		return nil, err
	}
	return decoded, nil
}

func jsonEncodeStr(a any) (string, error) {
	b, err := json.Marshal(a)
	return string(b), err
}
