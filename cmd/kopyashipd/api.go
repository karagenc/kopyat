package main

import (
	"crypto/subtle"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (v *svice) setupRouter(e *echo.Echo) {
	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "Pong")
	})
	e.GET("/jobs/watch", v.getWatchJobs)
}

func (v *svice) newAPIServer() (e *echo.Echo, s *http.Server, listen func() error, err error) {
	apiConfig := v.config.Daemon.API
	e = echo.New()
	s = &http.Server{
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

	v.setupRouter(e)

	if apiConfig.Listen == "ipc" {
		socketPath := filepath.Join(v.cacheDir, "api.socket")
		os.Remove(socketPath)
		listeningOn := " unix socket: " + socketPath
		const apiFallbackAddr = "127.0.0.1:56792"

		l, err := net.Listen("unix", socketPath)
		if err != nil && runtime.GOOS == "windows" {
			opErr, ok := err.(*net.OpError)
			if ok {
				_, ok := opErr.Unwrap().(*os.SyscallError)
				if ok {
					l, err = net.Listen("tcp", apiFallbackAddr)
					listeningOn = ": http://" + apiFallbackAddr
				}
			}
		}
		if err != nil {
			return nil, nil, nil, err
		}

		v.log.Sugar().Infof("Listening on%s", listeningOn)
		listen = func() error { return s.Serve(l) }
		return e, s, listen, err
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
			s.Addr = u.Hostname() + ":" + port
			v.log.Sugar().Infof("Listening on: %s://%s:%s", u.Scheme, u.Hostname(), port)

			listen = func() error { return s.ListenAndServeTLS(apiConfig.Cert, apiConfig.Key) }
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
