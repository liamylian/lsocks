package internal

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/liamylian/lsocks/pkg/log"
)

const (
	trafficsRotateFileTimeFormat = "20060102"
	trafficsRecordTimeFormat     = "20060102150405"
)

// 打开文件，用于记录使用流量
func openTrafficsFile(base string) (io.WriteCloser, error) {
	return log.OpenRotateWriter(base, trafficsRotateFileTimeFormat)
}

// 获取当前记录流量数据的文件名
func getCurrentTrafficsFile(base string) string {
	return log.RotateFilePath(base, trafficsRotateFileTimeFormat, time.Now())
}

// 列出记录流量数据的文件列表
func listTrafficsFiles(dir string, base string) ([]string, error) {
	ext := filepath.Ext(base)
	exp := strings.Replace(base, ext, `-\d{8}`+ext+"$", 1)
	var listRE = regexp.MustCompile(exp)

	var files []string
	err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
		if info.IsDir() {
			return nil // 跳过
		}

		if listRE.MatchString(path) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i] < files[j]
	})
	return files, nil
}

// writeTraffics 记录流量数据
// 流量数据格式： 20230215165500 admin 423469
func writeTraffics(writer io.Writer, period time.Time, identifier string, traffics int64) {
	entry := fmt.Sprintf("%s %s %d\n", period.Format(trafficsRecordTimeFormat), identifier, traffics)
	if _, err := writer.Write([]byte(entry)); err != nil {
		log.WithError(err).Errorf("log traffics failed")
	}
}

// readTrafficsLine 读取流量数据
// 流量数据格式： 20230215165500 admin 423469
func readTrafficsLine(line string) (t time.Time, identifier string, bytes int64, err error) {
	splits := strings.Split(line, " ")
	if len(splits) != 3 {
		err = errors.New("bad traffics format")
		return
	}

	if t, err = time.ParseInLocation(trafficsRecordTimeFormat, splits[0], time.Local); err != nil {
		return
	}
	identifier = splits[1]
	if bytes, err = strconv.ParseInt(splits[2], 10, 64); err != nil {
		return
	}
	return
}
