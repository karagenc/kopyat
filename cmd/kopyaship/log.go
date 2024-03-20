package main

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newLogger() (*zap.Logger, error) {
	level := zap.NewAtomicLevelAt(zap.DebugLevel)
	development := true

	outputPaths := []string{"stdout"}
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

	return logConfig.Build()
}
