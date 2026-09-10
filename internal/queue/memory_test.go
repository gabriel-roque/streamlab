package queue

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/streamlab/streamlab/internal/video"
)

func TestMemoryQueueRetriesThenDLQ(t *testing.T) {
	q := NewMemory(2)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer q.Close()
	q.Start(ctx, func(context.Context, video.Job) error { return errors.New("fixture failure") })
	if err := q.Enqueue(video.Job{VideoID: "vid-test", MaxRetries: 1}); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if len(q.DLQ()) == 1 {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("job did not reach DLQ: %+v", q.List())
}
