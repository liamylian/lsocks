package log

import (
	glog "log"
)

// golangAdapter Golang默认日志适配器
type golangAdapter struct {
	gl *glog.Logger
}

// FromGolangLog 适配Golang默认日志
func FromGolangLog(gl *glog.Logger, reportCaller bool) Logger {
	l := &golangAdapter{
		gl: gl,
	}
	return AdapterLogger(l, reportCaller)
}

func (l golangAdapter) Debug(args ...interface{}) {
	args = append([]interface{}{"DEBUG"}, args...)
	l.gl.Println(args...)
}

func (l golangAdapter) Info(args ...interface{}) {
	args = append([]interface{}{"INFO"}, args...)
	l.gl.Println(args...)
}

func (l golangAdapter) Warn(args ...interface{}) {
	args = append([]interface{}{"WARN"}, args...)
	l.gl.Println(args...)
}

func (l golangAdapter) Error(args ...interface{}) {
	args = append([]interface{}{"ERROR"}, args...)
	l.gl.Println(args...)
}

func (l golangAdapter) Fatal(args ...interface{}) {
	args = append([]interface{}{"FATAL"}, args...)
	l.gl.Panicln(args...)
}
