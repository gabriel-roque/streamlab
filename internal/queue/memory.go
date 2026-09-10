package queue

import (
	"context"
	"sync"
	"time"

	"github.com/streamlab/streamlab/internal/video"
)

type Handler func(context.Context, video.Job) error

type MemoryQueue struct {
	mu      sync.RWMutex
	jobs    map[string]*video.Job
	dlq     []*video.Job
	work    chan video.Job
	closed  chan struct{}
	closeOn sync.Once
}

func NewMemory(size int) *MemoryQueue {
	if size < 1 {
		size = 16
	}
	return &MemoryQueue{jobs: make(map[string]*video.Job), work: make(chan video.Job, size), closed: make(chan struct{})}
}

func (q *MemoryQueue) Enqueue(j video.Job) error {
	now := time.Now().UTC()
	if j.ID == "" {
		j.ID = video.NewID("job")
	}
	if j.MaxRetries < 1 {
		j.MaxRetries = 3
	}
	j.State, j.CreatedAt, j.UpdatedAt = "QUEUED", now, now
	q.mu.Lock()
	q.jobs[j.ID] = &j
	q.mu.Unlock()
	select {
	case q.work <- j:
		return nil
	case <-q.closed:
		return context.Canceled
	}
}

func (q *MemoryQueue) Get(id string) (video.Job, bool) {
	q.mu.RLock()
	j, ok := q.jobs[id]
	if ok {
		j = cloneJob(j)
	}
	q.mu.RUnlock()
	if !ok {
		return video.Job{}, false
	}
	return *j, true
}

func (q *MemoryQueue) List() []video.Job {
	q.mu.RLock()
	out := make([]video.Job, 0, len(q.jobs))
	for _, j := range q.jobs {
		out = append(out, *cloneJob(j))
	}
	q.mu.RUnlock()
	return out
}

func (q *MemoryQueue) DLQ() []video.Job {
	q.mu.RLock()
	out := make([]video.Job, len(q.dlq))
	for i := range q.dlq {
		out[i] = *cloneJob(q.dlq[i])
	}
	q.mu.RUnlock()
	return out
}

func (q *MemoryQueue) Start(ctx context.Context, handler Handler) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-q.closed:
				return
			case j := <-q.work:
				q.run(ctx, j, handler)
			}
		}
	}()
}

func (q *MemoryQueue) run(ctx context.Context, j video.Job, handler Handler) {
	q.set(j.ID, func(item *video.Job) { item.State = "RUNNING"; item.Attempts++ })
	err := handler(ctx, j)
	if err == nil {
		q.set(j.ID, func(item *video.Job) { item.State = "SUCCEEDED"; item.LastError = "" })
		return
	}
	q.set(j.ID, func(item *video.Job) { item.LastError = err.Error() })
	q.mu.RLock()
	current := cloneJob(q.jobs[j.ID])
	q.mu.RUnlock()
	if current.Attempts <= current.MaxRetries {
		q.set(j.ID, func(item *video.Job) { item.State = "RETRYING" })
		time.AfterFunc(time.Duration(current.Attempts)*25*time.Millisecond, func() {
			q.set(j.ID, func(item *video.Job) { item.State = "QUEUED" })
			select {
			case q.work <- *current:
			case <-q.closed:
			}
		})
		return
	}
	q.set(j.ID, func(item *video.Job) { item.State = "DLQ" })
	q.mu.Lock()
	q.dlq = append(q.dlq, cloneJob(q.jobs[j.ID]))
	q.mu.Unlock()
}

func (q *MemoryQueue) set(id string, update func(*video.Job)) {
	q.mu.Lock()
	if j, ok := q.jobs[id]; ok {
		update(j)
		j.UpdatedAt = time.Now().UTC()
	}
	q.mu.Unlock()
}

func (q *MemoryQueue) Close() { q.closeOn.Do(func() { close(q.closed) }) }

func cloneJob(j *video.Job) *video.Job {
	c := *j
	return &c
}
