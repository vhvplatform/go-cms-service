package processor

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// VideoProcessor handles video processing operations
type VideoProcessor struct {
	outputDir string
}

// NewVideoProcessor creates a new video processor
func NewVideoProcessor(outputDir string) *VideoProcessor {
	return &VideoProcessor{
		outputDir: outputDir,
	}
}

// ConvertToHLS converts video to HLS format (m3u8)
func (vp *VideoProcessor) ConvertToHLS(inputPath string, outputName string) (string, []VideoResolution, error) {
	baseDir := filepath.Join(vp.outputDir, outputName)
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return "", nil, err
	}

	m3u8Path := filepath.Join(baseDir, "master.m3u8")

	// Generate multiple resolutions
	resolutions := []VideoResolution{
		{Name: "360p", Width: 640, Height: 360, Bitrate: "800k"},
		{Name: "720p", Width: 1280, Height: 720, Bitrate: "2500k"},
		{Name: "1080p", Width: 1920, Height: 1080, Bitrate: "5000k"},
	}

	var generatedResolutions []VideoResolution

	for _, res := range resolutions {
		outputPath := filepath.Join(baseDir, fmt.Sprintf("%s.m3u8", res.Name))

		args := []string{
			"-i", inputPath,
			"-vf", fmt.Sprintf("scale=%d:%d", res.Width, res.Height),
			"-c:v", "libx264",
			"-b:v", res.Bitrate,
			"-c:a", "aac",
			"-b:a", "128k",
			"-hls_time", "10",
			"-hls_list_size", "0",
			"-hls_segment_filename", filepath.Join(baseDir, fmt.Sprintf("%s_%%03d.ts", res.Name)),
			"-f", "hls",
			outputPath,
		}

		cmd := exec.Command("ffmpeg", args...)
		if err := cmd.Run(); err != nil {
			// Skip this resolution if conversion fails
			continue
		}

		generatedResolutions = append(generatedResolutions, res)
	}

	// Create master playlist
	if err := vp.createMasterPlaylist(m3u8Path, generatedResolutions); err != nil {
		return "", nil, err
	}

	return m3u8Path, generatedResolutions, nil
}

// createMasterPlaylist creates an HLS master playlist
func (vp *VideoProcessor) createMasterPlaylist(masterPath string, resolutions []VideoResolution) error {
	var content strings.Builder
	content.WriteString("#EXTM3U\n")
	content.WriteString("#EXT-X-VERSION:3\n\n")

	for _, res := range resolutions {
		bandwidth, _ := strconv.Atoi(strings.TrimSuffix(res.Bitrate, "k"))
		bandwidth *= 1000

		content.WriteString(fmt.Sprintf("#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%dx%d\n",
			bandwidth, res.Width, res.Height))
		content.WriteString(fmt.Sprintf("%s.m3u8\n\n", res.Name))
	}

	return os.WriteFile(masterPath, []byte(content.String()), 0644)
}

// ExtractThumbnail extracts a thumbnail from video
func (vp *VideoProcessor) ExtractThumbnail(videoPath, outputPath string, timeOffset int) error {
	args := []string{
		"-i", videoPath,
		"-ss", fmt.Sprintf("%d", timeOffset),
		"-vframes", "1",
		"-vf", "scale=640:360:force_original_aspect_ratio=decrease",
		"-y",
		outputPath,
	}

	cmd := exec.Command("ffmpeg", args...)
	return cmd.Run()
}

// GetVideoDuration gets video duration in seconds
func (vp *VideoProcessor) GetVideoDuration(videoPath string) (float64, error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		videoPath,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	durationStr := strings.TrimSpace(string(output))
	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		return 0, err
	}

	return duration, nil
}

// GetVideoInfo gets comprehensive video information
func (vp *VideoProcessor) GetVideoInfo(videoPath string) (width, height int, duration float64, err error) {
	// Get duration
	duration, err = vp.GetVideoDuration(videoPath)
	if err != nil {
		return 0, 0, 0, err
	}

	// Get dimensions
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=s=x:p=0",
		videoPath,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, 0, duration, err
	}

	dimensions := strings.TrimSpace(string(output))
	_, err = fmt.Sscanf(dimensions, "%dx%d", &width, &height)

	return width, height, duration, err
}

// VideoResolution represents a video resolution configuration
type VideoResolution struct {
	Name    string
	Width   int
	Height  int
	Bitrate string
}
