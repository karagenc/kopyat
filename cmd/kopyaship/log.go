package main

import (
	"fmt"
	"os"
)

type logger struct{}

func newLogger() *logger { return &logger{} }

func (l *logger) Info(a ...any) { fmt.Print(a...) }

func (l *logger) Infof(format string, a ...any) { fmt.Printf(format, a...) }

func (l *logger) Infoln(a ...any) { fmt.Println(a...) }

func (l *logger) Warn(a ...any) { fmt.Print(a...) }

func (l *logger) Warnf(format string, a ...any) { fmt.Printf(format, a...) }

func (l *logger) Warnln(a ...any) { fmt.Println(a...) }

func (l *logger) Error(a ...any) { fmt.Fprint(os.Stderr, a...) }

func (l *logger) Errorf(format string, a ...any) { fmt.Fprintf(os.Stderr, format, a...) }

func (l *logger) Errorln(a ...any) { fmt.Fprintln(os.Stderr, a...) }
