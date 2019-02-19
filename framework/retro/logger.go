package retro

// Logger is the generic logging interface. It explicitly avoids including
// Fatal and Fatalf because of the relative brutal nature of os.Exit
// without a chance to clean up.
//
// In general tracing should be preferred to logging, however logging can
// always be valuable.
type Logger interface {
	Debug(...interface{})
	Debugf(string, ...interface{})
	Info(...interface{})
	Infof(string, ...interface{})
	Warn(...interface{})
	Warnf(string, ...interface{})
	Error(...interface{})
	Errorf(string, ...interface{})
}
