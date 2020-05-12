package main

import (
	"context"
	handler "cw1/cmd/auth-api/handlers"
	"cw1/cmd/auth-api/handlers/socket"
	"cw1/cmd/auth-api/handlers/trade"
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

	st, closers := initStorages(logger)

	defer handleClosers(logger, closers)

	hub := socket.NewHub()
	go hub.Run()

	h, err := handler.New(logger, st.u, st.s, st.r, hub)
	if err != nil {
		logger.Fatalf("Can't create new handler: %s", err)
	}

	srv := initServer(h, "", "5000")

	const Duration = 5
	go gracefulShutdown(srv, Duration*time.Second, logger)

	conn, err := grpc.Dial(":8000", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		logger.Fatalf("Can't create connection to price streamer: ", err)
	}

	defer handleCloser(logger, "price_streamer_connection", conn)

	tradingClient := pb.NewTradingServiceClient(conn)

	logger.Infof("Server is running at %s", "5000")
	tr := trade.New(logger, tradingClient, st.r, hub)

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

func handleClosers(l logger.Logger, m map[string]io.Closer) {
	for n, c := range m {
		if err := c.Close(); err != nil {
			l.Errorf("Can't close %q: %s", n, err)
		}
	}
}

type storages struct {
	u *postgres.UserStorage
	s *postgres.SessionStorage
	r *postgres.RobotStorage
}

func initStorages(logger logger.Logger) (*storages, map[string]io.Closer) {
	closers := make(map[string]io.Closer)

	db, err := postgres.New(logger, "C:\\Users\\rodion\\go\\src\\cw1\\configuration.json")
	if err != nil {
		logger.Fatalf("Can't create database instance %v", err)
	}

	closers["db"] = db

	err = db.CheckConnection()
	if err != nil {
		logger.Fatalf("Can't connect to database %v", err)
	}

	userStorage, err := postgres.NewUserStorage(db)
	if err != nil {
		logger.Fatalf("Can't create user storage: %s", err)
	}

	closers["user_storage"] = userStorage

	sessionStorage, err := postgres.NewSessionStorage(db)
	if err != nil {
		logger.Fatalf("Can't create session storage: %s", err)
	}

	closers["session_storage"] = sessionStorage

	robotStorage, err := postgres.NewRobotStorage(db)
	if err != nil {
		logger.Fatalf("Can't create robot storage: %s", err)
	}

	closers["robot_storage"] = robotStorage

	return &storages{userStorage, sessionStorage, robotStorage}, closers
}

func initServer(h *handler.Handler, host string, port string) *http.Server {
	r := routes(h)
	addr := net.JoinHostPort(host, port)
	srv := &http.Server{Addr: addr, Handler: r}

	return srv
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
