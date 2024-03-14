package main

import (
	"strings"

	"go.uber.org/zap"
)

type logger struct {
	z *zap.Logger
	s *zap.SugaredLogger
}

func newLogger(z *zap.Logger) *logger { return &logger{z: z, s: z.Sugar()} }

func (l *logger) removeNewLine(s *string) {
	*s = strings.ReplaceAll(*s, "\n", "")
}

func (l *logger) Info(a ...any) { l.s.Info(a...) }

func (l *logger) Infof(format string, a ...any) {
	l.removeNewLine(&format)
	l.s.Infof(format, a...)
}

func (l *logger) Infoln(a ...any) { l.s.Info(a...) }

func (l *logger) Warn(a ...any) { l.s.Warn(a...) }

func (l *logger) Warnf(format string, a ...any) {
	l.removeNewLine(&format)
	l.s.Warnf(format, a...)
}

func (l *logger) Warnln(a ...any) { l.s.Warn(a...) }

func (l *logger) Error(a ...any) { l.s.Error(a...) }

func (l *logger) Errorf(format string, a ...any) {
	l.removeNewLine(&format)
	l.s.Errorf(format, a...)
}

func (l *logger) Errorln(a ...any) { l.s.Error(a...) }
