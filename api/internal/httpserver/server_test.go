package httpserver

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"
)

// discard is a logger that throws its output away, for tests that ignore logs.
var discard = slog.New(slog.NewTextHandler(io.Discard, nil))

func TestRunServesUntilContextIsCanceled(t *testing.T) {
	// --- Arrange ---
	// Open a port for the server and create the context that will stop it.
	ln := listen(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Use an unusual status code, so a match proves our handler answered.
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	// Start the server in the background, because Run blocks until it stops.
	done := make(chan error, 1)
	go func() { done <- Run(ctx, ln, discard, h) }()

	// --- Act ---
	// Send a request; it waits in the port's queue until Run accepts it.
	status := get(url(ln))

	// --- Assert ---
	// Check that our handler produced the response.
	if status != http.StatusTeapot {
		t.Errorf("status = %d, want %d", status, http.StatusTeapot)
	}

	// --- Act ---
	// Ask the server to stop, as Ctrl+C would.
	cancel()

	// --- Assert ---
	// Run must return nil, meaning it stopped cleanly.
	if err := receive(t, done, "Run to return"); err != nil {
		t.Fatalf("Run() = %v, want nil", err)
	}
}

func TestRunLetsInFlightRequestsFinish(t *testing.T) {
	// --- Arrange ---
	// Open a port for the server and create the context that will stop it.
	ln := listen(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Use a handler that keeps the request open until the test releases it.
	h, arrived, release := holdingHandler()

	// Start the server in the background, because Run blocks until it stops.
	done := make(chan error, 1)
	go func() { done <- Run(ctx, ln, discard, h) }()

	// Send the request in the background, because it blocks until the handler is released.
	status := make(chan int, 1)
	go func() { status <- get(url(ln)) }()

	// Wait until the request is inside the handler, so it is truly in flight.
	receive(t, arrived, "the request to reach the handler")

	// --- Act ---
	// Start shutting down while that request is still running.
	cancel()

	// --- Assert ---
	// Run must keep waiting, because the request has not finished yet.
	select {
	case err := <-done:
		t.Fatalf("Run() = %v before the in-flight request finished", err)
	case <-time.After(100 * time.Millisecond):
	}

	// Meanwhile, the server must refuse new connections.
	if conn, err := net.DialTimeout("tcp", ln.Addr().String(), time.Second); err == nil {
		conn.Close()
		t.Error("new connection accepted during shutdown, want it refused")
	}

	// --- Act ---
	// Let the handler finish its response.
	close(release)

	// --- Assert ---
	// The in-flight request must succeed despite the shutdown.
	if got := receive(t, status, "the response"); got != http.StatusOK {
		t.Errorf("status = %d, want %d", got, http.StatusOK)
	}

	// With no requests left, Run must return nil.
	if err := receive(t, done, "Run to return"); err != nil {
		t.Fatalf("Run() = %v, want nil", err)
	}
}

func TestRunGivesUpWhenShutdownTakesTooLong(t *testing.T) {
	// --- Arrange ---
	// Shorten the shutdown timeout so the test runs fast, and restore it afterwards.
	original := shutdownTimeout
	shutdownTimeout = 100 * time.Millisecond
	t.Cleanup(func() { shutdownTimeout = original })

	// Open a port for the server and create the context that will stop it.
	ln := listen(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Use a handler that stays busy for the whole test, and free it at the end.
	h, arrived, release := holdingHandler()
	defer close(release)

	// Start the server in the background, because Run blocks until it stops.
	done := make(chan error, 1)
	go func() { done <- Run(ctx, ln, discard, h) }()

	// Send a request that will still be running when shutdown starts.
	go get(url(ln))

	// Wait until the request is inside the handler, so it is truly in flight.
	receive(t, arrived, "the request to reach the handler")

	// --- Act ---
	// Start shutting down while that request is still running.
	cancel()

	// --- Assert ---
	// Run must give up after the timeout and report it as an error.
	if err := receive(t, done, "Run to return"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("Run() = %v, want %v", err, context.DeadlineExceeded)
	}
}

func TestRunSendsServerErrorsToTheLogger(t *testing.T) {
	// --- Arrange ---
	// Write logs to a buffer instead of discarding them, so the test can read them.
	var logs bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&logs, nil))

	// Open a port for the server and create the context that will stop it.
	ln := listen(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Use a handler that panics; net/http recovers the panic and reports it through ErrorLog.
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})

	// Start the server in the background, because Run blocks until it stops.
	done := make(chan error, 1)
	go func() { done <- Run(ctx, ln, logger, h) }()

	// --- Act ---
	// Trigger the panic; the request itself fails because the server drops the connection.
	get(url(ln))

	// Stop the server and wait, so every log line is written before the test reads them.
	cancel()
	err := receive(t, done, "Run to return")

	// --- Assert ---
	// The server must have stopped cleanly.
	if err != nil {
		t.Fatalf("Run() = %v, want nil", err)
	}

	// The panic must appear in our logger as an Error-level record.
	if !strings.Contains(logs.String(), `level=ERROR msg="http: panic serving`) {
		t.Errorf("panic not logged at Error level; logs:\n%s", logs.String())
	}
}

func TestRunFailsWhenListenerFails(t *testing.T) {
	// --- Arrange ---
	// Open a port and close it at once, so every attempt to accept a connection fails.
	ln := listen(t)
	ln.Close()

	// Add a deadline: if Run missed the failure, it returns nil at the deadline instead of hanging.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// --- Act ---
	// Run the server on the broken listener.
	err := Run(ctx, ln, discard, http.NotFoundHandler())

	// --- Assert ---
	// Run must return the listener's error.
	if !errors.Is(err, net.ErrClosed) {
		t.Fatalf("Run() = %v, want %v", err, net.ErrClosed)
	}
}

// listen opens a listener on a free local port.
func listen(t *testing.T) net.Listener {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	return ln
}

// url returns the address of the server listening on ln.
func url(ln net.Listener) string {
	return "http://" + ln.Addr().String()
}

// get sends a GET request and returns the status code, or 0 if the request failed.
func get(target string) int {
	resp, err := http.Get(target)
	if err != nil {
		return 0
	}
	resp.Body.Close()
	return resp.StatusCode
}

// holdingHandler returns a handler that keeps each request open until release is closed.
// The arrived channel is closed when the first request reaches the handler.
func holdingHandler() (h http.Handler, arrived, release chan struct{}) {
	arrived = make(chan struct{})
	release = make(chan struct{})
	var once sync.Once
	h = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		once.Do(func() { close(arrived) })
		<-release
		w.WriteHeader(http.StatusOK)
	})
	return h, arrived, release
}

// receive waits for a value from ch, failing the test if none arrives within 5 seconds.
func receive[T any](t *testing.T, ch <-chan T, what string) T {
	t.Helper()
	select {
	case v := <-ch:
		return v
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for %s", what)
		var zero T
		return zero
	}
}
