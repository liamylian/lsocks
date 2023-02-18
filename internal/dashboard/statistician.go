package dashboard

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/liamylian/lsocks/internal"
	"github.com/liamylian/lsocks/pkg/log"
)

type Statistician struct {
	dir     string
	base    string
	current string
	storage Storage
}

func NewStatistician(trafficsFile string, storage Storage) *Statistician {
	trafficsDir := filepath.Dir(trafficsFile)
	trafficsBase := filepath.Base(trafficsFile)
	return &Statistician{
		dir:     trafficsDir,
		base:    trafficsBase,
		storage: storage,
	}
}

func (s *Statistician) Run() {
	s.recordHistory(context.Background())

	cancel := func() {}
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			current := internal.GetCurrentTrafficsFile(s.base)
			if current == s.current {
				log.Debugf("statistician: skip, same current file")
				continue
			}
			if _, err := os.Stat(current); err != nil {
				log.WithError(err).Debugf("statistician: skip, current file not exist: %s", current)
				continue
			}

			cancel()

			var ctx context.Context
			ctx, cancel = context.WithCancel(context.Background())
			s.current = current
			go s.recordLatest(ctx)
			log.Infof("statistician: start recordLatest, current=%s", s.current)
		}
	}
}

func (s *Statistician) recordLatest(ctx context.Context) {
	scanner, err := internal.NewTrafficsScanner(s.current)
	if err != nil {
		log.WithError(err).Errorf("statistician: fail to create scanner, current=%s", s.current)
		return
	}

	err = scanner.Tail(ctx, func(time time.Time, identifier string, bytes int64) {
		s.record(time, identifier, bytes)
	})
	if err != nil {
		log.WithError(err).Errorf("statistician: fail to tail, current=%s", s.current)
		return
	}
}

func (s *Statistician) recordHistory(ctx context.Context) {
	current := internal.GetCurrentTrafficsFile(s.base)
	files, err := internal.ListTrafficsFiles(s.dir, s.base)
	if err != nil {
		return
	}

	for _, file := range files {
		if file == current {
			continue
		}

		scanner, err := internal.NewTrafficsScanner(file)
		if err != nil {
			log.WithError(err).Errorf("statistician: fail to create scanner, current=%s", s.current)
			continue
		}

		err = scanner.Scan(ctx, func(time time.Time, identifier string, bytes int64) {
			s.record(time, identifier, bytes)
		})
		if err != nil {
			log.WithError(err).Errorf("statistician: fail to scan, current=%s", s.current)
			continue
		}
	}
}

func (s *Statistician) record(time time.Time, identifier string, bytes int64) {
	record := &Record{
		Identifier: identifier,
		Bytes:      bytes,
		Time:       time,
	}
	if err := s.storage.Put(record); err != nil {
		log.WithError(err).Errorf("statistician: put record failed, identifier=%s, bytes=%d, time=%s", identifier, bytes, time)
	}
}
