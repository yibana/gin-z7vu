package APIClient

import (
	"bytes"
	"errors"
	"golang.org/x/net/proxy"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
)

type APIHttpClient struct {
	client    *http.Client
	addHeader map[string]string
}

func NewAPIClient(proxyAddr string, addHeader map[string]string) *APIHttpClient {
	transport := &http.Transport{}
	if strings.TrimSpace(proxyAddr) != "" {
		u, err := url.Parse(proxyAddr)
		if err != nil {
			panic(err)
		}
		d := net.Dialer{}
		if strings.HasPrefix(proxyAddr, "http") { // http代理
			transport.DialContext = d.DialContext
			transport.Proxy = http.ProxyURL(u)
		} else {
			dialer, err := proxy.FromURL(u, &d)
			if err != nil {
				panic(err)
			}
			// set our socks5 as the dialer
			transport.Dial = dialer.Dial
		}
	}

	client := &http.Client{Transport: transport}

	return &APIHttpClient{
		client:    client,
		addHeader: addHeader,
	}
}

func (ac *APIHttpClient) Get(path string) ([]byte, error) {
	return ac.Request("GET", path, nil)
}

func (ac *APIHttpClient) Post(path string, body []byte) ([]byte, error) {
	return ac.Request("POST", path, body)
}

func (ac *APIHttpClient) Request(method, path string, body []byte) ([]byte, error) {

	var status int
	var Body []byte

	var err error
	var req *http.Request
	switch strings.ToUpper(method) {
	case "GET":
		req, err = http.NewRequest(method, path, nil)
	case "POST":
		req, err = http.NewRequest(method, path, bytes.NewReader(body))
	}

	if err != nil {
		return nil, err
	}
	req.Header.Add("accept", "*/*")
	req.Header.Add("content-type", "application/json")
	for k, v := range ac.addHeader {
		req.Header.Set(k, v)
	}
	res, err := ac.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	status = res.StatusCode

	Body, err = ioutil.ReadAll(res.Body)
	if !(status >= 200 && status < 300) {
		return nil, errors.New(string(Body))
	}
	return Body, err
}
