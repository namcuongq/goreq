package goreq

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

type Req struct {
	URL          string                 `json:"url"`
	Method       string                 `json:"method"`
	Header       map[string]string      `json:"header"`
	PayLoad      map[string]interface{} `json:"payload"`
	RawPayLoad   string                 `json:"raw_payload"`
	BytesPayLoad []byte                 `json:"bytes_payload"`
	Errors       []error                `json:"errors"`
	Transport    *http.Transport        `json:"transport"`
	FilePath     string                 `json:"file_path"`
	FileParam    string                 `json:"file_param"`
}

var ContentTypes = map[string]string{
	"html":       "text/html",
	"text":       "text/plain",
	"json":       "application/json",
	"xml":        "application/xml",
	"urlencoded": "application/x-www-form-urlencoded",
	"form":       "application/x-www-form-urlencoded",
	"form-data":  "application/x-www-form-urlencoded",
	"stream":     "application/octet-stream",
}

func (req *Req) Proxy(proxyURL string) *Req {
	parsedProxyURL, err := url.Parse(proxyURL)
	if err != nil {
		req.Errors = append(req.Errors, err)
		return req
	}

	transport := http.Transport{}
	transport.Proxy = http.ProxyURL(parsedProxyURL)
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	req.Transport = &transport
	return req

}

func (req *Req) ContentType(typeStr string) *Req {
	if ContentTypes[typeStr] != "" {
		typeStr = ContentTypes[typeStr]
	}
	req.SetHeader("Content-Type", typeStr)
	return req
}

func (req *Req) SetHeader(key, value string) *Req {
	req.Header[key] = value
	return req
}

func (req *Req) SendRawBytes(payload []byte) *Req {
	if req.Header["Content-Type"] == "" {
		req.Header["Content-Type"] = "application/octet-stream"
	}
	req.BytesPayLoad = payload
	return req
}

func (req *Req) SendRawString(payload string) *Req {
	if req.Header["Content-Type"] == "" {
		req.Header["Content-Type"] = "text/plain"
	}
	req.RawPayLoad = payload
	return req
}

func (req *Req) SendFile(key, path string) *Req {
	req.FilePath = path
	req.FileParam = key
	return req
}

func (req *Req) SendStruct(payload interface{}) *Req {
	data, err := json.Marshal(payload)
	if err != nil {
		req.Errors = append(req.Errors, err)
		return req
	}
	var s map[string]interface{}

	err = json.Unmarshal(data, &s)
	if err != nil {
		req.Errors = append(req.Errors, err)
		return req
	}

	req.PayLoad = s
	return req
}

func newfileUploadRequest(params map[string]interface{}, FileParam, FilePath string) (*bytes.Buffer, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	file, err := os.Open(FilePath)
	if err != nil {
		return body, err
	}
	defer file.Close()

	part, err := writer.CreateFormFile(FileParam, filepath.Base(file.Name()))
	if err != nil {
		return body, err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return body, err
	}

	for key, val := range params {
		_ = writer.WriteField(key, fmt.Sprintf("%v", val))
	}

	err = writer.Close()
	return body, err
}

func changeMapToURLValues(data map[string]interface{}) url.Values {
	var newURLValues = url.Values{}
	for k, v := range data {
		switch val := v.(type) {
		case string:
			newURLValues.Add(k, val)
		case []string:
			for _, element := range val {
				newURLValues.Add(k, element)
			}
		// if a number, change to string
		// json.Number used to protect against a wrong (for GoReq) default conversion
		// which always converts number to float64.
		// This type is caused by using Decoder.UseNumber()
		case json.Number:
			newURLValues.Add(k, string(val))
		}
	}
	return newURLValues
}
