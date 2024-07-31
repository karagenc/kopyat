package symbols

import (
	"os"
	"reflect"
	"sync"

	"github.com/traefik/yaegi/stdlib"
)

func init() {
	var (
		env    = make(map[string]string)
		mu     sync.Mutex
		getenv = func(key string) string {
			mu.Lock()
			value := env[key]
			mu.Unlock()
			return value
		}
	)
	env["KOPYAT_SCRIPT"] = "1"

	stdlib.Symbols["os/os"]["Clearenv"] = reflect.ValueOf(func() {
		mu.Lock()
		defer mu.Unlock()
		env = make(map[string]string)
		os.Clearenv()
	})
	stdlib.Symbols["os/os"]["ExpandEnv"] = reflect.ValueOf(func(s string) string {
		mu.Lock()
		defer mu.Unlock()
		s = os.Expand(s, getenv)
		s = os.ExpandEnv(s)
		return s
	})
	stdlib.Symbols["os/os"]["Getenv"] = reflect.ValueOf(func(key string) string {
		mu.Lock()
		defer mu.Unlock()
		value, ok := env[key]
		if ok {
			return value
		}
		return os.Getenv(key)
	})
	stdlib.Symbols["os/os"]["LookupEnv"] = reflect.ValueOf(func(key string) (s string, ok bool) {
		mu.Lock()
		defer mu.Unlock()
		s, ok = env[key]
		if ok {
			return
		}
		return os.LookupEnv(key)
	})
	stdlib.Symbols["os/os"]["Setenv"] = reflect.ValueOf(func(key, value string) error {
		mu.Lock()
		defer mu.Unlock()
		_, ok := env[key]
		if ok {
			env[key] = value
			return nil
		}
		return os.Setenv(key, value)
	})
	stdlib.Symbols["os/os"]["Unsetenv"] = reflect.ValueOf(func(key string) error {
		mu.Lock()
		defer mu.Unlock()
		_, ok := env[key]
		if ok {
			delete(env, key)
			return nil
		}
		return os.Unsetenv(key)
	})
	stdlib.Symbols["os/os"]["Environ"] = reflect.ValueOf(func() (a []string) {
		mu.Lock()
		defer mu.Unlock()
		for k, v := range env {
			a = append(a, k+"="+v)
		}
		a = append(a, os.Environ()...)
		return
	})
}
