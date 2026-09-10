package video

import "time"

type Status string

const (
	StatusProcessing Status = "PROCESSING"
	StatusReady      Status = "READY"
	StatusFailed     Status = "FAILED"
)

type Variant struct {
	Name       string `json:"name"`
	Width      int    `json:"width"`
	Height     int    `json:"height"`
	Bitrate    int    `json:"bitrate"`
	VideoCodec string `json:"video_codec"`
	AudioCodec string `json:"audio_codec"`
}

type Video struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Filename    string            `json:"filename,omitempty"`
	ContentType string            `json:"content_type"`
	Size        int64             `json:"size"`
	Duration    float64           `json:"duration_seconds"`
	Status      Status            `json:"status"`
	Source      string            `json:"source,omitempty"`
	SourceURL   string            `json:"source_url,omitempty"`
	Protocol    string            `json:"protocol,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	Variants    []Variant         `json:"variants"`
	Manifests   map[string]string `json:"manifests"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type Job struct {
	ID         string    `json:"id"`
	VideoID    string    `json:"video_id"`
	Kind       string    `json:"kind"`
	State      string    `json:"state"`
	Attempts   int       `json:"attempts"`
	MaxRetries int       `json:"max_retries"`
	LastError  string    `json:"last_error,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
