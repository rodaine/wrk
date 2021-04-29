// Package wrk is a utility for running multiple workers, such as web-servers
// (HTTP/gRPC/etc) as well as daemon-style goroutines in a coordinated
// fashion. On any errors, all workers are gracefully stopped either through
// context-cancellation or by executing a worker's defined Stop function.
package wrk
