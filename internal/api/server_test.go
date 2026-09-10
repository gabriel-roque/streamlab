package api

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	server, err := New(Config{StorageRoot: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(server.Close)
	return server
}

func TestParseRange(t *testing.T) {
	tests := []struct {
		input      string
		start, end int64
		partial    bool
	}{
		{"bytes=0-4", 0, 4, true}, {"bytes=5-", 5, 9, true}, {"bytes=-3", 7, 9, true}, {"", 0, 9, false},
	}
	for _, test := range tests {
		start, end, partial, err := ParseRange(test.input, 10)
		if err != nil || start != test.start || end != test.end || partial != test.partial {
			t.Errorf("ParseRange(%q) = %d-%d partial=%v err=%v", test.input, start, end, partial, err)
		}
	}
	if _, _, _, err := ParseRange("bytes=10-20", 10); err == nil {
		t.Error("expected invalid range")
	}
}

func TestUploadStatusAndRange(t *testing.T) {
	server := newTestServer(t)
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	_ = writer.WriteField("title", "Integration video")
	part, err := writer.CreateFormFile("file", "sample.mp4")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = part.Write([]byte("0123456789"))
	_ = writer.Close()
	req := httptest.NewRequest(http.MethodPost, "/videos", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	res := httptest.NewRecorder()
	server.Handler().ServeHTTP(res, req)
	if res.Code != http.StatusAccepted {
		t.Fatalf("upload status = %d: %s", res.Code, res.Body.String())
	}
	var uploaded struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(res.Body).Decode(&uploaded)
	if uploaded.ID == "" {
		t.Fatal("upload did not return id")
	}

	var statusRes *httptest.ResponseRecorder
	for i := 0; i < 30; i++ {
		statusRes = httptest.NewRecorder()
		server.Handler().ServeHTTP(statusRes, httptest.NewRequest(http.MethodGet, "/videos/"+uploaded.ID+"/status", nil))
		if strings.Contains(statusRes.Body.String(), `"READY"`) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if !strings.Contains(statusRes.Body.String(), `"READY"`) {
		t.Fatalf("video did not become ready: %s", statusRes.Body.String())
	}

	rangeReq := httptest.NewRequest(http.MethodGet, "/videos/"+uploaded.ID+"/stream", nil)
	rangeReq.Header.Set("Range", "bytes=2-5")
	rangeRes := httptest.NewRecorder()
	server.Handler().ServeHTTP(rangeRes, rangeReq)
	if rangeRes.Code != http.StatusPartialContent || rangeRes.Header().Get("Content-Range") != "bytes 2-5/10" || rangeRes.Body.String() != "2345" {
		t.Fatalf("range response: code=%d range=%q body=%q", rangeRes.Code, rangeRes.Header().Get("Content-Range"), rangeRes.Body.String())
	}

	fullRes := httptest.NewRecorder()
	server.Handler().ServeHTTP(fullRes, httptest.NewRequest(http.MethodGet, "/videos/"+uploaded.ID+"/stream", nil))
	if fullRes.Code != http.StatusOK || fullRes.Body.String() != "0123456789" || fullRes.Header().Get("Accept-Ranges") != "bytes" {
		t.Fatalf("full response: code=%d body=%q ranges=%q", fullRes.Code, fullRes.Body.String(), fullRes.Header().Get("Accept-Ranges"))
	}

	headReq := httptest.NewRequest(http.MethodHead, "/videos/"+uploaded.ID+"/stream", nil)
	headReq.Header.Set("Range", "bytes=2-5")
	headRes := httptest.NewRecorder()
	server.Handler().ServeHTTP(headRes, headReq)
	if headRes.Code != http.StatusPartialContent || headRes.Body.Len() != 0 || headRes.Header().Get("Content-Length") != "4" {
		t.Fatalf("head response: code=%d body=%q length=%q", headRes.Code, headRes.Body.String(), headRes.Header().Get("Content-Length"))
	}

	invalidReq := httptest.NewRequest(http.MethodGet, "/videos/"+uploaded.ID+"/stream", nil)
	invalidReq.Header.Set("Range", "bytes=10-20")
	invalidRes := httptest.NewRecorder()
	server.Handler().ServeHTTP(invalidRes, invalidReq)
	if invalidRes.Code != http.StatusRequestedRangeNotSatisfiable || invalidRes.Header().Get("Content-Range") != "bytes */10" {
		t.Fatalf("invalid range response: code=%d range=%q", invalidRes.Code, invalidRes.Header().Get("Content-Range"))
	}
}

func TestCreateThenUploadAndPlaybackContracts(t *testing.T) {
	server := newTestServer(t)

	createdRes := doJSON(server, http.MethodPost, "/videos", []byte(`{"title":"Created first","source":{"filename":"source.mov","contentType":"video/quicktime"}}`))
	if createdRes.StatusCode != http.StatusCreated {
		t.Fatalf("create status = %d", createdRes.StatusCode)
	}
	var created struct {
		ID     string `json:"id"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(createdRes.Body).Decode(&created); err != nil {
		t.Fatal(err)
	}
	_ = createdRes.Body.Close()
	if created.ID == "" || created.Status != "PROCESSING" {
		t.Fatalf("created video = %+v", created)
	}
	processingStatus := httptest.NewRecorder()
	server.Handler().ServeHTTP(processingStatus, httptest.NewRequest(http.MethodGet, "/videos/"+created.ID+"/status", nil))
	if processingStatus.Code != http.StatusOK || !strings.Contains(processingStatus.Body.String(), `"status":"PROCESSING"`) {
		t.Fatalf("processing status response: code=%d body=%s", processingStatus.Code, processingStatus.Body.String())
	}

	processingPlayback := httptest.NewRecorder()
	server.Handler().ServeHTTP(processingPlayback, httptest.NewRequest(http.MethodGet, "/videos/"+created.ID+"/playback", nil))
	if processingPlayback.Code != http.StatusConflict {
		t.Fatalf("processing playback status = %d", processingPlayback.Code)
	}

	body, contentType := multipartBody(t, "uploaded.mov", "0123456789")
	uploadReq := httptest.NewRequest(http.MethodPost, "/videos/"+created.ID+"/upload", body)
	uploadReq.Header.Set("Content-Type", contentType)
	uploadRes := httptest.NewRecorder()
	server.Handler().ServeHTTP(uploadRes, uploadReq)
	if uploadRes.Code != http.StatusAccepted || uploadRes.Header().Get("Location") != "/videos/"+created.ID {
		t.Fatalf("follow-up upload: code=%d location=%q body=%s", uploadRes.Code, uploadRes.Header().Get("Location"), uploadRes.Body.String())
	}

	waitUntilReady(t, server, created.ID)
	readyStatus := httptest.NewRecorder()
	server.Handler().ServeHTTP(readyStatus, httptest.NewRequest(http.MethodGet, "/videos/"+created.ID+"/status", nil))
	if readyStatus.Code != http.StatusOK || !strings.Contains(readyStatus.Body.String(), `"status":"READY"`) {
		t.Fatalf("ready status response: code=%d body=%s", readyStatus.Code, readyStatus.Body.String())
	}
	playback := httptest.NewRecorder()
	server.Handler().ServeHTTP(playback, httptest.NewRequest(http.MethodGet, "/videos/"+created.ID+"/playback", nil))
	if playback.Code != http.StatusOK || !strings.Contains(playback.Body.String(), `"videoId":"`+created.ID+`"`) || !strings.Contains(playback.Body.String(), `"protocol":"HLS"`) {
		t.Fatalf("playback response: code=%d body=%s", playback.Code, playback.Body.String())
	}

	for _, test := range []struct {
		format      string
		contentType string
		body        string
	}{
		{format: "hls", contentType: "application/vnd.apple.mpegurl", body: "#EXTM3U"},
		{format: "dash", contentType: "application/dash+xml", body: "<MPD"},
	} {
		manifest := httptest.NewRecorder()
		server.Handler().ServeHTTP(manifest, httptest.NewRequest(http.MethodGet, "/videos/"+created.ID+"/playback/"+test.format, nil))
		if manifest.Code != http.StatusOK || manifest.Header().Get("Content-Type") != test.contentType || !strings.Contains(manifest.Body.String(), test.body) {
			t.Fatalf("%s manifest: code=%d content-type=%q body=%q", test.format, manifest.Code, manifest.Header().Get("Content-Type"), manifest.Body.String())
		}
	}
}

func TestHealthAliasesAndStorageEnvironment(t *testing.T) {
	root := t.TempDir()
	t.Setenv("STREAMLAB_STORAGE_ROOT", root)
	t.Setenv("STORAGE_PATH", t.TempDir())
	server, err := New(Config{DisableSeed: true})
	if err != nil {
		t.Fatal(err)
	}
	defer server.Close()
	if server.Store.Root() != root {
		t.Fatalf("storage root = %q, want %q", server.Store.Root(), root)
	}

	for _, path := range []string{"/health", "/healthz"} {
		res := httptest.NewRecorder()
		server.Handler().ServeHTTP(res, httptest.NewRequest(http.MethodGet, path, nil))
		if res.Code != http.StatusOK || !strings.Contains(res.Body.String(), `"status":"ok"`) {
			t.Fatalf("%s response: code=%d body=%s", path, res.Code, res.Body.String())
		}
	}
}

func TestPlaybackAndTelemetry(t *testing.T) {
	server := newTestServer(t)
	payload, _ := json.Marshal(map[string]string{"video_id": "big-buck-bunny", "user_id": "test-user"})
	res := doJSON(server, http.MethodPost, "/playback/sessions", payload)
	if res.StatusCode != http.StatusCreated {
		t.Fatalf("session status = %d", res.StatusCode)
	}
	var session struct {
		ID string `json:"id"`
	}
	_ = json.NewDecoder(res.Body).Decode(&session)
	_ = res.Body.Close()
	event, _ := json.Marshal(map[string]any{"video_id": "big-buck-bunny", "session_id": session.ID, "type": "play", "position": 0})
	res = doJSON(server, http.MethodPost, "/telemetry/playback", event)
	if res.StatusCode != http.StatusAccepted {
		t.Fatalf("event status = %d", res.StatusCode)
	}
	metrics := httptest.NewRecorder()
	server.Handler().ServeHTTP(metrics, httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if !strings.Contains(metrics.Body.String(), "streamlab_playback_events_total 1") {
		t.Fatalf("metrics missing event: %s", metrics.Body.String())
	}
	playback := httptest.NewRecorder()
	server.Handler().ServeHTTP(playback, httptest.NewRequest(http.MethodGet, "/videos/big-buck-bunny/playback", nil))
	if playback.Code != http.StatusOK || !strings.Contains(playback.Body.String(), "1080p") || !strings.Contains(playback.Body.String(), `"protocol":"MP4"`) || !strings.Contains(playback.Body.String(), "source_url") {
		t.Fatalf("playback response: %s", playback.Body.String())
	}
	manifest := httptest.NewRecorder()
	server.Handler().ServeHTTP(manifest, httptest.NewRequest(http.MethodGet, "/videos/big-buck-bunny/playback/hls", nil))
	if manifest.Code != http.StatusNotFound {
		t.Fatalf("remote catalog should not claim a local manifest: %d", manifest.Code)
	}
}

func doJSON(server *Server, method, path string, payload []byte) *http.Response {
	request := httptest.NewRequest(method, path, bytes.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	return recorderResult(server.Handler(), request)
}

// httptest.ResponseRecorder does not expose a response conversion helper on old
// Go versions; this tiny adapter keeps the test request code readable.
func recorderResult(handler http.Handler, request *http.Request) *http.Response {
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder.Result()
}

func multipartBody(t *testing.T, filename, contents string) (*bytes.Buffer, string) {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte(contents)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return &body, writer.FormDataContentType()
}

func waitUntilReady(t *testing.T, server *Server, id string) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		res := httptest.NewRecorder()
		server.Handler().ServeHTTP(res, httptest.NewRequest(http.MethodGet, "/videos/"+id+"/status", nil))
		if strings.Contains(res.Body.String(), `"READY"`) {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("video %s did not become ready", id)
}
