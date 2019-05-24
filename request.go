package goreq

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
)

func New(url string) *Req {
	var req = new(Req)
	req.URL = url
	return req
}

func (req *Req) Get() *Req {
	req.Method = http.MethodGet
	return req
}

func (req *Req) Post() *Req {
	req.Method = http.MethodPost
	return req
}

func (req *Req) Put() *Req {
	req.Method = http.MethodPut
	return req
}

func (req *Req) Delete() *Req {
	req.Method = http.MethodDelete
	return req
}

func (req *Req) Head() *Req {
	req.Method = http.MethodHead
	return req
}

func (req *Req) Options() *Req {
	req.Method = http.MethodOptions
	return req
}

func (req *Req) Patch() *Req {
	req.Method = http.MethodPatch
	return req
}

func (req *Req) Call() (string, *http.Response, []error) {
	client := &http.Client{}
	var request *http.Request
	var err error

	if len(req.Errors) > 0 {
		return "", nil, req.Errors
	}

	for k, v := range req.Header {
		request.Header.Set(k, v)
	}

	if req.Transport != nil {
		client.Transport = req.Transport
	}

	switch req.Method {
	case http.MethodGet, http.MethodHead, http.MethodDelete, http.MethodOptions:
		request, err = http.NewRequest(req.Method, req.URL, nil)
	case http.MethodPost, http.MethodPut, http.MethodPatch:
		if req.Header["Content-Type"] == "" {
			req.Header["Content-Type"] = "application/json"
		}

		if req.FilePath != "" {
			payload, err := newfileUploadRequest(req.PayLoad, req.FileParam, req.FilePath)
			if err == nil {
				request, err = http.NewRequest(req.Method, req.URL, payload)
			}
		} else if req.Header["Content-Type"] == "application/json" && len(req.PayLoad) > 0 { //json
			contentJSON, err := json.Marshal(req.PayLoad)
			if err == nil {
				contentReader := bytes.NewReader(contentJSON)
				request, err = http.NewRequest(req.Method, req.URL, contentReader)
			}
		} else if req.Header["Content-Type"] == "application/x-www-form-urlencoded" { //form
			formData := changeMapToURLValues(req.PayLoad)
			request, err = http.NewRequest(req.Method, req.URL, strings.NewReader(formData.Encode()))
		} else if len(req.BytesPayLoad) > 0 { //raw bytes
			request, err = http.NewRequest(req.Method, req.URL, bytes.NewReader(req.BytesPayLoad))
		} else { //raw string
			request, err = http.NewRequest(req.Method, req.URL, strings.NewReader(req.RawPayLoad))
		}
	default:
		req.Errors = append(req.Errors, errors.New("No method specified"))
	}

	if err != nil {
		req.Errors = append(req.Errors, err)
	}

	return req.do(client, request)
}

func (req *Req) do(client *http.Client, request *http.Request) (string, *http.Response, []error) {
	if len(req.Errors) > 0 {
		return "", nil, req.Errors
	}

	resp, err := client.Do(request)
	if err != nil {
		req.Errors = append(req.Errors, err)
		return "", resp, req.Errors
	}

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		req.Errors = append(req.Errors, err)
	}

	return string(bodyBytes), resp, req.Errors

}
