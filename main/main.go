package main

import (
	"flag"
	"fmt"
	"github.com/weishi258/http-log-collector/log"
	"github.com/weishi258/http-log-collector/rest"
	"github.com/weishi258/http-log-collector/rest/handlers"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

const ServerName = "Http Log Collector"

func main() {
	sigChan := make(chan os.Signal, 1)
	done := make(chan bool, 1)
	resetServerChan := make(chan bool, 1)
	signal.Notify(sigChan,
		syscall.SIGTERM,
		syscall.SIGINT)

	var localPort int
	var logFile string
	var logLevel string
	flag.StringVar(&logFile, "log", "", "log output file path")
	flag.StringVar(&logLevel, "l", "info", "log level")
	flag.IntVar(&localPort, "port", 8000, "rest listening port")
	flag.Parse()

	var err error
	defer func() {
		if err != nil {
			os.Exit(1)
		} else {
			os.Exit(0)
		}
	}()

	logger := log.InitLogger(logFile, logLevel, false)
	server := rest.NewRestServer(fmt.Sprintf("0.0.0.0:%d", localPort))
	server.AddRoutes(handlers.GetLogRoutes())
	if err = server.Start(resetServerChan, false); err != nil {
		log.GetLogger().Fatal(fmt.Sprintf("start %s failed", ServerName), zap.String("error", err.Error()))
		return
	}

	logger.Info(fmt.Sprintf("%s start successful", ServerName))
	go func() {
		select {
		case sig := <-sigChan:
			logger.Debug("caught signal for exit", zap.Any("signal", sig))

			done <- true
		case <-resetServerChan:
			logger.Fatal(fmt.Sprintf("%s crashed, quiting", ServerName))
			done <- true
		}

	}()
	<-done
	logger.Info(fmt.Sprintf("%s quited", ServerName))
}
