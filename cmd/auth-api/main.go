package main

import (
	"context"
	"cw1/internal/postgres"
	"cw1/pkg/log/logger"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

var config = logger.Configuration{
	EnableConsole:     true,
	ConsoleLevel:      logger.Debug,
	ConsoleJSONFormat: true,
	EnableFile:        true,
	FileLevel:         logger.Info,
	FileJSONFormat:    true,
	FileLocation:      "log.log",
}

func main() {
	logger, err := logger.New(config, logger.InstanceZapLogger)
	if err != nil {
		log.Fatal("could not instantiate logger: ", err)
	}

	db, err := postgres.New(logger, "C:\\Users\\rodion\\go\\src\\cw1\\configuration.json")
	if err != nil {
		logger.Fatalf("Can't create database instance %v", err)
	}
	defer handleCloser(logger, "db", db)

	err = db.CheckConnection()
	if err != nil {
		logger.Fatalf("Can't connect to database %v", err)
	}

	userStorage, err := postgres.NewUserStorage(db)
	if err != nil {
		logger.Fatalf("Can't create user storage: %s", err)
	}
	defer handleCloser(logger, "user_storage", userStorage)

	sessionStorage, err := postgres.NewSessionStorage(db)
	if err != nil {
		logger.Fatalf("Can't create session storage: %s", err)
	}
	defer handleCloser(logger, "session_storage", sessionStorage)

	h := NewHandler(logger, userStorage, sessionStorage)
	r := routes(h)
	addr := net.JoinHostPort("", "5000")
	srv := &http.Server{Addr: addr, Handler: r}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-done

	const Timeout = 5
	ctx, cancel := context.WithTimeout(context.Background(), Timeout*time.Second)
	defer func() {
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("could not shutdown server:%v", err)
	}
}

func routes(h *Handler) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Mount("/", h.Routes())

	return r
}

func handleCloser(l logger.Logger, resource string, closer io.Closer) {
	if err := closer.Close(); err != nil {
		l.Errorf("Can't close %q: %s", resource, err)
	}
}
