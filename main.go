package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/ssongin/tartarus/server"
)

const (
	version = "1.0.0"
)

type config struct {
	port int
	env  string
	// dsn  string
}

type Application struct {
	config config
	logger *slog.Logger
	router *server.TartarusRouter
	// repositories data.Repositories
	// services     services.Services
	// routes rout
}

func main() {
	var cfg config
	flag.IntVar(&cfg.port, "port", 8080, "HTTP network address")
	flag.StringVar(&cfg.env, "env", "dev", "Environment (dev|test|prod)")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	var router *server.TartarusRouter

	app := &Application{
		config: cfg,
		logger: logger,
		router: router,
	}

	addr := fmt.Sprintf(":%d", cfg.port)

	srv := &http.Server{
		Addr:         addr,
		Handler:      app.router.Route(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.Info("Starting %s server on %s", cfg.env, addr)

	if err := srv.ListenAndServe(); err != nil {
		panic(err)
	}
}
