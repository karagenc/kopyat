package main

import (
	"fmt"
	"net/url"
	"os"
	"runtime"

	"github.com/tomruk/kopyaship/internal/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func (s *svc) newLogger(debug bool) (*zap.Logger, error) {
	development := false
	level := zap.NewAtomicLevelAt(zap.InfoLevel)
	if debug {
		level = zap.NewAtomicLevelAt(zap.DebugLevel)
		development = true
	}

	outputPaths := []string{"stdout"}
	logFile := config.Service.Log
	if logFile == "disabled" {
		return zap.NewNop(), nil
	} else if logFile != "" {
		// https://github.com/uber-go/zap/issues/621
		if runtime.GOOS == "windows" {
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

func initLogging() (err error) {
	enable, _ := rootCmd.PersistentFlags().GetBool("enable-log")
	if enable {
		debugLog, err = utils.NewDebugLogger()
		if err != nil {
			return fmt.Errorf("could not create a new logger: %v", err)
		}
	} else {
		debugLog = zap.NewNop()
	}
	return
}
