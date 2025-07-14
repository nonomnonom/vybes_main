package media

import (
	"fmt"
	"os/exec"

	"github.com/rs/zerolog/log"
)

// Processor defines the interface for media processing operations.
type Processor interface {
	GenerateThumbnail(inputPath, outputPath string) error
}

type ffmpegProcessor struct {
	ffmpegPath string
}

// NewFFmpegProcessor creates a new FFmpeg processor.
// It checks if the ffmpeg command exists in the system's PATH.
func NewFFmpegProcessor() (Processor, error) {
	path, err := exec.LookPath("ffmpeg")
	if err != nil {
		return nil, fmt.Errorf("ffmpeg not found in PATH: %w", err)
	}
	return &ffmpegProcessor{ffmpegPath: path}, nil
}

// GenerateThumbnail extracts a single frame from a video file to use as a thumbnail.
// It extracts the frame at the 1-second mark.
func (p *ffmpegProcessor) GenerateThumbnail(inputPath, outputPath string) error {
	// Command: ffmpeg -i <input> -ss 00:00:01 -vframes 1 <output>
	cmd := exec.Command(p.ffmpegPath, "-i", inputPath, "-ss", "00:00:01", "-vframes", "1", outputPath)

	// Run the command and capture any output/error
	output, err := cmd.CombinedOutput()
	if err != nil {
		// In case of an error, it's useful to log the output from ffmpeg
		log.Error().Err(err).Str("ffmpeg_output", string(output)).Msg("FFmpeg command failed")
		return err
	}

	return nil
}