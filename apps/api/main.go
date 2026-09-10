package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/streamlab/streamlab/internal/api"
)

func main() {
	port := 8080
	if value, err := strconv.Atoi(os.Getenv("PORT")); err == nil && value > 0 {
		port = value
	}
	server, err := api.New(api.Config{StorageRoot: os.Getenv("STREAMLAB_STORAGE_ROOT"), StoragePath: os.Getenv("STORAGE_PATH"), CORSOrigin: os.Getenv("CORS_ORIGIN")})
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	httpServer := &http.Server{Addr: ":" + strconv.Itoa(port), Handler: server.Handler()}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-stop
		_ = httpServer.Close()
	}()
	log.Printf("streamlab api listening on %s", httpServer.Addr)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
