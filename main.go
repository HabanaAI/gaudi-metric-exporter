// Copyright (C) 2025 Intel Corporation

// This program is free software; you can redistribute it and/or modify it
// under the terms of the GNU General Public License version 2 or later, as published
// by the Free Software Foundation.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program; if not, see <http://www.gnu.org/licenses/>.

// SPDX-License-Identifier: GPL-2.0-or-later

package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	httpprof "net/http/pprof"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/time/rate"
)

func main() {
	ctx := context.Background()
	log := initLogger()

	if err := run(ctx, log); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func initLogger() *slog.Logger {
	level := slog.LevelInfo

	switch os.Getenv("LOG_LEVEL") {
	case slog.LevelDebug.String():
		level = slog.LevelDebug
	case slog.LevelWarn.String():
		level = slog.LevelWarn
	}

	h := slog.Handler(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     level,
		AddSource: true,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				if source, ok := a.Value.Any().(*slog.Source); ok {
					v := fmt.Sprintf("%s:%d", filepath.Base(source.File), source.Line)
					return slog.Attr{Key: "file", Value: slog.StringValue(v)}
				}
			}
			return a
		},
	}))
	h = h.WithAttrs([]slog.Attr{slog.String("service", "metric-exporter")})

	return slog.New(h)
}

func run(ctx context.Context, log *slog.Logger) error {
	port := flag.Int("port", 41611, "port for the application to run on")
	flag.Parse()

	app, err := initializeApplication(log, *port)
	if err != nil {
		return err
	}

	log.Info("Watching habanalabs driver for removal")

	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGTERM, os.Interrupt)

	serverError := make(chan error)
	go func() {
		log.Info("Starting server...", "port", *port)
		serverError <- app.server.ListenAndServe()
	}()

	select {
	// Shutting down due to signal
	case <-signalCh:
		log.Info("Received signal to shutdown")
		return shutdown(ctx, app.server)
	case err := <-serverError:
		log.With("error", err).Error("Failed to start server")
		return err
	}
}

type app struct {
	server   *http.Server
	exporter *Exporter
}

func initializeApplication(log *slog.Logger, port int) (*app, error) {
	e := Exporter{
		log: log,
		err: make(chan error),
	}
	prometheus.MustRegister(&e)

	handler := debugStandardLibraryMux()
	handler.Handle("/metrics", rateLimit(promhttp.Handler()))

	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      handler,
		WriteTimeout: time.Second * 10,
		ReadTimeout:  time.Second * 10,
	}

	// returning reference so caller can call Shutdown()
	return &app{
		server:   server,
		exporter: &e,
	}, nil
}

func debugStandardLibraryMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Register all the standard library debug endpoints.
	mux.HandleFunc("/debug/pprof/", httpprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", httpprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", httpprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", httpprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", httpprof.Trace)

	return mux
}

func shutdown(ctx context.Context, server *http.Server) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutting down server gracefully: %w", err)
	}

	return nil
}

// Since we want to avoid parallel request, and the call to the driver
// takes about 2-3 second, we limit requests for 1 per 5 seconds.
func rateLimit(next http.Handler) http.Handler {
	limiter := rate.NewLimiter(rate.Every(5*time.Second), 1)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !limiter.Allow() {
			http.Error(w, "rate limiting exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
