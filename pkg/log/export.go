package log

import (
	glog "log"
	"os"
)

var defaultLogger = FromGolangLog(glog.New(os.Stderr, "", glog.LstdFlags), true)

func GetDefaultLogger() Logger {
	return defaultLogger
}

func SetDefaultLogger(l Logger) {
	defaultLogger = l
}

func Debug(args ...interface{}) {
	defaultLogger.Debug(args...)
}

func Info(args ...interface{}) {
	defaultLogger.Info(args...)
}

func Warn(args ...interface{}) {
	defaultLogger.Warn(args...)
}

func Error(args ...interface{}) {
	defaultLogger.Error(args...)
}

func Fatal(args ...interface{}) {
	defaultLogger.Fatal(args...)
}

func Debugf(format string, args ...interface{}) {
	defaultLogger.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}

func Warnf(format string, args ...interface{}) {
	defaultLogger.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	defaultLogger.Fatalf(format, args...)
}

func WithError(err error) Entry {
	return defaultLogger.WithError(err)
}

func WithField(key string, val interface{}) Entry {
	return defaultLogger.WithField(key, val)
}

func WithFields(fields map[string]interface{}) Entry {
	return defaultLogger.WithFields(fields)
}
