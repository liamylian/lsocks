package main

import (
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/liamylian/lsocks/internal/worker"
	"github.com/liamylian/lsocks/pkg/log"
	"github.com/liamylian/lsocks/pkg/proxy"
	"github.com/liamylian/lsocks/pkg/types"
)

var (
	socksPort, _ = types.EnvDefault("SOCKS_PORT", "9080").Int()
	logLevel     = types.EnvDefault("LOG_LEVEL", "info").String()
	logFile      = types.EnvDefault("LOG_FILE", "worker.log").String()
	trafficsFile = types.EnvDefault("TRAFFICS_FILE", "traffics.log").String()
	credentials  = types.Env("CREDENTIALS").StringArray()
)

func main() {
	credentialStore := makeCredentialStore()

	server, err := worker.NewWorker(socksPort, credentialStore, trafficsFile)
	if err != nil {
		log.WithError(err).Errorf("new socks: port=%d", socksPort)
	}
	go func() {
		if err := server.Serve(); err != nil {
			log.WithError(err).Errorf("serve %d failed", socksPort)
		}
	}()

	configLog(logFile, logLevel)
	waitSignal(syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
}

func makeCredentialStore() proxy.CredentialStore {
	store := make(proxy.StaticCredentials)
	for _, c := range credentials {
		splits := strings.Split(c, "/")
		if len(splits) != 2 {
			continue
		}
		user := splits[0]
		pass := splits[1]
		if user == "" || pass == "" {
			continue
		}
		store[user] = pass
	}
	if len(store) == 0 {
		return nil
	}

	return store
}

func configLog(logFilePath, level string) {
	out := os.Stdout
	if logFilePath != "" {
		var err error
		out, err = os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			log.WithError(err).Fatalf("failed to open file: %s", logFilePath)
		}
	}

	l := logrus.New()
	if lv, err := logrus.ParseLevel(level); err == nil {
		l.SetLevel(lv)
	} else {
		log.WithError(err).Errorf("parse level failed: level=%s", level)
	}
	l.SetOutput(out)
	l.SetReportCaller(true)
	log.SetDefaultLogger(log.FromLogrus(l, true))
}

func waitSignal(signals ...os.Signal) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, signals...)
	<-signalChan
}
