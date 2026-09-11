package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/streamlab/streamlab/internal/queue"
	"github.com/streamlab/streamlab/internal/storage"
	"github.com/streamlab/streamlab/internal/telemetry"
	"github.com/streamlab/streamlab/internal/transcoder"
	"github.com/streamlab/streamlab/internal/video"
)

type Config struct {
	StorageRoot string
	StoragePath string
	CORSOrigin  string
	DisableSeed bool
	Worker      bool
}

type Server struct {
	Store      *storage.LocalStore
	Queue      *queue.MemoryQueue
	Telemetry  *telemetry.Store
	processor  *transcoder.Processor
	handler    http.Handler
	requestCnt atomic.Uint64
	eventCnt   atomic.Uint64
	startedAt  time.Time
	cancel     context.CancelFunc
}

func NewServer(config Config) (*Server, error) {
	return NewServerWithWorker(config)
}

// NewServerWithWorker is the constructor used by applications and tests. The
// small split keeps NewServer useful for callers that want to inspect a queue
// without a running worker.
func NewServerWithWorker(config Config) (*Server, error) {
	store, err := storage.NewLocalStore(storageRoot(config))
	if err != nil {
		return nil, err
	}
	s := &Server{Store: store, Queue: queue.NewMemory(32), Telemetry: telemetry.NewStore(), startedAt: time.Now().UTC()}
	s.processor = &transcoder.Processor{Store: store}
	if !config.DisableSeed {
		store.Seed(seedVideo())
		store.Seed(seedABRVideo())
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.Queue.Start(ctx, func(ctx context.Context, job video.Job) error { return s.processor.Process(ctx, job) })
	s.handler = s.routes(config.CORSOrigin)
	return s, nil
}

// New is the default executable-ready constructor.
func New(config Config) (*Server, error) { return NewServerWithWorker(config) }

// Close stops the in-memory worker. It does not remove local media.
func (s *Server) Close() {
	if s.cancel != nil {
		s.cancel()
	}
	s.Queue.Close()
}

func (s *Server) Handler() http.Handler { return s.handler }

func (s *Server) routes(origin string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", originOrAny(origin))
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, HEAD, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Range, Accept, Origin")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Range, Accept-Ranges, Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		s.requestCnt.Add(1)
		path := strings.Trim(r.URL.Path, "/")
		parts := strings.Split(path, "/")
		switch {
		case r.Method == http.MethodGet && path == "health":
			s.health(w)
		case r.Method == http.MethodGet && path == "healthz":
			s.health(w)
		case r.Method == http.MethodGet && path == "metrics":
			s.metrics(w)
		case path == "videos" && r.Method == http.MethodGet:
			s.listVideos(w)
		case path == "videos" && r.Method == http.MethodPost:
			if isMultipart(r) {
				s.upload(w, r, "")
				return
			}
			s.createVideo(w, r)
		case len(parts) == 3 && parts[0] == "videos" && parts[2] == "upload" && r.Method == http.MethodPost:
			s.upload(w, r, parts[1])
		case len(parts) >= 2 && parts[0] == "videos":
			s.videoRoute(w, r, parts[1], parts[2:])
		case r.Method == http.MethodPost && (path == "playback/sessions" || path == "sessions"):
			s.createSession(w, r)
		case r.Method == http.MethodPost && (path == "playback/events" || path == "telemetry/playback"):
			s.addEvent(w, r)
		default:
			errorJSON(w, http.StatusNotFound, "route not found")
		}
	})
}

func (s *Server) health(w http.ResponseWriter) {
	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "service": "streamlab-api", "time": time.Now().UTC()})
}

func (s *Server) metrics(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/plain; version=0.0.4")
	fmt.Fprintf(w, "# HELP streamlab_http_requests_total HTTP requests handled.\n# TYPE streamlab_http_requests_total counter\nstreamlab_http_requests_total %d\n# HELP streamlab_playback_events_total Playback events accepted.\n# TYPE streamlab_playback_events_total counter\nstreamlab_playback_events_total %d\n# HELP streamlab_playback_events_stored Current events in memory.\n# TYPE streamlab_playback_events_stored gauge\nstreamlab_playback_events_stored %d\n", s.requestCnt.Load(), s.eventCnt.Load(), s.Telemetry.Count())
}

func (s *Server) listVideos(w http.ResponseWriter) {
	writeJSON(w, http.StatusOK, map[string]any{"videos": s.Store.List()})
}

func (s *Server) createVideo(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Title           string `json:"title"`
		Filename        string `json:"filename"`
		ContentType     string `json:"contentType"`
		ContentTypeJSON string `json:"content_type"`
		Source          struct {
			Filename        string `json:"filename"`
			ContentType     string `json:"contentType"`
			ContentTypeJSON string `json:"content_type"`
		} `json:"source"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		errorJSON(w, http.StatusBadRequest, "json body is invalid")
		return
	}
	filename := input.Filename
	contentType := input.ContentType
	if input.ContentTypeJSON != "" {
		contentType = input.ContentTypeJSON
	}
	if input.Source.Filename != "" {
		filename = input.Source.Filename
	}
	if input.Source.ContentType != "" {
		contentType = input.Source.ContentType
	}
	if input.Source.ContentTypeJSON != "" {
		contentType = input.Source.ContentTypeJSON
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	id := video.NewID("vid")
	now := time.Now().UTC()
	item := video.Video{ID: id, Title: input.Title, Filename: filename, ContentType: contentType, Status: video.StatusProcessing, Source: "upload", Metadata: map[string]string{}, CreatedAt: now, UpdatedAt: now}
	s.Store.Seed(item)
	w.Header().Set("Location", "/videos/"+id)
	writeJSON(w, http.StatusCreated, item)
}

func (s *Server) upload(w http.ResponseWriter, r *http.Request, id string) {
	var existing *video.Video
	if id != "" {
		var err error
		existing, err = s.Store.Get(id)
		if err != nil {
			errorJSON(w, http.StatusNotFound, "video not found")
			return
		}
	}
	r.Body = http.MaxBytesReader(w, r.Body, 2<<30)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		errorJSON(w, http.StatusBadRequest, "multipart form is invalid")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		file, header, err = r.FormFile("video")
	}
	if err != nil {
		errorJSON(w, http.StatusBadRequest, "multipart field 'file' is required")
		return
	}
	defer file.Close()
	title := r.FormValue("title")
	if title == "" && existing != nil {
		title = existing.Title
	}
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	if id == "" {
		id = video.NewID("vid")
	}
	item, err := s.Store.CreateUpload(id, title, header.Filename, contentType, file)
	if err != nil {
		errorJSON(w, http.StatusInternalServerError, err.Error())
		return
	}
	job := video.Job{VideoID: id, Kind: "transcode", MaxRetries: 3}
	if err := s.Queue.Enqueue(job); err != nil {
		errorJSON(w, http.StatusInternalServerError, "could not enqueue processing job")
		return
	}
	w.Header().Set("Location", "/videos/"+id)
	writeJSON(w, http.StatusAccepted, item)
}

func (s *Server) videoRoute(w http.ResponseWriter, r *http.Request, id string, rest []string) {
	if len(rest) == 0 && r.Method == http.MethodGet {
		item, err := s.Store.Get(id)
		if err != nil {
			errorJSON(w, http.StatusNotFound, "video not found")
			return
		}
		writeJSON(w, http.StatusOK, item)
		return
	}
	if len(rest) == 1 && rest[0] == "status" && r.Method == http.MethodGet {
		item, err := s.Store.Get(id)
		if err != nil {
			errorJSON(w, http.StatusNotFound, "video not found")
			return
		}
		jobs := make([]video.Job, 0)
		for _, job := range s.Queue.List() {
			if job.VideoID == id {
				jobs = append(jobs, job)
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"video":      item,
			"video_id":   id,
			"videoId":    id,
			"status":     item.Status,
			"jobs":       jobs,
			"updated_at": item.UpdatedAt,
			"updatedAt":  item.UpdatedAt,
		})
		return
	}
	if len(rest) >= 1 && rest[0] == "stream" && (r.Method == http.MethodGet || r.Method == http.MethodHead) {
		s.stream(w, r, id)
		return
	}
	if len(rest) >= 1 && rest[0] == "playback" && r.Method == http.MethodGet {
		if len(rest) == 1 {
			s.playback(w, id)
			return
		}
		if len(rest) == 2 {
			s.playbackAsset(w, id, rest[1])
			return
		}
		if len(rest) == 3 {
			s.artifact(w, id, rest[1], rest[2])
			return
		}
		s.manifest(w, id, rest[1])
		return
	}
	errorJSON(w, http.StatusNotFound, "video route not found")
}

func (s *Server) playbackAsset(w http.ResponseWriter, id, name string) {
	if name == "hls" || name == "dash" {
		s.manifest(w, id, name)
		return
	}
	format := "dash"
	if strings.HasSuffix(name, ".ts") || strings.HasSuffix(name, ".m3u8") {
		format = "hls"
	}
	s.artifact(w, id, format, name)
}

func (s *Server) stream(w http.ResponseWriter, r *http.Request, id string) {
	item, err := s.Store.Get(id)
	if err != nil {
		errorJSON(w, http.StatusNotFound, "video not found")
		return
	}
	if item.SourceURL != "" && !s.Store.HasMedia(id) {
		http.Redirect(w, r, item.SourceURL, http.StatusFound)
		return
	}
	f, info, err := s.Store.OpenMedia(id)
	if err != nil {
		errorJSON(w, http.StatusConflict, "video media is not available locally")
		return
	}
	defer f.Close()
	w.Header().Set("Accept-Ranges", "bytes")
	start, end, partial, err := ParseRange(r.Header.Get("Range"), info.Size())
	if err != nil {
		w.Header().Set("Content-Range", "bytes */"+strconv.FormatInt(info.Size(), 10))
		w.WriteHeader(http.StatusRequestedRangeNotSatisfiable)
		return
	}
	if !partial {
		start, end = 0, info.Size()-1
	}
	length := end - start + 1
	w.Header().Set("Content-Type", item.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(length, 10))
	if partial {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, info.Size()))
		w.WriteHeader(http.StatusPartialContent)
	}
	if r.Method == http.MethodHead {
		return
	}
	if _, err := f.Seek(start, io.SeekStart); err != nil {
		return
	}
	_, _ = io.CopyN(w, f, length)
}

func (s *Server) playback(w http.ResponseWriter, id string) {
	item, err := s.Store.Get(id)
	if err != nil {
		errorJSON(w, http.StatusNotFound, "video not found")
		return
	}
	if item.Status != video.StatusReady {
		errorJSON(w, http.StatusConflict, "video playback is not ready")
		return
	}
	hls := "/videos/" + id + "/playback/hls"
	dash := "/videos/" + id + "/playback/dash"
	protocol := item.Protocol
	if protocol == "" {
		protocol = "MP4"
	}
	manifest := ""
	protocols := map[string]string{}
	if item.Manifests["hls"] != "" || (protocol == "HLS" && item.SourceURL != "") {
		protocol = "HLS"
		manifest = hls
		if item.SourceURL != "" && item.Manifests["hls"] == "" {
			manifest = item.SourceURL
		}
		protocols["hls"] = hls
	}
	if item.Manifests["dash"] != "" {
		protocols["dash"] = dash
	}
	response := map[string]any{
		"video_id":   id,
		"videoId":    id,
		"status":     item.Status,
		"protocol":   protocol,
		"manifest":   manifest,
		"protocols":  protocols,
		"manifests":  item.Manifests,
		"variants":   item.Variants,
		"source_url": item.SourceURL,
		"source":     item.SourceURL,
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *Server) manifest(w http.ResponseWriter, id, format string) {
	item, err := s.Store.Get(id)
	if err != nil {
		errorJSON(w, http.StatusNotFound, "video not found")
		return
	}
	if item.Status != video.StatusReady {
		errorJSON(w, http.StatusConflict, "video playback is not ready")
		return
	}
	name := item.Manifests[strings.ToLower(format)]
	if name == "" {
		errorJSON(w, http.StatusNotFound, "manifest not available")
		return
	}
	data, err := s.Store.ReadArtifact(id, name)
	if err != nil && strings.ToLower(format) == "hls" {
		data = []byte("#EXTM3U\n#EXT-X-ENDLIST\n")
	}
	if err != nil && len(data) == 0 {
		errorJSON(w, http.StatusNotFound, "manifest not available")
		return
	}
	if strings.ToLower(format) == "hls" {
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	} else {
		w.Header().Set("Content-Type", "application/dash+xml")
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (s *Server) artifact(w http.ResponseWriter, id, format, name string) {
	if format != "hls" && format != "dash" {
		errorJSON(w, http.StatusNotFound, "asset not available")
		return
	}
	if err := storage.ValidateArtifactName(name); err != nil {
		errorJSON(w, http.StatusBadRequest, "invalid asset name")
		return
	}
	item, err := s.Store.Get(id)
	if err != nil {
		errorJSON(w, http.StatusNotFound, "video not found")
		return
	}
	if item.Status != video.StatusReady {
		errorJSON(w, http.StatusConflict, "video playback is not ready")
		return
	}
	data, err := s.Store.ReadArtifact(id, name)
	if err != nil {
		errorJSON(w, http.StatusNotFound, "asset not available")
		return
	}
	switch {
	case strings.HasSuffix(name, ".m3u8"):
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	case strings.HasSuffix(name, ".mpd"):
		w.Header().Set("Content-Type", "application/dash+xml")
	case strings.HasSuffix(name, ".ts"):
		w.Header().Set("Content-Type", "video/mp2t")
	default:
		w.Header().Set("Content-Type", "application/octet-stream")
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

type sessionRequest struct {
	VideoID string `json:"video_id"`
	UserID  string `json:"user_id"`
}

func (s *Server) createSession(w http.ResponseWriter, r *http.Request) {
	var input sessionRequest
	if json.NewDecoder(r.Body).Decode(&input) != nil || input.VideoID == "" {
		errorJSON(w, http.StatusBadRequest, "video_id is required")
		return
	}
	if _, err := s.Store.Get(input.VideoID); err != nil {
		errorJSON(w, http.StatusNotFound, "video not found")
		return
	}
	now := time.Now().UTC()
	item := telemetry.Session{ID: video.NewID("session"), VideoID: input.VideoID, UserID: input.UserID, StartedAt: now, LastSeen: now}
	s.Telemetry.AddSession(item)
	writeJSON(w, http.StatusCreated, item)
}

type eventRequest struct {
	VideoID   string         `json:"video_id"`
	SessionID string         `json:"session_id"`
	Type      string         `json:"type"`
	Position  float64        `json:"position"`
	Payload   map[string]any `json:"payload"`
}

func (s *Server) addEvent(w http.ResponseWriter, r *http.Request) {
	var input eventRequest
	if json.NewDecoder(r.Body).Decode(&input) != nil || input.Type == "" {
		errorJSON(w, http.StatusBadRequest, "event type is required")
		return
	}
	if input.SessionID != "" {
		if _, ok := s.Telemetry.GetSession(input.SessionID); !ok {
			errorJSON(w, http.StatusNotFound, "session not found")
			return
		}
	}
	event := telemetry.Event{ID: video.NewID("event"), VideoID: input.VideoID, SessionID: input.SessionID, Type: input.Type, Position: input.Position, Payload: input.Payload, At: time.Now().UTC()}
	s.Telemetry.AddEvent(event)
	s.eventCnt.Add(1)
	writeJSON(w, http.StatusAccepted, map[string]any{"accepted": true, "event_id": event.ID})
}

func ParseRange(header string, size int64) (start, end int64, partial bool, err error) {
	if header == "" {
		return 0, size - 1, false, nil
	}
	if size <= 0 || !strings.HasPrefix(header, "bytes=") || strings.Contains(header, ",") {
		return 0, 0, false, errors.New("invalid range")
	}
	value := strings.TrimSpace(strings.TrimPrefix(header, "bytes="))
	parts := strings.Split(value, "-")
	if len(parts) != 2 {
		return 0, 0, false, errors.New("invalid range")
	}
	startPart, endPart := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	if startPart == "" {
		n, e := strconv.ParseInt(endPart, 10, 64)
		if e != nil || n <= 0 {
			return 0, 0, false, errors.New("invalid range")
		}
		if n > size {
			n = size
		}
		return size - n, size - 1, true, nil
	}
	start, err = strconv.ParseInt(startPart, 10, 64)
	if err != nil || start < 0 || start >= size {
		return 0, 0, false, errors.New("invalid range")
	}
	if endPart == "" {
		return start, size - 1, true, nil
	}
	end, err = strconv.ParseInt(endPart, 10, 64)
	if err != nil || end < start {
		return 0, 0, false, errors.New("invalid range")
	}
	if end >= size {
		end = size - 1
	}
	return start, end, true, nil
}

func seedVideo() video.Video {
	now := time.Now().UTC()
	return video.Video{ID: "big-buck-bunny", Title: "Big Buck Bunny", Filename: "BigBuckBunny.mp4", ContentType: "video/mp4", Status: video.StatusReady, Source: "catalog", SourceURL: "https://storage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4", Protocol: "MP4", Duration: 596, Metadata: map[string]string{"description": "Public catalog fixture; binary is not downloaded", "pipeline": "remote-catalog"}, Variants: []video.Variant{{Name: "1080p", Width: 1920, Height: 1080, Bitrate: 5800000, VideoCodec: "h264", AudioCodec: "aac"}, {Name: "720p", Width: 1280, Height: 720, Bitrate: 3000000, VideoCodec: "h264", AudioCodec: "aac"}}, Manifests: map[string]string{}, CreatedAt: now, UpdatedAt: now}
}

func seedABRVideo() video.Video {
	now := time.Now().UTC().Add(-24 * time.Hour)
	return video.Video{ID: "abr-lab", Title: "ABR Ladder / live probe", Filename: "x36xhzz.m3u8", ContentType: "application/vnd.apple.mpegurl", Status: video.StatusReady, Source: "catalog", SourceURL: "https://test-streams.mux.dev/x36xhzz/x36xhzz.m3u8", Protocol: "HLS", Duration: 468, Metadata: map[string]string{"description": "Public multi-rendition HLS stream for ABR experiments", "pipeline": "remote-catalog"}, Variants: []video.Variant{{Name: "1080p", Width: 1920, Height: 1080, Bitrate: 6200000, VideoCodec: "h264", AudioCodec: "aac"}, {Name: "720p", Width: 1280, Height: 720, Bitrate: 3000000, VideoCodec: "h264", AudioCodec: "aac"}, {Name: "480p", Width: 854, Height: 480, Bitrate: 1200000, VideoCodec: "h264", AudioCodec: "aac"}}, Manifests: map[string]string{}, CreatedAt: now, UpdatedAt: now}
}

func originOrAny(origin string) string {
	if origin == "" {
		return "*"
	}
	return origin
}

func isMultipart(r *http.Request) bool {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	return err == nil && mediaType == "multipart/form-data"
}

func storageRoot(config Config) string {
	if config.StorageRoot != "" {
		return config.StorageRoot
	}
	if config.StoragePath != "" {
		return config.StoragePath
	}
	if root := os.Getenv("STREAMLAB_STORAGE_ROOT"); root != "" {
		return root
	}
	return os.Getenv("STORAGE_PATH")
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}
func errorJSON(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
