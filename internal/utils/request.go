package utils

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type RequestInfo struct {
	Method      string
	Path        string
	Host        string
	Get_Params  map[string][]string
	Headers     map[string][]string
	Cookies     map[string]string
	Post_Params map[string][]string
	Body        string
}

func ParseRequest(r http.Request) *RequestInfo {
	ri := &RequestInfo{
		Method: r.Method,
		Path:   r.URL.Path,
	}

	getParamVals := make(url.Values)
	for k, values := range r.URL.Query() {
		getParamVals[k] = append(getParamVals[k], values...)
	}
	ri.Get_Params = getParamVals

	ri.Host = r.Host

	headers := make(http.Header)
	for k, values := range r.Header {
		headers[k] = append(headers[k], values...)
	}
	ri.Headers = headers

	cookies := make(map[string]string)
	for _, v := range r.Cookies() {
		cookies[v.Name] = v.Value
	}
	ri.Cookies = cookies

	if err := r.ParseForm(); err == nil {
		postFormVals := make(url.Values)
		for k, values := range r.PostForm {
			postFormVals[k] = append(postFormVals[k], values...)
		}

		ri.Post_Params = postFormVals
	} else {
		body := &strings.Builder{}
		defer r.Body.Close()
		if _, err := io.Copy(body, r.Body); err == nil {
			ri.Body = body.String()
		}
	}

	return ri
}

func MakeRequest(ri *RequestInfo) (*http.Request, error) {
	var body io.Reader
	if ri.Body != "" {
		body = strings.NewReader(ri.Body)
	}

	r, err := http.NewRequest(
		ri.Method,
		fmt.Sprintf("http://%s%s", ri.Host, ri.Path),
		body,
	)
	if err != nil {
		return nil, err
	}

	query := r.URL.Query()
	for param, valList := range ri.Get_Params {
		for _, val := range valList {
			query.Add(param, val)
		}
	}
	r.URL.RawQuery = query.Encode()

	for name, val := range ri.Cookies {
		r.AddCookie(&http.Cookie{Name: name, Value: val})
	}

	for header, headerValList := range ri.Headers {
		for _, headerVal := range headerValList {
			r.Header.Add(header, headerVal)
		}
	}

	for param, valList := range ri.Post_Params {
		for _, val := range valList {
			r.PostForm.Add(param, val)
		}
	}

	return r, nil
}
