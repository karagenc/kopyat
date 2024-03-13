package main

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/labstack/echo/v4"
)

func setupRouter(e *echo.Echo) {}

func (v *svice) newAPIServer() (e *echo.Echo, s *http.Server, listen func() error, err error) {
	e = echo.New()
	s = &http.Server{
		Handler: e,
	}

	setupRouter(e)

	if v.config.Daemon.API.Listen == "ipc" {
		socketPath := filepath.Join(v.cacheDir, "api.socket")
		l, err := net.Listen("unix", socketPath)
		if err != nil {
			return nil, nil, nil, err
		}
		v.log.Sugar().Infof("Listening on unix socket: %s", socketPath)
		listen = func() error { return s.Serve(l) }
		return e, s, listen, err
	} else {
		u, err := url.Parse(v.config.Daemon.API.Listen)
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
			s.Addr = u.Hostname() + ":" + port
			v.log.Sugar().Infof("Listening on: %s://%s:%s", u.Scheme, u.Hostname(), port)

			listen = func() error { return s.ListenAndServeTLS(v.config.Daemon.API.Cert, v.config.Daemon.API.Key) }
			return e, s, listen, nil
		case "http":
			port := u.Port()
			if port == "" {
				port = "80"
			}
			s.Addr = u.Hostname() + ":" + port
			v.log.Sugar().Infof("Listening on: %s://%s:%s", u.Scheme, u.Hostname(), port)
			listen = func() error { return s.ListenAndServe() }
			return e, s, listen, nil
		default:
			return nil, nil, nil, fmt.Errorf("invalid scheme: %s", u.Scheme)
		}
	}
}
