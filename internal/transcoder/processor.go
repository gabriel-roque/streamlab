package transcoder

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strconv"

	"github.com/streamlab/streamlab/internal/storage"
	"github.com/streamlab/streamlab/internal/video"
)

type Processor struct {
	Store *storage.LocalStore
}

func (p *Processor) Process(ctx context.Context, job video.Job) error {
	v, err := p.Store.Get(job.VideoID)
	if err != nil {
		return err
	}
	if !p.Store.HasMedia(v.ID) {
		return errors.New("source media is not local")
	}

	metadata := map[string]string{"pipeline": "fixture", "hls": "fixture", "dash": "fixture"}
	if probe, ok := p.probe(ctx, p.Store.MediaPath(v.ID)); ok {
		metadata["pipeline"] = "ffprobe+fixture"
		for key, value := range probe {
			metadata[key] = value
		}
	}

	// A valid media file gets a real HLS playlist when ffmpeg is installed. The
	// fixture path below keeps uploads of arbitrary bytes and local development usable.
	ffmpegReady := p.tryFFmpeg(ctx, v.ID)
	if ffmpegReady {
		metadata["pipeline"] = "ffmpeg"
		metadata["hls"] = "ffmpeg"
	}
	hls := []byte(fixtureHLS(v))
	if ffmpegReady {
		if generated, readErr := p.Store.ReadArtifact(v.ID, "ffmpeg.m3u8"); readErr == nil {
			hls = generated
		}
	}
	dash := fixtureDASH(v)
	if err := p.Store.WriteArtifact(v.ID, "hls.m3u8", hls); err != nil {
		return err
	}
	if err := p.Store.WriteArtifact(v.ID, "manifest.mpd", []byte(dash)); err != nil {
		return err
	}
	if !ffmpegReady {
		if err := p.Store.WriteArtifact(v.ID, "segment-000.ts", nil); err != nil {
			return err
		}
	}
	return p.Store.Update(v.ID, func(item *video.Video) {
		item.Status = video.StatusReady
		item.Protocol = "HLS"
		item.Metadata = metadata
		item.Manifests = map[string]string{"hls": "hls.m3u8", "dash": "manifest.mpd"}
		if value, ok := metadata["duration_seconds"]; ok {
			item.Duration, _ = strconv.ParseFloat(value, 64)
		}
	})
}

type probeResult struct {
	Streams []struct {
		CodecType string `json:"codec_type"`
		CodecName string `json:"codec_name"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
	} `json:"streams"`
	Format struct {
		Duration string `json:"duration"`
		Format   string `json:"format_name"`
	} `json:"format"`
}

func (p *Processor) probe(ctx context.Context, source string) (map[string]string, bool) {
	if _, err := exec.LookPath("ffprobe"); err != nil {
		return nil, false
	}
	command := exec.CommandContext(ctx, "ffprobe", "-v", "quiet", "-print_format", "json", "-show_streams", "-show_format", source)
	out, err := command.Output()
	if err != nil {
		return nil, false
	}
	var result probeResult
	if json.Unmarshal(out, &result) != nil {
		return nil, false
	}
	metadata := map[string]string{"format": result.Format.Format}
	if duration, err := strconv.ParseFloat(result.Format.Duration, 64); err == nil {
		metadata["duration_seconds"] = strconv.FormatFloat(duration, 'f', 3, 64)
	}
	for _, stream := range result.Streams {
		if stream.CodecType == "video" {
			metadata["video_codec"] = stream.CodecName
			metadata["width"] = strconv.Itoa(stream.Width)
			metadata["height"] = strconv.Itoa(stream.Height)
		}
		if stream.CodecType == "audio" {
			metadata["audio_codec"] = stream.CodecName
		}
	}
	return metadata, true
}

func (p *Processor) tryFFmpeg(ctx context.Context, id string) bool {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return false
	}
	output := p.Store.ArtifactPath(id, "ffmpeg.m3u8")
	segment := p.Store.ArtifactPath(id, "segment-%03d.ts")
	command := exec.CommandContext(ctx, "ffmpeg", "-hide_banner", "-loglevel", "error", "-y", "-i", p.Store.MediaPath(id), "-c:v", "libx264", "-pix_fmt", "yuv420p", "-c:a", "aac", "-ac", "2", "-f", "hls", "-hls_time", "6", "-hls_list_size", "0", "-hls_segment_filename", segment, output)
	return command.Run() == nil
}

func fixtureHLS(v *video.Video) string {
	duration := 6
	if v.Duration > 0 && v.Duration < 6 {
		duration = int(v.Duration)
	}
	return fmt.Sprintf("#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-TARGETDURATION:%d\n#EXT-X-MEDIA-SEQUENCE:0\n#EXTINF:%d.000,\nsegment-000.ts\n#EXT-X-ENDLIST\n", duration, duration)
}

func fixtureDASH(v *video.Video) string {
	duration := "PT6S"
	if v.Duration > 0 {
		duration = "PT" + strconv.FormatFloat(v.Duration, 'f', 3, 64) + "S"
	}
	return fmt.Sprintf("<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<MPD xmlns=\"urn:mpeg:dash:schema:mpd:2011\" type=\"static\" mediaPresentationDuration=\"%s\" minBufferTime=\"PT1.5S\"><Period><AdaptationSet contentType=\"video\" mimeType=\"video/mp4\"><Representation id=\"fixture-720p\" width=\"1280\" height=\"720\" bandwidth=\"3000000\" /></AdaptationSet></Period></MPD>\n", duration)
}
