package proxy

// TrafficReporter 用于流量采集
type TrafficReporter interface {
	Report(userIdentifier string, trafficBytes int64) error
}
