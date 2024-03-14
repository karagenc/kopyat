package utils

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
}
