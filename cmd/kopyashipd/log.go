package main

import (
	"net/url"
	"os"
	"runtime"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func (v *svice) newLogger(debug bool) (*zap.Logger, *logger, error) {
	development := false
	level := zap.NewAtomicLevelAt(zap.InfoLevel)
	if debug {
		level = zap.NewAtomicLevelAt(zap.DebugLevel)
		development = true
	}

	outputPaths := []string{"stdout"}
	logFile := v.config.Daemon.Log
	if logFile != "" {
		// https://github.com/uber-go/zap/issues/621
		if runtime.GOOS == "windows" {
			logFile = "winfile:///" + logFile

			newWinFileSink := func(u *url.URL) (zap.Sink, error) {
				// Remove leading slash left by url.Parse()
				return os.OpenFile(u.Path[1:], os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
			}
			err := zap.RegisterSink("winfile", newWinFileSink)
			if err != nil {
				return nil, nil, err
			}
		}
		outputPaths = append(outputPaths, logFile)
	}

	logConfig := &zap.Config{
		Encoding:    "json",
		Level:       level,
		Development: development,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		OutputPaths: outputPaths,
		EncoderConfig: zapcore.EncoderConfig{
			NameKey:       "logger",
			TimeKey:       "ts",
			LevelKey:      "level",
			CallerKey:     "caller",
			MessageKey:    "msg",
			StacktraceKey: "stacktrace",
			LineEnding:    zapcore.DefaultLineEnding,
			EncodeLevel:   zapcore.LowercaseLevelEncoder,
			EncodeTime:    zapcore.EpochTimeEncoder,
			// EncodeTime: zapcore.TimeEncoderOfLayout(""),
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
	}

	l, err := logConfig.Build()
	if err != nil {
		return nil, nil, err
	}
	return l, newLogger(l), err
}

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

func (l *logger) Debug(a ...any) { l.s.Debug(a...) }

func (l *logger) Debugf(format string, a ...any) {
	l.removeNewLine(&format)
	l.s.Debugf(format, a...)
}

func (l *logger) Debugln(a ...any) { l.s.Debug(a...) }
