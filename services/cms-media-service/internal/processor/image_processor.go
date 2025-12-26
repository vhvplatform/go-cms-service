package processor

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ImageProcessor handles image compression and manipulation
type ImageProcessor struct {
	maxWidth  int
	maxHeight int
	quality   int
}

// NewImageProcessor creates a new image processor
func NewImageProcessor(maxWidth, maxHeight, quality int) *ImageProcessor {
	return &ImageProcessor{
		maxWidth:  maxWidth,
		maxHeight: maxHeight,
		quality:   quality,
	}
}

// CompressImage compresses an image using ImageMagick/FFmpeg
func (p *ImageProcessor) CompressImage(inputPath, outputPath string) (int64, error) {
	// Check if ImageMagick is available
	if _, err := exec.LookPath("convert"); err == nil {
		return p.compressWithImageMagick(inputPath, outputPath)
	}

	// Fallback to FFmpeg
	if _, err := exec.LookPath("ffmpeg"); err == nil {
		return p.compressWithFFmpeg(inputPath, outputPath)
	}

	// No processor available, just copy the file
	return p.copyFile(inputPath, outputPath)
}

// compressWithImageMagick uses ImageMagick to compress image
func (p *ImageProcessor) compressWithImageMagick(inputPath, outputPath string) (int64, error) {
	args := []string{
		inputPath,
		"-resize", fmt.Sprintf("%dx%d>", p.maxWidth, p.maxHeight),
		"-quality", fmt.Sprintf("%d", p.quality),
		"-strip", // Remove EXIF data
		outputPath,
	}

	cmd := exec.Command("convert", args...)
	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("imagemagick conversion failed: %w", err)
	}

	// Get output file size
	info, err := os.Stat(outputPath)
	if err != nil {
		return 0, err
	}

	return info.Size(), nil
}

// compressWithFFmpeg uses FFmpeg to compress image
func (p *ImageProcessor) compressWithFFmpeg(inputPath, outputPath string) (int64, error) {
	args := []string{
		"-i", inputPath,
		"-vf", fmt.Sprintf("scale='min(%d,iw)':min'(%d,ih)':force_original_aspect_ratio=decrease", p.maxWidth, p.maxHeight),
		"-q:v", fmt.Sprintf("%d", 100-p.quality),
		outputPath,
		"-y",
	}

	cmd := exec.Command("ffmpeg", args...)
	if err := cmd.Run(); err != nil {
		return 0, fmt.Errorf("ffmpeg conversion failed: %w", err)
	}

	info, err := os.Stat(outputPath)
	if err != nil {
		return 0, err
	}

	return info.Size(), nil
}

// copyFile copies a file without processing
func (p *ImageProcessor) copyFile(src, dst string) (int64, error) {
	input, err := os.ReadFile(src)
	if err != nil {
		return 0, err
	}

	if err := os.WriteFile(dst, input, 0644); err != nil {
		return 0, err
	}

	return int64(len(input)), nil
}

// GetImageDimensions gets image dimensions using FFprobe
func (p *ImageProcessor) GetImageDimensions(imagePath string) (width, height int, err error) {
	cmd := exec.Command("ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries", "stream=width,height",
		"-of", "csv=s=x:p=0",
		imagePath,
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}

	dimensions := strings.TrimSpace(string(output))
	_, err = fmt.Sscanf(dimensions, "%dx%d", &width, &height)
	return width, height, err
}
