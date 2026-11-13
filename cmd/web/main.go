package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"prj_withdrawal/internal/bank"
	"prj_withdrawal/internal/storage"
	"prj_withdrawal/internal/web"
)

func main() {
	address := flag.String("address", "127.0.0.1:8080", "dirección HTTP del servidor")
	dataPath := flag.String("data", "data/state.json", "ruta del archivo JSON de cuenta y transacciones")
	flag.Parse()

	service := bank.NewService(storage.NewJSONStore(*dataPath))
	server := &http.Server{
		Addr:              *address,
		Handler:           web.NewServer(service),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("cerrar servidor: %v", err)
		}
	}()

	log.Printf("App web disponible en http://%s", *address)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
