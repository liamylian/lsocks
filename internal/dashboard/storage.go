package dashboard

import (
	"sort"
	"sync"
	"time"
)

type Record struct {
	Identifier string    `json:"identifier"`
	Bytes      int64     `json:"bytes"`
	Time       time.Time `json:"time"`
}

type Storage interface {
	Put(record *Record) error
	List(identifier string, interval time.Duration, begin time.Time, end time.Time) (records []*Record, err error)
}

type staticStorage struct {
	recordsMu sync.RWMutex
	records   map[string][]*Record // identifier => records
}

func NewStaticStorage() Storage {
	return &staticStorage{
		records: make(map[string][]*Record),
	}
}

// Put 插入统计数据，时间必须递增，否则将引起 List 返回结果乱序
func (s *staticStorage) Put(record *Record) error {
	s.recordsMu.Lock()
	defer s.recordsMu.Unlock()

	if _, ok := s.records[record.Identifier]; !ok {
		s.records[record.Identifier] = []*Record{}
	}
	s.records[record.Identifier] = append(s.records[record.Identifier], record)

	return nil
}

func (s *staticStorage) List(identifier string, interval time.Duration, begin time.Time, end time.Time) (records []*Record, err error) {
	s.recordsMu.RLock()
	defer s.recordsMu.RUnlock()

	list, ok := s.records[identifier]
	if !ok {
		return nil, nil
	}
	if len(list) == 0 {
		return nil, nil
	}

	var beginIndex, endIndex int
	if begin.Before(list[0].Time) {
		beginIndex = 0
	} else {
		beginIndex = sort.Search(len(list), func(i int) bool {
			return list[i].Time.Equal(begin) || list[i].Time.After(begin)
		})
	}
	endIndex = sort.Search(len(list), func(i int) bool {
		return list[i].Time.Equal(end) || list[i].Time.After(end)
	})

	if beginIndex == endIndex {
		return nil, nil
	} else {
		return list[beginIndex:endIndex], nil
	}
}
