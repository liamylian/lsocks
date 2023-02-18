package internal

import (
	"io"
	"time"

	"github.com/liamylian/lsocks/pkg/log"
)

const (
	trafficsReportBufSize = 1000
)

type trafficsEntry struct {
	identifier string
	bytes      int64
}

// trafficsCollection 非线程安全
type trafficsCollection map[string]int64

func (c trafficsCollection) Range(f func(identifier string, bytes int64)) {
	for i, b := range c {
		f(i, b)
	}
}

func (c trafficsCollection) Add(identifier string, bytes int64) {
	old, _ := c[identifier]
	c[identifier] = old + bytes
}

func (c trafficsCollection) Reset() {
	for k := range c {
		delete(c, k)
	}
}

// TrafficsReporter 流量采集器
// 因性能考虑，统计流量可能会少量丢失，并且流量产生的时间可能不精确
type TrafficsReporter struct {
	interval time.Duration      // 上报流量间隔
	traffics chan trafficsEntry // 上报流量管道
	writer   io.WriteCloser
}

func NewTrafficsReporter(interval time.Duration, filePath string) (*TrafficsReporter, error) {
	w, err := openTrafficsFile(filePath)
	if err != nil {
		return nil, err
	}

	c := &TrafficsReporter{
		interval: interval,
		traffics: make(chan trafficsEntry, trafficsReportBufSize),
		writer:   w,
	}

	go c.run()
	return c, nil
}

// Report 上报采集到的流量
func (c *TrafficsReporter) Report(identifier string, bytes int64) error {
	entry := trafficsEntry{identifier: identifier, bytes: bytes}
	select {
	case c.traffics <- entry:
		// 空操作
	default:
		// 缓冲区满了，将流量累加到未上报流量中
		log.Warnf("traffic report channel full")
	}

	return nil
}

func (c *TrafficsReporter) Close() error {
	return c.writer.Close()
}

func (c *TrafficsReporter) run() {
	currentPeriod := c.getPeriod(time.Now())    // 当前上报阶段
	currentTraffics := make(trafficsCollection) // 当前上报阶段流量
	for {
		select {
		case traffic := <-c.traffics:
			period := c.getPeriod(time.Now())
			if period == currentPeriod {
				// 累计当前阶段流量
				currentTraffics.Add(traffic.identifier, traffic.bytes)
			} else {
				// 下一阶段流量到来了，开始上报当前阶段流量
				currentTraffics.Range(func(identifier string, bytes int64) {
					if bytes > 0 {
						c.logTraffics(currentPeriod, identifier, bytes)
					}
				})

				currentPeriod = period
				currentTraffics.Reset()
			}
		}
	}
}

func (c *TrafficsReporter) getPeriod(t time.Time) time.Time {
	nano := t.UnixNano() - t.UnixNano()%int64(c.interval)
	return time.Unix(nano/int64(time.Second), nano%int64(time.Second))
}

func (c *TrafficsReporter) logTraffics(period time.Time, identifier string, traffics int64) {
	writeTraffics(c.writer, period, identifier, traffics)
}
