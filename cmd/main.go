package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"dorm/pkg/infrastructure/di"
)

const shutdownTimeout = 20 * time.Second

func main() {
	if err := run(); err != nil {
		log.Printf("Application stopped with error: %v", err)
		os.Exit(1)
	}
}

func run() (err error) {
	signals, stopSignals := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stopSignals()

	ctn, err := di.NewContainer("config.json")
	if err != nil {
		return fmt.Errorf("initialize container: %w", err)
	}

	var listenErr chan error
	var ready <-chan struct{}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err != nil {
			log.Printf("Stopping after error: %v", err)
		}
		log.Println("Shutting down...")
		err = errors.Join(err, stopApplication(ctx, ctn, ready, listenErr))
		if err == nil {
			log.Println("Application stopped gracefully")
		}
	}()

	log.Println("Starting Scheduler...")
	if err := ctn.Scheduler.Start(); err != nil {
		return fmt.Errorf("start scheduler: %w", err)
	}

	listener, err := net.Listen("tcp4", ":8080")
	if err != nil {
		return fmt.Errorf("listen HTTP on :8080: %w", err)
	}
	ln := &httpListener{Listener: listener, ready: make(chan struct{})}
	ready = ln.ready
	listenErr = make(chan error, 1)
	go func() {
		log.Println("Starting HTTP server on :8080...")
		err := ctn.HTTPServer.Listener(ln)
		ln.once.Do(func() { close(ln.ready) })
		listenErr <- err
	}()

	select {
	case <-signals.Done():
		log.Println("Received shutdown signal")
		select {
		case err := <-listenErr:
			listenErr = nil
			return unexpectedHTTPExit(err)
		default:
			return nil
		}
	case err := <-listenErr:
		listenErr = nil
		return unexpectedHTTPExit(err)
	}
}

type httpListener struct {
	net.Listener
	ready chan struct{}
	once  sync.Once
}

func (ln *httpListener) Accept() (net.Conn, error) {
	ln.once.Do(func() { close(ln.ready) })
	return ln.Listener.Accept()
}

func unexpectedHTTPExit(err error) error {
	if err == nil {
		return errors.New("HTTP server stopped unexpectedly without an error")
	}
	return fmt.Errorf("HTTP server stopped unexpectedly: %w", err)
}

func stopApplication(ctx context.Context, ctn *di.Container, ready <-chan struct{}, listenErr <-chan error) error {
	cronDone := ctn.Scheduler.Stop().Done()
	log.Println("Scheduler stopped accepting jobs; active jobs cancelled")

	var httpDone <-chan error
	if ready != nil {
		done := make(chan error, 1)
		httpDone = done
		go func() {
			select {
			case <-ready:
				log.Println("Starting HTTP graceful shutdown")
				done <- ctn.HTTPServer.ShutdownWithContext(ctx)
			case <-ctx.Done():
				done <- ctx.Err()
			}
		}()
	}

	var result error
wait:
	for cronDone != nil || httpDone != nil || listenErr != nil {
		select {
		case <-cronDone:
			cronDone = nil
			log.Println("Scheduler jobs finished")
		case err := <-httpDone:
			httpDone = nil
			if err != nil {
				result = errors.Join(result, fmt.Errorf("HTTP shutdown: %w", err))
			} else {
				log.Println("HTTP requests finished")
			}
		case err := <-listenErr:
			listenErr = nil
			if err != nil {
				result = errors.Join(result, fmt.Errorf("HTTP server exit during shutdown: %w", err))
			}
		case <-ctx.Done():
			if cronDone != nil {
				result = errors.Join(result, fmt.Errorf("wait for scheduler jobs: %w", ctx.Err()))
			}
			if httpDone != nil {
				result = errors.Join(result, fmt.Errorf("wait for HTTP shutdown: %w", ctx.Err()))
			}
			if listenErr != nil {
				result = errors.Join(result, fmt.Errorf("wait for HTTP server exit: %w", ctx.Err()))
			}
			break wait
		}
	}

	eventsDone := ctn.EventBus.Stop()
	select {
	case <-eventsDone.Done():
		log.Println("Event handlers finished")
	case <-ctx.Done():
		result = errors.Join(result, fmt.Errorf("wait for event handlers: %w", ctx.Err()))
	}

	closed := make(chan error, 1)
	go func() { closed <- ctn.Close() }()
	select {
	case err := <-closed:
		if err != nil {
			return errors.Join(result, err)
		}
		log.Println("Database connection closed")
		return result
	case <-ctx.Done():
		return errors.Join(result, fmt.Errorf("wait for database close: %w", ctx.Err()))
	}
}
