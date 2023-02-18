package worker

import (
	"fmt"
	"time"

	"github.com/liamylian/lsocks/internal"
	"github.com/liamylian/lsocks/pkg/log"
	"github.com/liamylian/lsocks/pkg/proxy"
	"github.com/liamylian/lsocks/pkg/proxy/socks5"
)

type Worker struct {
	serverPort int
	server     *socks5.Server
}

func NewWorker(port int, credentials proxy.CredentialStore, trafficsFile string) (*Worker, error) {
	reporter, err := internal.NewTrafficsReporter(time.Minute, trafficsFile)
	if err != nil {
		return nil, err
	}

	conf := &socks5.Config{
		RequestReporter:  nil,
		ResponseReporter: reporter,
		Credentials:      credentials,
		Logger:           nil,
	}
	server, err := socks5.New(conf)
	if err != nil {
		return nil, err
	}

	return &Worker{
		serverPort: port,
		server:     server,
	}, nil
}

func (s *Worker) Serve() error {
	serveAddr := fmt.Sprintf(":%d", s.serverPort)
	if err := s.server.ListenAndServe("tcp", serveAddr); err != nil {
		log.WithError(err).Errorf("listen error: addr=%s", serveAddr)
		return err
	}

	return nil
}
