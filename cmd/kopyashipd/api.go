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

func (v *svice) serve() (err error) {
	var (
		e = echo.New()
		s = http.Server{
			Handler: e,
		}
	)

	setupRouter(e)

	if v.config.Daemon.API.Listen == "ipc" {
		socketPath := filepath.Join(v.cacheDir, "api.socket")
		l, err := net.Listen("unix", socketPath)
		if err != nil {
			return err
		}
		v.log.Sugar().Infof("Listening on unix socket: %s", socketPath)
		return s.Serve(l)
	} else {
		u, err := url.Parse(v.config.Daemon.API.Listen)
		if err != nil {
			return err
		} else if u.Path != "/" && u.Path != "" {
			return fmt.Errorf("custom path in URL not supported")
		}

		switch u.Scheme {
		case "https":
			port := u.Port()
			if port == "" {
				port = "80"
			}
			s.Addr = u.Hostname() + ":" + port
			v.log.Sugar().Infof("Listening on: %s://%s:%s", u.Scheme, u.Hostname(), port)
			return s.ListenAndServeTLS(v.config.Daemon.API.Cert, v.config.Daemon.API.Key)
		case "http":
			port := u.Port()
			if port == "" {
				port = "80"
			}
			s.Addr = u.Hostname() + ":" + port
			v.log.Sugar().Infof("Listening on: %s://%s:%s", u.Scheme, u.Hostname(), port)
			return s.ListenAndServe()
		default:
			return fmt.Errorf("invalid scheme: %s", u.Scheme)
		}
	}
}
