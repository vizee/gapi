package log

type Logger interface {
	Debugf(format string, args ...any)
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
}

var exported Logger = &nopLogger{}

func SetLogger(logger Logger) {
	exported = logger
}

func Debugf(format string, args ...any) {
	exported.Debugf(format, args...)
}

func Warnf(format string, args ...any) {
	exported.Warnf(format, args...)
}

func Errorf(format string, args ...any) {
	exported.Errorf(format, args...)
}

type nopLogger struct{}

func (*nopLogger) Debugf(format string, args ...any) {
}

func (*nopLogger) Warnf(format string, args ...any) {
}

func (*nopLogger) Errorf(string, ...any) {
}
