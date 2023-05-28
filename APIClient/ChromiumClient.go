package APIClient

import (
	"bytes"
	"fmt"
	"github.com/Danny-Dasilva/CycleTLS/cycletls"
	fhttp "github.com/saucesteals/fhttp"
	"github.com/saucesteals/mimic"
	"golang.org/x/net/proxy"
	"io/ioutil"
	"net"
	"net/url"
	"strings"
)

type ChromiumClient struct {
	IMAPClient *fhttp.Client
	UserAgent  string
	m          *mimic.ClientSpec
}

func New(proxyAddr string) (*ChromiumClient, error) {
	//https://versionhistory.googleapis.com/v1/chrome/platforms/win/channels/stable/versions
	latestVersion := "114.0.5735.45" //mimic.MustGetLatestVersion(mimic.PlatformWindows)
	m, _ := mimic.Chromium(mimic.BrandChrome, latestVersion)
	fhttpTransport := &fhttp.Transport{ /* Proxy: ... */ }
	if strings.TrimSpace(proxyAddr) != "" {
		u, err := url.Parse(proxyAddr)
		if err != nil {
			return nil, err
		}
		d := net.Dialer{}
		if strings.HasPrefix(proxyAddr, "http") { // http代理
			fhttpTransport.DialContext = d.DialContext
			fhttpTransport.Proxy = fhttp.ProxyURL(u)
		} else {
			dialer, err := proxy.FromURL(u, &d)
			if err != nil {
				return nil, err
			}
			// set our socks5 as the dialer
			fhttpTransport.Dial = dialer.Dial
		}
	}
	minicclient := &fhttp.Client{Transport: m.ConfigureTransport(fhttpTransport)}
	UserAgent := fmt.Sprintf("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/%s Safari/537.36", m.Version())

	return &ChromiumClient{IMAPClient: minicclient, UserAgent: UserAgent, m: m}, nil
}

func (client *ChromiumClient) Get(_url string) ([]byte, error) {
	//nowTime := time.Now()
	//defer func() {
	//	fmt.Printf("[%d]ms [GET]: %s\n", time.Now().Sub(nowTime).Milliseconds(), _url)
	//}()
	req, err := fhttp.NewRequest("GET", _url, nil)
	resp, err := client.ClientDo(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	encoding := resp.Header["Content-Encoding"]
	content := resp.Header["Content-Type"]

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	Body := cycletls.DecompressBody(bodyBytes, encoding, content)
	return []byte(Body), nil
}

func (client *ChromiumClient) Post(_url string, data []byte) ([]byte, error) {
	//nowTime := time.Now()
	//defer func() {
	//	fmt.Printf("[%d]ms [POST]: %s\n", time.Now().Sub(nowTime).Milliseconds(), _url)
	//}()
	req, err := fhttp.NewRequest("POST", _url, bytes.NewBuffer(data))
	resp, err := client.ClientDo(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	//return ioutil.ReadAll(resp.Body)
	encoding := resp.Header["Content-Encoding"]
	content := resp.Header["Content-Type"]

	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	Body := cycletls.DecompressBody(bodyBytes, encoding, content)

	return []byte(Body), nil
}

func (client *ChromiumClient) ClientDo(req *fhttp.Request) (*fhttp.Response, error) {
	Header := fhttp.Header{
		"sec-ch-ua":          {client.m.ClientHintUA()},
		"content-type":       {"application/json"},
		"rtt":                {"50"},
		"sec-ch-ua-mobile":   {"?0"},
		"user-agent":         {client.UserAgent},
		"accept":             {"text/html,*/*"},
		"x-requested-with":   {"XMLHttpRequest"},
		"downlink":           {"3.9"},
		"ect":                {"4g"},
		"sec-ch-ua-platform": {`"Windows"`},
		"sec-fetch-site":     {"same-origin"},
		"sec-fetch-mode":     {"cors"},
		"sec-fetch-dest":     {"empty"},
		"accept-encoding":    {"gzip, deflate, br"},
		"accept-language":    {"en,en_US;q=0.9"},
		fhttp.HeaderOrderKey: {
			"sec-ch-ua", "rtt", "sec-ch-ua-mobile",
			"user-agent", "accept", "x-requested-with",
			"downlink", "ect", "sec-ch-ua-platform",
			"sec-fetch-site", "sec-fetch-mode", "sec-fetch-dest",
			"accept-encoding", "accept-language",
		},
		fhttp.PHeaderOrderKey: client.m.PseudoHeaderOrder(),
	}
	for name, values := range req.Header {
		for _, value := range values {
			Header.Set(name, value)
		}
	}
	req.Header = Header
	return client.IMAPClient.Do(req)
}
