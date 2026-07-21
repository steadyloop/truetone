// internal/transcode/transcode.go
// Runs ffmpeg to turn one input video into an HLS rendition for streaming.
package transcode

import (
	"fmt"
	"math"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// Transcode reads inputFile and writes an HLS rendition — master.m3u8,
// stream.m3u8, segment_*.m4s — into outDir. outDir must already exist.
func ToHLS(inputFile, outDir string) error {
	cmd := exec.Command("ffmpeg",
		"-i", inputFile,
		"-c:v", "libx264",
		"-crf", "21",
		"-preset", "veryfast",
		"-pix_fmt", "yuv420p",
		"-force_key_frames", "expr:gte(t,n_forced*6)",
		"-c:a", "aac",
		"-b:a", "128k",
		"-f", "hls",
		"-hls_segment_type", "fmp4",
		"-hls_time", "6",
		"-hls_playlist_type", "vod",
		"-hls_flags", "independent_segments",
		"-master_pl_name", "master.m3u8",
		"-hls_segment_filename", filepath.Join(outDir, "segment_%03d.m4s"),
		filepath.Join(outDir, "stream.m3u8"),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg failed: %w\n%s", err, output)
	}
	return nil
}

// Thumbnail extracts a single representative frame from inputFile into outFile,
// using ffmpeg's thumbnail filter to pick the frame least different from its
// surrounding batch — avoids black/intro/transition frames.
func Thumbnail(inputFile, outFile string) error {
	cmd := exec.Command("ffmpeg",
		"-i", inputFile,
		"-vf", "thumbnail,scale=320:-1",
		"-frames:v", "1",
		outFile,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg thumbnail failed: %w\n%s", err, output)
	}
	return nil
}

// CHECK
// Duration returns length of video in seconds (via ffprobe)
func Duration(inputFile string) (int, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		inputFile,
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("ffprobe failed: %w\n%s", err, output)
	}
	seconds, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		return 0, fmt.Errorf("parsing ffprobe output %q: %w", output, err)
	}
	return int(math.Round(seconds)), nil
}
