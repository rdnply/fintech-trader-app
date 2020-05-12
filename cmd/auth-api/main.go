package main

import (
	"context"
	"cw1/cmd/auth-api/handlers"
	"cw1/cmd/auth-api/handlers/trade"
	"cw1/cmd/auth-api/handlers/websocket"
	"cw1/internal/postgres"
	pb "cw1/internal/streamer"
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
	"google.golang.org/grpc"
)

func main() {
	logger := initLogger()

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

	robotStorage, err := postgres.NewRobotStorage(db)
	if err != nil {
		logger.Fatalf("Can't create robot storage: %s", err)
	}

	defer handleCloser(logger, "robot_storage", sessionStorage)

	hub := websocket.NewHub()
	go hub.Run()

	h, err := handler.New(logger, userStorage, sessionStorage, robotStorage, hub)
	if err != nil {
		logger.Fatalf("Can't create new handler: %s", err)
	}

	r := routes(h)
	addr := net.JoinHostPort("", "5000")
	srv := &http.Server{Addr: addr, Handler: r}

	const Duration = 5
	go gracefulShutdown(srv, Duration*time.Second, logger)

	conn, err := grpc.Dial(":8000", grpc.WithInsecure())
	if err != nil {
		logger.Fatalf("Can't create connection to price streamer: ", err)
	}

	defer handleCloser(logger, "price_streamer_connection", conn)

	tradingClient := pb.NewTradingServiceClient(conn)

	logger.Infof("Server is running at %v", addr)
	tr := trade.New(logger, tradingClient, robotStorage)
	quit := make(chan bool)
	go tr.StartDeals(quit)

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatal(err)
	}
	quit <- true
}

func handleCloser(l logger.Logger, resource string, closer io.Closer) {
	if err := closer.Close(); err != nil {
		l.Errorf("Can't close %q: %s", resource, err)
	}
}

func routes(h *handler.Handler) *chi.Mux {
	r := chi.NewRouter()

	const Duration = 60

	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(Duration * time.Second))

	r.Mount("/", h.Routes())

	return r
}

func gracefulShutdown(srv *http.Server, timeout time.Duration, logger logger.Logger) {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-done

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	logger.Infof("Shutting down server with %s timeout", timeout)

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Could not shutdown server:%v", err)
	}
}

func initLogger() logger.Logger {
	config := logger.Configuration{
		EnableConsole:     true,
		ConsoleLevel:      logger.Debug,
		ConsoleJSONFormat: true,
		EnableFile:        true,
		FileLevel:         logger.Info,
		FileJSONFormat:    true,
		FileLocation:      "log.log",
	}

	logger, err := logger.New(config, logger.InstanceZapLogger)
	if err != nil {
		log.Fatal("could not instantiate logger: ", err)
	}

	return logger
}
