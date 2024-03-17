package utils

import (
	"fmt"
	"os"
)

type Logger interface {
	Info(a ...any)
	Infof(format string, a ...any)
	Infoln(a ...any)

	Warn(a ...any)
	Warnf(format string, a ...any)
	Warnln(a ...any)

	Error(a ...any)
	Errorf(format string, a ...any)
	Errorln(a ...any)

	Debug(a ...any)
	Debugf(format string, a ...any)
	Debugln(a ...any)
}

// Logs to stdout and stderr.
type cliLogger struct {
	enableDebugLevel bool
}

// cliLogger logs to stdout and stderr.
func NewCLILogger(enableDebugLevel bool) Logger {
	return &cliLogger{enableDebugLevel: enableDebugLevel}
}

func (l *cliLogger) Info(a ...any) { fmt.Print(a...) }

func (l *cliLogger) Infof(format string, a ...any) { fmt.Printf(format, a...) }

func (l *cliLogger) Infoln(a ...any) { fmt.Println(a...) }

func (l *cliLogger) Warn(a ...any) { fmt.Print(a...) }

func (l *cliLogger) Warnf(format string, a ...any) { fmt.Printf(format, a...) }

func (l *cliLogger) Warnln(a ...any) { fmt.Println(a...) }

func (l *cliLogger) Error(a ...any) { fmt.Fprint(os.Stderr, a...) }

func (l *cliLogger) Errorf(format string, a ...any) { fmt.Fprintf(os.Stderr, format, a...) }

func (l *cliLogger) Errorln(a ...any) { fmt.Fprintln(os.Stderr, a...) }

func (l *cliLogger) Debug(a ...any) {
	if l.enableDebugLevel {
		a = append([]any{"debug:"}, a...)
		fmt.Fprint(os.Stderr, a...)
	}
}

func (l *cliLogger) Debugf(format string, a ...any) {
	if l.enableDebugLevel {
		fmt.Fprintf(os.Stderr, "debug: "+format, a...)
	}
}

func (l *cliLogger) Debugln(a ...any) {
	if l.enableDebugLevel {
		a = append([]any{"debug:"}, a...)
		fmt.Fprintln(os.Stderr, a...)
	}
}
