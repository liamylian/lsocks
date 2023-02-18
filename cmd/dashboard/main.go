package main

import (
	"github.com/liamylian/lsocks/internal/dashboard"
	"github.com/liamylian/lsocks/pkg/log"
	"github.com/liamylian/lsocks/pkg/types"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

var (
	httpPort, _  = types.EnvDefault("HTTP_PORT", "80").Int()
	logLevel     = types.EnvDefault("LOG_LEVEL", "info").String()
	logFile      = types.EnvDefault("LOG_FILE", "dashboard.log").String()
	trafficsFile = types.EnvDefault("TRAFFICS_FILE", "traffics.log").String()
)

func main() {
	configLog(logFile, logLevel)

	s := dashboard.NewStatistician(trafficsFile)
	go s.Run()
	waitSignal(syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
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
