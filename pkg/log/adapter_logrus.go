package log

import (
	"github.com/sirupsen/logrus"
)

// logrusAdapter Logrus日志适配器
type logrusAdapter struct {
	logrus *logrus.Logger
}

// FromLogrus 适配Logrus
func FromLogrus(logurs *logrus.Logger, reportCaller bool) Logger {
	logrus.SetReportCaller(false) // 禁用 logrus ReportCaller
	l := &logrusAdapter{logrus: logurs}
	return AdapterLogger(l, reportCaller)
}

func (l logrusAdapter) Debug(args ...interface{}) {
	l.logrus.Debug(args...)
}

func (l logrusAdapter) Info(args ...interface{}) {
	l.logrus.Info(args...)
}

func (l logrusAdapter) Warn(args ...interface{}) {
	l.logrus.Warn(args...)
}

func (l logrusAdapter) Error(args ...interface{}) {
	l.logrus.Error(args...)
}

func (l logrusAdapter) Fatal(args ...interface{}) {
	l.logrus.Fatal(args...)
}

func (l logrusAdapter) WithError(err error) FieldAdapterEntry {
	return l.WithFields(map[string]interface{}{"error": err})
}

func (l logrusAdapter) WithField(key string, val interface{}) FieldAdapterEntry {
	return l.WithFields(map[string]interface{}{key: val})
}

func (l logrusAdapter) WithFields(fields map[string]interface{}) FieldAdapterEntry {
	return FieldAdapterEntry{
		Fields: fields,
		DebugFunc: func(fields map[string]interface{}, args ...interface{}) {
			l.logrus.WithFields(fields).Debug(args...)
		},
		InfoFunc: func(fields map[string]interface{}, args ...interface{}) {
			l.logrus.WithFields(fields).Info(args...)
		},
		WarnFunc: func(fields map[string]interface{}, args ...interface{}) {
			l.logrus.WithFields(fields).Warn(args...)
		},
		ErrorFunc: func(fields map[string]interface{}, args ...interface{}) {
			l.logrus.WithFields(fields).Error(args...)
		},
		FatalFunc: func(fields map[string]interface{}, args ...interface{}) {
			l.logrus.WithFields(fields).Fatal(args...)
		},
	}
}
