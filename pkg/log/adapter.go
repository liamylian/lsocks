package log

import (
	"fmt"
	"strings"
)

// BasicAdapter 基础日志适配器，用于接入日志系统
type BasicAdapter interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
}

// FieldAdapterEntry 日志项适配器，用于接入以支持额外字段的日志系统
type FieldAdapterEntry struct {
	Fields map[string]interface{}

	DebugFunc func(fields map[string]interface{}, args ...interface{})
	InfoFunc  func(fields map[string]interface{}, args ...interface{})
	WarnFunc  func(fields map[string]interface{}, args ...interface{})
	ErrorFunc func(fields map[string]interface{}, args ...interface{})
	FatalFunc func(fields map[string]interface{}, args ...interface{})
}

func (e FieldAdapterEntry) Debug(args ...interface{}) {
	e.DebugFunc(e.Fields, args...)
}

func (e FieldAdapterEntry) Info(args ...interface{}) {
	e.InfoFunc(e.Fields, args...)
}

func (e FieldAdapterEntry) Warn(args ...interface{}) {
	e.WarnFunc(e.Fields, args...)
}

func (e FieldAdapterEntry) Error(args ...interface{}) {
	e.ErrorFunc(e.Fields, args...)
}

func (e FieldAdapterEntry) Fatal(args ...interface{}) {
	e.FatalFunc(e.Fields, args...)
}

func (e FieldAdapterEntry) WithError(err error) FieldAdapterEntry {
	return e.WithFields(map[string]interface{}{"error": err})
}

func (e FieldAdapterEntry) WithField(key string, val interface{}) FieldAdapterEntry {
	return e.WithFields(map[string]interface{}{key: val})
}

func (e FieldAdapterEntry) WithFields(fields map[string]interface{}) FieldAdapterEntry {
	clone := e
	clone.Fields = make(map[string]interface{})

	for k, v := range e.Fields {
		clone.Fields[k] = v
	}
	for k, v := range fields {
		clone.Fields[k] = v
	}
	return clone
}

// FieldAdapter 日志适配器，用于接入以支持额外字段的日志系统
type FieldAdapter interface {
	BasicAdapter
	WithError(err error) FieldAdapterEntry
	WithField(key string, val interface{}) FieldAdapterEntry
	WithFields(fields map[string]interface{}) FieldAdapterEntry
}

// fieldAdapter 支持额外字段的日志实现
type fieldAdapter struct {
	BasicAdapter
}

// toFieldAdapter 将基础日志转换为支持额外字段的日志
func toFieldAdapter(l BasicAdapter) FieldAdapter {
	return &fieldAdapter{
		BasicAdapter: l,
	}
}

func (l fieldAdapter) WithError(err error) FieldAdapterEntry {
	return l.WithFields(map[string]interface{}{"error": err})
}

func (l fieldAdapter) WithField(key string, val interface{}) FieldAdapterEntry {
	return l.WithFields(map[string]interface{}{key: val})
}

func (l fieldAdapter) WithFields(fields map[string]interface{}) FieldAdapterEntry {
	getAllArgs := func(fields map[string]interface{}, args []interface{}) []interface{} {
		if len(fields) == 0 {
			return args
		}

		fieldArgs := make([]string, 0, len(fields))
		for k, v := range fields {
			fieldArgs = append(fieldArgs, fmt.Sprintf("%s: %v", k, v))
		}
		fieldArg := "[ " + strings.Join(fieldArgs, ", ") + " ]"
		return append(args, fieldArg)
	}
	return FieldAdapterEntry{
		Fields: fields,
		DebugFunc: func(fields map[string]interface{}, args ...interface{}) {
			allArgs := getAllArgs(fields, args)
			l.BasicAdapter.Debug(allArgs...)
		},
		InfoFunc: func(fields map[string]interface{}, args ...interface{}) {
			allArgs := getAllArgs(fields, args)
			l.BasicAdapter.Info(allArgs...)
		},
		WarnFunc: func(fields map[string]interface{}, args ...interface{}) {
			allArgs := getAllArgs(fields, args)
			l.BasicAdapter.Warn(allArgs...)
		},
		ErrorFunc: func(fields map[string]interface{}, args ...interface{}) {
			allArgs := getAllArgs(fields, args)
			l.BasicAdapter.Error(allArgs...)
		},
		FatalFunc: func(fields map[string]interface{}, args ...interface{}) {
			allArgs := getAllArgs(fields, args)
			l.BasicAdapter.Fatal(allArgs...)
		},
	}
}
