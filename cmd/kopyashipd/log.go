package main

import (
	"net/url"
	"os"

	"github.com/tomruk/kopyaship/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func (v *svice) newLogger(debug bool) (*zap.Logger, error) {
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
		if utils.RunningOnWindows {
			logFile = "winfile:///" + logFile

			newWinFileSink := func(u *url.URL) (zap.Sink, error) {
				// Remove leading slash left by url.Parse()
				return os.OpenFile(u.Path[1:], os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
			}
			err := zap.RegisterSink("winfile", newWinFileSink)
			if err != nil {
				return nil, err
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
			EncodeTime:    zapcore.ISO8601TimeEncoder,
			// EncodeTime: zapcore.TimeEncoderOfLayout(""),
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
	}

	return logConfig.Build()
}
