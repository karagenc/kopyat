package scripting

import (
	"context"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"

	_s "github.com/tomruk/kopyaship/scripting/s"
	"github.com/tomruk/kopyaship/utils"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

var symbols = make(map[string]map[string]reflect.Value)

type YaegiScript struct {
	ctx      context.Context
	location string

	i                   *interp.Interpreter
	prog                *interp.Program
	structNamePrimary   string
	structNameSecondary string
}

func newYaegiScript(ctx context.Context, sw ...string) (*YaegiScript, error) {
	i := interp.New(interp.Options{
		Unrestricted: true,
		Args:         sw[1:],
	})

	i.Use(stdlib.Symbols)
	i.Use(symbols)

	base := strings.TrimSuffix(filepath.Base(sw[0]), ".go")
	structNamePascalCase := utils.ConvertToPascalCase(base)
	structNameCamelCase := utils.ConvertToCamelCase(base)

	prog, err := i.CompilePath(sw[0])
	if err != nil {
		return nil, err
	}

	return &YaegiScript{
		ctx:                 ctx,
		location:            sw[0],
		i:                   i,
		prog:                prog,
		structNamePrimary:   structNamePascalCase,
		structNameSecondary: structNameCamelCase,
	}, nil
}

func (s *YaegiScript) Location() string { return s.location }

func (s *YaegiScript) Run() (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()
	_, err = s.i.ExecuteWithContext(s.ctx, s.prog)
	if err != nil {
		return
	}

	_, err = s.i.Eval(`import _scripting_s "github.com/tomruk/kopyaship/scripting/s"`)
	if err != nil {
		return
	}
	varName := utils.RandString(10)
	_, err = s.i.Eval(fmt.Sprintf("var %s _scripting_s.Runner = &%s{}", varName, s.structNamePrimary))
	if err != nil && strings.Contains(err.Error(), "undefined type") {
		_, err = s.i.Eval(fmt.Sprintf("var %s _scripting_s.Runner = &%s{}", varName, s.structNameSecondary))
		if err != nil {
			return
		}
	} else if err != nil {
		return
	}

	var res reflect.Value
	res, err = s.i.Eval(varName)
	if err != nil {
		return
	}
	v, ok := res.Interface().(_s.Runner)
	if !ok {
		return fmt.Errorf("type assertion failed: %s is not convertible to s.Runner. make sure you have correct signature for the method `Run`", s.structNamePrimary)
	}
	v.Run()
	return nil
}
