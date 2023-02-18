package log

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type rotateWriter struct {
	basePath   string
	timeFormat string
	filePath   string
	file       *os.File
}

func RotateFilePath(basePath string, timeFormat string, t time.Time) string {
	ext := filepath.Ext(basePath)
	suffix := "-" + t.Format(timeFormat)
	if ext != "" {
		return strings.Replace(basePath, ext, suffix+ext, 1)
	} else {
		return basePath + suffix
	}
}

func OpenRotateWriter(basePath string, timeFormat string) (io.WriteCloser, error) {
	filePath := RotateFilePath(basePath, timeFormat, time.Now())
	f, err := openFile(filePath)
	if err != nil {
		return nil, err
	}

	return &rotateWriter{
		basePath:   basePath,
		filePath:   filePath,
		timeFormat: timeFormat,
		file:       f,
	}, nil
}

// Write 不支持并发写，请自行保证
func (l *rotateWriter) Write(p []byte) (int, error) {
	filePath := RotateFilePath(l.basePath, l.timeFormat, time.Now())
	if filePath != l.filePath {
		if file, err := openFile(filePath); err == nil {
			_ = l.file.Close()
			l.filePath = filePath
			l.file = file
		}
	}

	return l.file.Write(p)
}

func (l *rotateWriter) Close() error {
	return l.file.Close()
}

func openFile(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
}
