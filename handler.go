package gracefulshutdown

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

type Closeable interface {
	Close() error
}

type Logger interface {
	Info(string, ...any)
	Warn(string, ...any)
	Error(string, ...any)
}

func Handle(logger Logger, closeable ...Closeable) <-chan bool {
	do(logger, closeable...)
	terminated := make(chan bool, 1)
	defer close(terminated)
	terminated <- true
	return terminated
}

func HandleAndTerminate(logger Logger, closeable ...Closeable) {
	do(logger, closeable...)
	os.Exit(0)
}

func do(logger Logger, closeable ...Closeable) {
	osSignals := make(chan os.Signal, 1)
	terminate := make(chan bool, 1)
	signal.Notify(osSignals, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGQUIT, syscall.SIGINT)
	go func() {
		osSignal := <-osSignals
		logger.Warn(fmt.Sprintf("system call receipt -> %v", osSignal))
		terminate <- true
	}()
	<-terminate
	logger.Info("closing resources...")
	for i, c := range closeable {
		logger.Info(fmt.Sprintf("trying to close resource %d", i))
		if err := c.Close(); err != nil {
			logger.Error(fmt.Sprintf("error on close resource: %s", err.Error()))
		}
	}
	logger.Warn("system was terminated by system call")
}
