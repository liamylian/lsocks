package log

import (
	"fmt"
	"runtime"
	"strings"
	"sync"
)

const (
	maximumCallerDepth int = 25
	knownLogrusFrames  int = 3
)

var (
	// qualified package name, cached at first use
	logPackage string

	// Positions in the call stack when tracing to report the calling method
	minimumCallerDepth int

	// Used for caller information initialisation
	callerInitOnce sync.Once
)

// Entry 日志
type Entry interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})
	WithError(err error) Entry
	WithField(key string, val interface{}) Entry
	WithFields(fields map[string]interface{}) Entry
}

// entry 日志项
type entry struct {
	be           FieldAdapterEntry
	reportCaller bool
}

func (e entry) Debug(args ...interface{}) {
	e.setCaller()
	e.be.DebugFunc(e.be.Fields, args...)
}

func (e entry) Info(args ...interface{}) {
	e.setCaller()
	e.be.InfoFunc(e.be.Fields, args...)
}

func (e entry) Warn(args ...interface{}) {
	e.setCaller()
	e.be.WarnFunc(e.be.Fields, args...)
}

func (e entry) Error(args ...interface{}) {
	e.setCaller()
	e.be.ErrorFunc(e.be.Fields, args...)
}

func (e entry) Fatal(args ...interface{}) {
	e.setCaller()
	e.be.FatalFunc(e.be.Fields, args...)
}

func (e entry) Debugf(format string, args ...interface{}) {
	e.setCaller()
	e.be.DebugFunc(e.be.Fields, fmt.Sprintf(format, args...))
}

func (e entry) Infof(format string, args ...interface{}) {
	e.setCaller()
	e.be.InfoFunc(e.be.Fields, fmt.Sprintf(format, args...))
}

func (e entry) Warnf(format string, args ...interface{}) {
	e.setCaller()
	e.be.WarnFunc(e.be.Fields, fmt.Sprintf(format, args...))
}

func (e entry) Errorf(format string, args ...interface{}) {
	e.setCaller()
	e.be.ErrorFunc(e.be.Fields, fmt.Sprintf(format, args...))
}

func (e entry) Fatalf(format string, args ...interface{}) {
	e.setCaller()
	e.be.FatalFunc(e.be.Fields, fmt.Sprintf(format, args...))
}

func (e *entry) WithError(err error) Entry {
	if e.be.Fields == nil {
		e.be.Fields = make(map[string]interface{})
	}

	e.be.Fields["error"] = err
	return e
}

func (e *entry) WithField(key string, val interface{}) Entry {
	if e.be.Fields == nil {
		e.be.Fields = make(map[string]interface{})
	}

	e.be.Fields[key] = val
	return e
}

func (e *entry) WithFields(fields map[string]interface{}) Entry {
	if e.be.Fields == nil {
		e.be.Fields = make(map[string]interface{})
	}

	for k, v := range fields {
		e.be.Fields[k] = v
	}
	return e
}

func (e *entry) setCaller() {
	if e.be.Fields == nil {
		e.be.Fields = make(map[string]interface{})
	}

	if e.reportCaller {
		caller := getCaller()
		e.be.Fields["file"] = fmt.Sprintf("%s:%d", caller.File, caller.Line)
	}
}

// Logger 日志
type Logger interface {
	Entry

	// TODO
}

// logger 实现
type logger struct {
	fl           FieldAdapter
	reportCaller bool
}

// AdapterLogger 将基础日志转换为日志
func AdapterLogger(basic BasicAdapter, reportCaller bool) Logger {
	if fl, ok := basic.(FieldAdapter); ok {
		return &logger{
			fl:           fl,
			reportCaller: reportCaller,
		}
	}

	fl := toFieldAdapter(basic)
	return &logger{
		fl:           fl,
		reportCaller: reportCaller,
	}
}

func (l logger) Debug(args ...interface{}) {
	l.entry().Debug(args...)
}

func (l logger) Info(args ...interface{}) {
	l.entry().Info(args...)
}

func (l logger) Warn(args ...interface{}) {
	l.entry().Warn(args...)
}

func (l logger) Error(args ...interface{}) {
	l.entry().Error(args...)
}

func (l logger) Fatal(args ...interface{}) {
	l.entry().Fatal(args...)
}

func (l logger) Debugf(format string, args ...interface{}) {
	l.entry().Debug(fmt.Sprintf(format, args...))
}

func (l logger) Infof(format string, args ...interface{}) {
	l.entry().Info(fmt.Sprintf(format, args...))
}

func (l logger) Warnf(format string, args ...interface{}) {
	l.entry().Warn(fmt.Sprintf(format, args...))
}

func (l logger) Errorf(format string, args ...interface{}) {
	l.entry().Error(fmt.Sprintf(format, args...))
}

func (l logger) Fatalf(format string, args ...interface{}) {
	l.entry().Fatal(fmt.Sprintf(format, args...))
}

func (l logger) WithError(err error) Entry {
	return &entry{
		be:           l.fl.WithError(err),
		reportCaller: l.reportCaller,
	}
}

func (l logger) WithField(key string, val interface{}) Entry {
	return &entry{
		be:           l.fl.WithField(key, val),
		reportCaller: l.reportCaller,
	}
}

func (l logger) WithFields(fields map[string]interface{}) Entry {
	return &entry{
		be:           l.fl.WithFields(fields),
		reportCaller: l.reportCaller,
	}
}

func (l logger) entry() entry {
	return entry{
		be:           l.fl.WithFields(nil),
		reportCaller: l.reportCaller,
	}
}

// getCaller retrieves the name of the first non-logrus calling function
func getCaller() *runtime.Frame {
	// cache this package's fully-qualified name
	callerInitOnce.Do(func() {
		pcs := make([]uintptr, maximumCallerDepth)
		_ = runtime.Callers(0, pcs)

		// dynamic get the package name and the minimum caller depth
		for i := 0; i < maximumCallerDepth; i++ {
			funcName := runtime.FuncForPC(pcs[i]).Name()
			if strings.Contains(funcName, "getCaller") {
				logPackage = getPackageName(funcName)
				break
			}
		}

		minimumCallerDepth = knownLogrusFrames
	})

	// Restrict the lookback frames to avoid runaway lookups
	pcs := make([]uintptr, maximumCallerDepth)
	depth := runtime.Callers(minimumCallerDepth, pcs)
	frames := runtime.CallersFrames(pcs[:depth])

	for f, again := frames.Next(); again; f, again = frames.Next() {
		pkg := getPackageName(f.Function)

		// If the caller isn't part of this package, we're done
		if pkg != logPackage {
			return &f //nolint:scopelint
		}
	}

	// if we got here, we failed to find the caller's context
	return nil
}

// getPackageName reduces a fully qualified function name to the package name
// There really ought to be to be a better way...
func getPackageName(f string) string {
	for {
		lastPeriod := strings.LastIndex(f, ".")
		lastSlash := strings.LastIndex(f, "/")
		if lastPeriod > lastSlash {
			f = f[:lastPeriod]
		} else {
			break
		}
	}

	return f
}
