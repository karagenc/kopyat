package symbols

import (
	"log"
	"reflect"
	"strconv"

	"github.com/traefik/yaegi/stdlib/unrestricted"
)

type ExitInfo struct {
	Code int
}

func (e *ExitInfo) Error() string { return "exit status code: " + strconv.Itoa(e.Code) }

func init() {
	unrestricted.Symbols["os/os"]["Exit"] = reflect.ValueOf(func(code int) {
		panic(&ExitInfo{Code: code})
	})
	unrestricted.Symbols["log/log"]["Fatal"] = reflect.ValueOf(func(v ...any) {
		log.Print(v...)
		panic(&ExitInfo{Code: 1})
	})
	unrestricted.Symbols["log/log"]["Fatalf"] = reflect.ValueOf(func(format string, v ...any) {
		log.Printf(format, v...)
		panic(&ExitInfo{Code: 1})
	})
	unrestricted.Symbols["log/log"]["Fatalln"] = reflect.ValueOf(func(v ...any) {
		log.Println(v...)
		panic(&ExitInfo{Code: 1})
	})
}
