package main

import (
	"context"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/tomruk/finddirs-go"
	"github.com/tomruk/kopyaship/internal/utils"
)

const apiSocketFileName = "api.socket"

type httpClient struct {
	*http.Client
	u          *url.URL
	socketAddr string
	basicAuth  func() string
}

func newHTTPClient() (*httpClient, error) {
	listen := config.Service.API.Listen
	basicAuthConfig := config.Service.API.BasicAuth

	setBasicAuth := func(hc *httpClient) {
		if basicAuthConfig.Enabled {
			username := basicAuthConfig.Username
			password := basicAuthConfig.Password
			hc.basicAuth = func() string {
				auth := username + ":" + password
				return base64.StdEncoding.EncodeToString([]byte(auth))
			}

			redirectPolicyFunc := func(req *http.Request, via []*http.Request) error {
				req.Header.Set("Authorization", "Basic "+hc.basicAuth())
				return nil
			}
			hc.Client.CheckRedirect = redirectPolicyFunc
		}
	}

	// Unix socket is supported on Windows 10 Insider Build 17063 and later.
	// For older versions, fall back to HTTP.
	if listen == "ipc" && runtime.GOOS == "windows" {
		tempSocketDir, err := os.MkdirTemp("", "kopyaship_*")
		if err != nil {
			return nil, err
		}
		tempSocketFile := filepath.Join(tempSocketDir, "tmp.socket")
		defer os.RemoveAll(tempSocketDir)

		l, err := net.Listen("unix", tempSocketFile)
		if err != nil {
			opErr, ok := err.(*net.OpError)
			if ok {
				_, ok := opErr.Unwrap().(*os.SyscallError)
				if ok {
					listen = "http://" + utils.APIFallbackAddr
				}
			}
		}
		if l != nil {
			l.Close()
		}
	}

	if listen == "ipc" {
		socketAddr := filepath.Join(stateDir, apiSocketFileName)
		if _, err := os.Stat(socketAddr); os.IsNotExist(err) {
			dirs, err := finddirs.RetrieveAppDirs(true, &utils.FindDirsConfig)
			if err != nil {
				return nil, err
			}
			socketAddrSystemWide := filepath.Join(dirs.StateDir, apiSocketFileName)
			if _, err := os.Stat(socketAddrSystemWide); os.IsNotExist(err) {
				return nil, fmt.Errorf("unix socket file for IPC communication: %s not found in following state directories: %s and %s", apiSocketFileName, filepath.Dir(socketAddr), filepath.Dir(socketAddrSystemWide))
			}
			socketAddr = socketAddrSystemWide
		}

		client := &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
					dialer := net.Dialer{}
					return dialer.DialContext(ctx, "unix", socketAddr)
				},
			},
		}
		hc := &httpClient{Client: client, socketAddr: socketAddr}
		setBasicAuth(hc)
		return hc, nil
	} else {
		u, err := url.Parse(listen)
		if err != nil {
			return nil, err
		}
		hc := &httpClient{Client: &http.Client{}, u: u}
		setBasicAuth(hc)
		return hc, nil
	}
}

func (hc *httpClient) String() string {
	if hc.u == nil {
		return "unix: " + hc.socketAddr
	} else {
		return hc.u.String()
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
	req, err := http.NewRequest(method, hc.url(path), body)
	if err != nil {
		return nil, err
	}
	if hc.basicAuth != nil {
		req.Header.Set("Authorization", "Basic "+hc.basicAuth())
	}
	return req, nil
}

func (hc *httpClient) Do(req *http.Request) (*http.Response, error) {
	if hc.basicAuth != nil {
		req.Header.Set("Authorization", "Basic "+hc.basicAuth())
	}
	resp, err := hc.Client.Do(req)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		return resp, fmt.Errorf("non-200 status code received: %s", resp.Status)
	}
	return resp, nil
}

func (hc *httpClient) Get(path string) (*http.Response, error) {
	req, err := http.NewRequest("GET", hc.url(path), nil)
	if err != nil {
		return nil, err
	}
	if hc.basicAuth != nil {
		req.Header.Set("Authorization", "Basic "+hc.basicAuth())
	}
	resp, err := hc.Client.Do(req)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		return resp, fmt.Errorf("non-200 status code received: %s", resp.Status)
	}
	return resp, nil
}

func (hc *httpClient) Head(path string) (*http.Response, error) {
	req, err := http.NewRequest("HEAD", hc.url(path), nil)
	if err != nil {
		return nil, err
	}
	if hc.basicAuth != nil {
		req.Header.Set("Authorization", "Basic "+hc.basicAuth())
	}
	resp, err := hc.Client.Do(req)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		return resp, fmt.Errorf("non-200 status code received: %s", resp.Status)
	}
	return resp, nil
}

func (hc *httpClient) Post(path string, contentType string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest("POST", hc.url(path), body)
	if err != nil {
		return nil, err
	}
	if hc.basicAuth != nil {
		req.Header.Set("Authorization", "Basic "+hc.basicAuth())
	}
	req.Header.Set("Content-Type", contentType)
	resp, err := hc.Client.Do(req)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		return resp, fmt.Errorf("non-200 status code received: %s", resp.Status)
	}
	return resp, nil
}

func (hc *httpClient) PostForm(path string, data url.Values) (*http.Response, error) {
	return hc.Post(path, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
}

func (s *svc) setupRouter(e *echo.Echo) {
	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "Pong")
	})
	e.GET("/watch-job", s.getWatchJobs)
	e.GET("/watch-job/stop", s.stopWatchJobs)
	e.GET("/service/reload", s.reload)
}

func (s *svc) newAPIServer() (
	e *echo.Echo,
	hs *http.Server,
	listen func() error,
	err error,
) {
	apiConfig := config.Service.API
	e = echo.New()
	hs = &http.Server{
		Handler: e,
	}

	if apiConfig.BasicAuth.Enabled {
		e.Use(middleware.BasicAuth(func(username, password string, ctx echo.Context) (bool, error) {
			u := subtle.ConstantTimeCompare([]byte(username), []byte(apiConfig.BasicAuth.Username))
			p := subtle.ConstantTimeCompare([]byte(password), []byte(apiConfig.BasicAuth.Password))
			if u == 1 && p == 1 {
				return true, nil
			}
			return false, nil
		}))
	}

	s.setupRouter(e)

	if apiConfig.Listen == "ipc" {
		socketPath := filepath.Join(stateDir, apiSocketFileName)
		os.Remove(socketPath)
		listeningOn := " unix socket: " + socketPath

		l, err := net.Listen("unix", socketPath)
		// Unix socket is supported on Windows 10 Insider Build 17063 and later.
		// For older versions, fall back to HTTP.
		if err != nil && runtime.GOOS == "windows" {
			opErr, ok := err.(*net.OpError)
			if ok {
				_, ok := opErr.Unwrap().(*os.SyscallError)
				if ok {
					l, err = net.Listen("tcp", utils.APIFallbackAddr)
					listeningOn = ": http://" + utils.APIFallbackAddr
				}
			}
		}
		if err != nil {
			return nil, nil, nil, err
		}

		s.log.Sugar().Infof("Listening on%s", listeningOn)
		listen = func() error { return hs.Serve(l) }
		return e, hs, listen, err
	} else {
		u, err := url.Parse(apiConfig.Listen)
		if err != nil {
			return nil, nil, nil, err
		} else if u.Path != "/" && u.Path != "" {
			return nil, nil, nil, fmt.Errorf("custom path in URL is not supported")
		}

		switch u.Scheme {
		case "https":
			port := u.Port()
			if port == "" {
				port = "80"
			}
			hs.Addr = u.Hostname() + ":" + port
			s.log.Sugar().Infof("Listening on: %s://%s:%s", u.Scheme, u.Hostname(), port)

			listen = func() error { return hs.ListenAndServeTLS(apiConfig.Cert, apiConfig.Key) }
			return e, hs, listen, nil
		case "http":
			port := u.Port()
			if port == "" {
				port = "80"
			}
			hs.Addr = u.Hostname() + ":" + port
			s.log.Sugar().Infof("Listening on: %s://%s:%s", u.Scheme, u.Hostname(), port)
			listen = func() error { return hs.ListenAndServe() }
			return e, hs, listen, nil
		default:
			return nil, nil, nil, fmt.Errorf("invalid scheme: %s", u.Scheme)
		}
	}
}
