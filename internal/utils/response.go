package utils

import (
	"io"
	"net/http"
	"strings"
)

type ResponseInfo struct {
	Code    int
	Headers map[string][]string
	Body    string
}

func ParseResponse(r http.Response) ResponseInfo {
	ri := ResponseInfo{
		Code: r.StatusCode,
	}

	headers := make(http.Header)
	for k, values := range r.Header {
		headers[k] = append(headers[k], values...)
	}
	ri.Headers = headers

	body := &strings.Builder{}
	defer r.Body.Close()
	if _, err := io.Copy(body, r.Body); err == nil {
		ri.Body = body.String()
	}

	return ri
}
