package main

import (
	"context"
	"cw1/internal/db"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const Timeout = 5

func main() {
	addr := net.JoinHostPort("", "5000")
	srv := &http.Server{Addr: addr, Handler: router}

	db.Init("configuration.json")

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	<-done

	ctx, cancel := context.WithTimeout(context.Background(), Timeout*time.Second)
	defer func() {
		db := db.GetDBConn()
		db.Close()
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("could not shutdown server:%v", err)
	}

}
