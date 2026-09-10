package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/streamlab/streamlab/internal/queue"
	"github.com/streamlab/streamlab/internal/storage"
	"github.com/streamlab/streamlab/internal/transcoder"
)

// This process is a standalone worker shell for deployments that later replace
// the in-memory queue with a durable broker. The API embeds the same worker by
// default, so local development needs only apps/api.
func main() {
	store, err := storage.NewLocalStore(os.Getenv("STREAMLAB_STORAGE_ROOT"))
	if err != nil {
		log.Fatal(err)
	}
	jobs := queue.NewMemory(32)
	processor := &transcoder.Processor{Store: store}
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	jobs.Start(ctx, processor.Process)
	log.Printf("streamlab transcoder worker ready (local root: %s)", store.Root())
	<-ctx.Done()
	jobs.Close()
}
