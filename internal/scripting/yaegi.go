package scripting

import (
	"context"
	"fmt"
	"reflect"

	"github.com/karagenc/kopyat"
	"github.com/karagenc/kopyat/internal/scripting/ctx"
	"github.com/karagenc/kopyat/internal/scripting/symbols"
	"github.com/mitchellh/go-homedir"
	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

type YaegiScript struct {
	path       string
	ctx        context.Context
	i          *interp.Interpreter
	prog       *interp.Program
	getContext func() ctx.Context
}

func newYaegiScript(ctx context.Context, sw ...string) (*YaegiScript, error) {
	scriptPath := sw[0]
	scriptPath, err := homedir.Expand(scriptPath)
	if err != nil {
		return nil, err
	}

	i := interp.New(interp.Options{
		Unrestricted: true,
		Args:         sw[1:],
	})
	s := &YaegiScript{
		ctx:  ctx,
		path: scriptPath,
		i:    i,
	}

	err = i.Use(stdlib.Symbols)
	if err != nil {
		return nil, err
	}

	symbols := symbols.Clone()
	symbols["github.com/karagenc/kopyat/kopyat"]["GetContext"] = reflect.ValueOf(func() kopyat.Context {
		return s.getContext()
	})
	err = i.Use(symbols)
	if err != nil {
		return nil, err
	}

	s.prog, err = i.CompilePath(scriptPath)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (s *YaegiScript) Path() string { return s.path }

func (s *YaegiScript) Run(c ctx.Context) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
		}
	}()

	s.getContext = func() ctx.Context { return c }
	_, err = s.i.ExecuteWithContext(s.ctx, s.prog)
	return
}
