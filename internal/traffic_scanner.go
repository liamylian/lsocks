package internal

import (
	"bufio"
	"context"
	"io"
	"os"
	"strings"
	"time"
)

func GetCurrentTrafficsFile(base string) string {
	return getCurrentTrafficsFile(base)
}

func ListTrafficsFiles(dir string, base string) ([]string, error) {
	return listTrafficsFiles(dir, base)
}

type TrafficsScanner struct {
	file *os.File
}

func NewTrafficsScanner(filePath string) (*TrafficsScanner, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	return &TrafficsScanner{
		file: file,
	}, nil
}

// Scan 扫描数据
func (s *TrafficsScanner) Scan(ctx context.Context, f func(time time.Time, identifier string, bytes int64)) error {
	defer s.file.Close()
	return scanLine(ctx, s.file, func(line string) (continue_ bool) {
		t, i, b, err := readTrafficsLine(line)
		if err != nil {
			return false
		}
		f(t, i, b)
		return true
	})
}

// Tail 扫描数据，不支持日志轮询场景（即文件被重命名归档，然后创建一个新文件用于使用）
func (s *TrafficsScanner) Tail(ctx context.Context, f func(time time.Time, identifier string, bytes int64)) error {
	defer s.file.Close()
	return tailLine(ctx, s.file, func(line string) (continue_ bool) {
		t, i, b, err := readTrafficsLine(line)
		if err != nil {
			return false
		}
		f(t, i, b)
		return true
	})
}

func scanLine(ctx context.Context, file *os.File, f func(line string) (continue_ bool)) error {
	reader := bufio.NewReader(file)
	for {
		// 如果上下文已结束，则退出
		select {
		case <-ctx.Done():
			return nil
		default:
			// 继续执行
		}

		if line, err := reader.ReadString('\n'); err == nil {
			line = strings.TrimSpace(line) // 换行符可能为 \r\n
			if f(line) {
				continue
			} else {
				return nil
			}
		} else if err == io.EOF {
			return nil
		} else {
			return err
		}
	}
}

func tailLine(ctx context.Context, file *os.File, f func(line string) (continue_ bool)) error {
	reader := bufio.NewReader(file)
	for {
		// 如果上下文已结束，则退出
		select {
		case <-ctx.Done():
			return nil
		default:
			// 继续执行
		}

		if line, err := reader.ReadString('\n'); err == nil {
			line = strings.TrimSpace(line) // 换行符可能为 \r\n
			if f(line) {
				continue
			} else {
				return nil
			}
		} else if err == io.EOF {
			// 需要一直监听文件，则先睡眠（避免 CPU 飙升），再检查是否有新的数据，或者文件被截取
			time.Sleep(500 * time.Millisecond)
			// 如果文件被截取，则从头读取
			if truncated, err := isTruncated(file); err != nil {
				return err
			} else {
				if truncated {
					// 文件被截取，从头读取
					if _, err := file.Seek(0, io.SeekStart); err != nil {
						return err
					}
				}
			}
		} else {
			return err
		}
	}
}

func isTruncated(file *os.File) (bool, error) {
	currentPos, err := file.Seek(0, io.SeekCurrent)
	if err != nil {
		return false, err
	}
	fileInfo, err := file.Stat()
	if err != nil {
		return false, err
	}
	return currentPos > fileInfo.Size(), nil
}
