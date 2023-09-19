package api

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

func doClientRequest(req *http.Request) (*http.Response, error) {
	return http.DefaultClient.Do(req)
}

var ErrorCastRequest = errors.New("decoded request cast error")

func makeRequest(d map[string]any) (*http.Request, error) {
	method, ok := d["method"].(string)
	if !ok {
		return nil, ErrorCastRequest
	}

	path, ok := d["path"].(string)
	if !ok {
		return nil, ErrorCastRequest
	}

	host, ok := d["host"].(string)
	if !ok {
		return nil, ErrorCastRequest
	}

	var body io.Reader
	if _, ok := d["body"]; ok {
		bodyStr, ok := d["body"].(string)
		if !ok {
			return nil, ErrorCastRequest
		}
		body = strings.NewReader(bodyStr)
	}

	uri := fmt.Sprintf("http://%s%s", host, path)

	r, err := http.NewRequest(method, uri, body)
	if err != nil {
		return nil, err
	}

	if _, ok := d["get_params"]; ok {
		params, ok := d["get_params"].(map[any]any)
		if !ok {
			return nil, ErrorCastRequest
		}
		for pI, vListI := range params {
			param, ok := pI.(string)
			if !ok {
				return nil, ErrorCastRequest
			}

			vList, ok := vListI.([]any)
			if !ok {
				return nil, ErrorCastRequest
			}

			for _, v := range vList {
				vStr, ok := v.(string)
				if !ok {
					return nil, ErrorCastRequest
				}

				r.URL.Query().Add(param, vStr)
			}
		}
	}

	if _, ok := d["cookies"]; ok {
		cookies, ok := d["cookies"].(map[string]string)
		if !ok {
			return nil, ErrorCastRequest
		}
		for name, val := range cookies {
			r.AddCookie(&http.Cookie{Name: name, Value: val})
		}
	}

	return r, nil
}
