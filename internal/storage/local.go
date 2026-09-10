package storage

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/streamlab/streamlab/internal/video"
)

var ErrNotFound = errors.New("video not found")

// LocalStore is the default adapter. Metadata is held in memory while media and
// generated artifacts are kept on disk, so the API works without external services.
type LocalStore struct {
	root   string
	mu     sync.RWMutex
	videos map[string]*video.Video
}

func NewLocalStore(root string) (*LocalStore, error) {
	if root == "" {
		root = filepath.Join(os.TempDir(), "streamlab")
	}
	if err := os.MkdirAll(filepath.Join(root, "media"), 0o755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Join(root, "artifacts"), 0o755); err != nil {
		return nil, err
	}
	return &LocalStore{root: root, videos: make(map[string]*video.Video)}, nil
}

func (s *LocalStore) Root() string { return s.root }

func (s *LocalStore) CreateUpload(id, title, filename, contentType string, src io.Reader) (*video.Video, error) {
	if title == "" {
		title = strings.TrimSuffix(filename, filepath.Ext(filename))
	}
	path := s.MediaPath(id)
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	size, copyErr := io.Copy(f, src)
	closeErr := f.Close()
	if copyErr != nil {
		return nil, copyErr
	}
	if closeErr != nil {
		return nil, closeErr
	}
	now := time.Now().UTC()
	v := &video.Video{ID: id, Title: title, Filename: filename, ContentType: contentType, Size: size, Status: video.StatusProcessing, Source: "upload", Protocol: "MP4", Metadata: map[string]string{}, Variants: defaultVariants(), Manifests: map[string]string{}, CreatedAt: now, UpdatedAt: now}
	s.mu.Lock()
	s.videos[id] = v
	s.mu.Unlock()
	return clone(v), nil
}

func (s *LocalStore) Seed(v video.Video) {
	if v.Metadata == nil {
		v.Metadata = map[string]string{}
	}
	if v.Variants == nil {
		v.Variants = defaultVariants()
	}
	if v.Manifests == nil {
		v.Manifests = map[string]string{}
	}
	s.mu.Lock()
	s.videos[v.ID] = clone(&v)
	s.mu.Unlock()
}

func (s *LocalStore) Get(id string) (*video.Video, error) {
	s.mu.RLock()
	v, ok := s.videos[id]
	if !ok {
		s.mu.RUnlock()
		return nil, ErrNotFound
	}
	copy := clone(v)
	s.mu.RUnlock()
	return copy, nil
}

func (s *LocalStore) List() []*video.Video {
	s.mu.RLock()
	items := make([]*video.Video, 0, len(s.videos))
	for _, v := range s.videos {
		items = append(items, clone(v))
	}
	s.mu.RUnlock()
	sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.Before(items[j].CreatedAt) })
	return items
}

func (s *LocalStore) Update(id string, update func(*video.Video)) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.videos[id]
	if !ok {
		return ErrNotFound
	}
	update(v)
	v.UpdatedAt = time.Now().UTC()
	return nil
}

func (s *LocalStore) MediaPath(id string) string { return filepath.Join(s.root, "media", id+".bin") }

func (s *LocalStore) ArtifactPath(id, name string) string {
	dir := filepath.Join(s.root, "artifacts", id)
	_ = os.MkdirAll(dir, 0o755)
	return filepath.Join(dir, filepath.Base(name))
}

func (s *LocalStore) HasMedia(id string) bool {
	_, err := os.Stat(s.MediaPath(id))
	return err == nil
}

func (s *LocalStore) OpenMedia(id string) (*os.File, os.FileInfo, error) {
	f, err := os.Open(s.MediaPath(id))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, ErrNotFound
		}
		return nil, nil, err
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, nil, err
	}
	return f, info, nil
}

func (s *LocalStore) WriteArtifact(id, name string, data []byte) error {
	if err := os.WriteFile(s.ArtifactPath(id, name), data, 0o644); err != nil {
		return err
	}
	return nil
}

func (s *LocalStore) ReadArtifact(id, name string) ([]byte, error) {
	return os.ReadFile(s.ArtifactPath(id, name))
}

func (s *LocalStore) ArtifactExists(id, name string) bool {
	_, err := os.Stat(s.ArtifactPath(id, name))
	return err == nil
}

func defaultVariants() []video.Variant {
	return []video.Variant{{Name: "1080p", Width: 1920, Height: 1080, Bitrate: 5800000, VideoCodec: "h264", AudioCodec: "aac"}, {Name: "720p", Width: 1280, Height: 720, Bitrate: 3000000, VideoCodec: "h264", AudioCodec: "aac"}, {Name: "480p", Width: 854, Height: 480, Bitrate: 1200000, VideoCodec: "h264", AudioCodec: "aac"}}
}

func clone(v *video.Video) *video.Video {
	c := *v
	c.Metadata = map[string]string{}
	for k, val := range v.Metadata {
		c.Metadata[k] = val
	}
	c.Variants = append([]video.Variant(nil), v.Variants...)
	c.Manifests = map[string]string{}
	for k, val := range v.Manifests {
		c.Manifests[k] = val
	}
	return &c
}

func ValidateArtifactName(name string) error {
	if name == "" || filepath.Base(name) != name || strings.Contains(name, "..") {
		return fmt.Errorf("invalid artifact name")
	}
	return nil
}
