## Graceful Shutdown

This Golang package provides a graceful way to close application resources when the operating system requests to kill processes started by your application.

This is highly recommended for applications that are always running with critical components like **Database Connections** or **Message Brokers**. If those components implement the **gracefulshutdown.Closeable** interface, then you can finish the process gracefully ;)

## Technologies

- Golang 1.22

### Contracts

```go
// To close resource before os.Exit
type Closeable interface {
	Close() error
}

// to log gracefulshutdown internal steps
type Logger interface {
	Info(string, ...any)
	Warn(string, ...any)
	Error(string, ...any)
}
```

Basically, this API offers two methods that will monitor and handle operating system terminate signals:

### Handle(logger Logger, closeable ...Closeable)

This method receives a **Logger** implementation and one or more **Closeable** structs. When (or if) syscalls occur, the method will execute all **Close** functions and then return control to the original thread. If any **Close** call encounters an error, the **Handle** function will print an error using the **logger.Error** provided earlier:

```go
package main

import (
	"errors"
	"log/slog"
	"os"

	"github.com/eviccari/graceful-shutdown/gracefulshutdown"
)

// could be SQL Connection, Kafka Client or anything that you want ;)
type FakeDB struct {}

func (fDB *FakeDB) Close() error {
	return nil
}

func main() {
	// get gracefulshutdown.Logger implementation
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.Level(0),
	}))
	logger.Info("MY_APP", "pID", os.Getpid(), "ppID", os.Getppid())

	// get a gracefulshutdown.Closeable implementation
	fDB := &FakeDB{}

	// inject the fDB into your awesome component
    go NewBusinessLogic(fDB).Execute()

	terminate := make(chan bool, 1) // to hold the flow after OS syscall requirements
	go func() {
		<-gracefulshutdown.Handle(logger, fDB) // to monitor and handle OS syscalls
		terminate <- true
	}()
	<-terminate // you will get the control again
	logger.Info("MY_APP", "status", "terminated")
	os.Exit(0)
}

```

Run the program, wait a second, and then press **Ctrl + c**. You will see the **ˆC** symbol at the beginning of the second log line and all lines printed by the gracefulshutdown API (excluding **"msg": "MY_APP"**).:

```bash
❯ go run main.go
{"time":"2024-07-17T15:21:12.116042-03:00","level":"INFO","msg":"MY_APP","pID":41233,"ppID":41217}
^C{"time":"2024-07-17T15:21:15.153878-03:00","level":"WARN","msg":"system call receipt -> interrupt"}
{"time":"2024-07-17T15:21:15.153911-03:00","level":"INFO","msg":"closing resources..."}
{"time":"2024-07-17T15:21:15.153918-03:00","level":"INFO","msg":"trying to close resource 0"}
{"time":"2024-07-17T15:21:15.153921-03:00","level":"WARN","msg":"system was terminated by system call"}
{"time":"2024-07-17T15:21:15.153934-03:00","level":"INFO","msg":"MY_APP","status":"terminated"}
```

### HandleAndTerminate(logger Logger, closeable ...Closeable)

This method receives a **Logger** implementation and one or more **Closeable** structs in the same way that the Handle method does. Essentially, the API will execute **os.Exit(1)** after all Close methods have finished:

```go
/////////////////////
// same code above //
/////////////////////

func main() {
	// get gracefulshutdown.Logger implementation
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.Level(0),
	}))
	logger.Info("MY_APP", "pID", os.Getpid(), "ppID", os.Getppid())

	// get a gracefulshutdown.Closeable implementation
	fDB := &FakeDB{}

    // delegates Close methods and os.Exit(0) to the API
	go gracefulshutdown.HandleAndTerminate(logger, fDB)

	// inject the fDB into your awesome component
    NewBusinessLogic(fDB).Execute()
	logger.Info("MY_APP", "status", "terminated")
	os.Exit(0)
}

```

Thank you! Enjoy!
