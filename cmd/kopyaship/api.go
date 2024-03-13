package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"path/filepath"
)

type httpClient struct {
	*http.Client
	u *url.URL
}

func newHTTPClient() (*httpClient, error) {
	listen := config.Daemon.API.Listen
	if listen == "ipc" {
		client := &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					dialer := net.Dialer{}
					socketAddr := filepath.Join(cacheDir, "api.socket")
					return dialer.DialContext(ctx, "unix", socketAddr)
				},
			},
		}
		return &httpClient{Client: client}, nil
	} else {
		u, err := url.Parse(listen)
		if err != nil {
			return nil, err
		}
		return &httpClient{Client: &http.Client{}, u: u}, nil
	}
}

func (hc *httpClient) url(path string) string {
	if hc.u == nil {
		return "http://unix" + path
	} else {
		u := *hc.u
		u.Path = path
		return u.String()
	}
}

func (hc *httpClient) CloseIdleConnections() {
	hc.Client.CloseIdleConnections()
}

func (hc *httpClient) NewRequest(method, path string, body io.ReadCloser) (*http.Request, error) {
	return http.NewRequest(method, hc.url(path), body)
}

func (hc *httpClient) Do(req *http.Request) (*http.Response, error) {
	resp, err := hc.Client.Do(req)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		return resp, fmt.Errorf("non-200 status code received: %s", resp.Status)
	}
	return resp, nil
}

func (hc *httpClient) Get(path string) (*http.Response, error) {
	resp, err := hc.Client.Get(hc.url(path))
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		return resp, fmt.Errorf("non-200 status code received: %s", resp.Status)
	}
	return resp, nil
}

func (hc *httpClient) Head(path string) (*http.Response, error) {
	resp, err := hc.Client.Head(hc.url(path))
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		return resp, fmt.Errorf("non-200 status code received: %s", resp.Status)
	}
	return resp, nil
}

func (hc *httpClient) Post(path string, contentType string, body io.Reader) (*http.Response, error) {
	resp, err := hc.Client.Post(hc.url(path), contentType, body)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		return resp, fmt.Errorf("non-200 status code received: %s", resp.Status)
	}
	return resp, nil
}

func (hc *httpClient) PostForm(path string, data url.Values) (*http.Response, error) {
	resp, err := hc.Client.PostForm(hc.url(path), data)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		return resp, fmt.Errorf("non-200 status code received: %s", resp.Status)
	}
	return resp, nil
}
