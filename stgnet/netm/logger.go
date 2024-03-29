package netm

import "git.oschina.net/cloudzone/smartgo/stgcommon/logger"

// Debug logs a debug statement
func (bootstrap *Bootstrap) Debug(v ...interface{}) {
	logger.Debug(v...)
}

// Trace logs a trace statement
func (bootstrap *Bootstrap) Trace(v ...interface{}) {
	logger.Trace(v...)
}

// Notice logs a notice statement
func (bootstrap *Bootstrap) Notice(v ...interface{}) {
	logger.Info(v...)
}

// Warn logs an warn
func (bootstrap *Bootstrap) Warn(v ...interface{}) {
	logger.Warn(v...)
}

// Error logs an error
func (bootstrap *Bootstrap) Error(v ...interface{}) {
	logger.Error(v...)
}

// Fatal logs a fatal error
func (bootstrap *Bootstrap) Fatal(v ...interface{}) {
	logger.Fatal(v...)
}

// Debugf logs a debug statement
func (bootstrap *Bootstrap) Debugf(format string, v ...interface{}) {
	logger.Debugf(format, v...)
}

// Tracef logs a trace statement
func (bootstrap *Bootstrap) Tracef(format string, v ...interface{}) {
	logger.Tracef(format, v...)
}

// Noticef logs a notice statement
func (bootstrap *Bootstrap) Noticef(format string, v ...interface{}) {
	logger.Infof(format, v...)
}

// Warnf logs an warn
func (bootstrap *Bootstrap) Warnf(format string, v ...interface{}) {
	logger.Warnf(format, v...)
}

// Errorf logs an error
func (bootstrap *Bootstrap) Errorf(format string, v ...interface{}) {
	logger.Errorf(format, v...)
}

// Fatalf logs a fatal error
func (bootstrap *Bootstrap) Fatalf(format string, v ...interface{}) {
	logger.Fatalf(format, v...)
}

// LogFlush logs flush
func (bootstrap *Bootstrap) LogFlush() {
	logger.Flush()
}
