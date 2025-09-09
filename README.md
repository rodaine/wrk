# wrk

`wrk` is a utility for running multiple workers, such as web servers (HTTP/gRPC/etc) as well as daemon-style goroutines
in a coordinated fashion. On any errors, all workers are gracefully stopped either through context-cancellation or by
executing a worker's defined Stop function.

## Usage

To import this module into a Go project, use `go get`:

```shell
$ go get -u github.com/rodaine/wrk
```

### Example

```go
package main

import (
	"context"
	"log"
	"net/http"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rodaine/wrk"
)

func main() {
	// captures any signals to initiate a graceful shutdown
	sigs := wrk.Signals{
		syscall.SIGINT,
		syscall.SIGTERM,
	}

	// launches the http.DefaultServeMux on port 8080
	web := &wrk.HTTPServer{
		Server:              &http.Server{Addr: ":8080"},
		StopTimeout:         10 * time.Second,
		OverrideBaseContext: true,
	}

	// launches the prometheus metrics on port 2112
	metrics := &wrk.HTTPServer{
		Server: &http.Server{
			Addr:    ":2112",
			Handler: promhttp.Handler(),
		},
	}

	// spawns a daemon worker
	daemon := wrk.Named{
		Name:     "my daemon",
		Delegate: wrk.WorkerFunc(MyDaemon),
	} 

	// blocks until all workers are done
	err := wrk.Work(context.Background(), 
		sigs, web, metrics, daemon)
	
	log.Println(err)
}

func MyDaemon(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(time.Second):
			log.Println("tick")
		}
	}
}
```
