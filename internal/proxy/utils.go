package proxy

import (
	"encoding/json"
	"io"
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

func parseRequest(r http.Request) map[string]any {
	d := make(map[string]any)
	d["method"] = r.Method
	d["path"] = r.URL.Path

	getParamVals := make(url.Values)
	for k, values := range r.URL.Query() {
		getParamVals[k] = append(getParamVals[k], values...)
	}
	d["get_params"] = getParamVals

	d["host"] = r.Host

	headers := make(http.Header)
	for k, values := range r.Header {
		headers[k] = append(headers[k], values...)
	}
	d["headers"] = headers

	cookies := make(map[string]string)
	for _, v := range r.Cookies() {
		cookies[v.Name] = v.Value
	}
	d["cookes"] = cookies

	if err := r.ParseForm(); err == nil {
		postFormVals := make(url.Values)
		for k, values := range r.PostForm {
			postFormVals[k] = append(postFormVals[k], values...)
		}

		d["post_params"] = postFormVals
	} else {
		body := &strings.Builder{}
		defer r.Body.Close()
		if _, err := io.Copy(body, r.Body); err == nil {
			d["body"] = body.String()
		}
	}

	return d
}

func parseResponce(r http.Response) map[string]any {
	d := make(map[string]any)
	d["code"] = r.StatusCode

	headers := make(http.Header)
	for k, values := range r.Header {
		headers[k] = append(headers[k], values...)
	}
	d["headers"] = headers

	body := &strings.Builder{}
	defer r.Body.Close()
	if _, err := io.Copy(body, r.Body); err == nil {
		d["body"] = body.String()
	}

	return d
}
