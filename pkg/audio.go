package audio

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
)

// ConvertToPCM converts the audio data in bytes to int16 PCM format.
func ConvertToPCM(data []byte) ([]int16, error) {
	if len(data)%2 != 0 {
		return nil, fmt.Errorf("audio data length is not even, cannot convert to int16 PCM")
	}

	buf := bytes.NewReader(data)
	pcm := make([]int16, len(data)/2)
	if err := binary.Read(buf, binary.LittleEndian, &pcm); err != nil {
		return nil, fmt.Errorf("error converting to PCM: %v", err)
	}

	return pcm, nil
}

// GetOpusCompatibleFrames splits the PCM data into frames of fixed sizes, suitable for encoding with Opus codec.
func GetOpusCompatibleFrames(pcm []int16, channels, sampleRate int) ([][]int16, error) {
	var frames [][]int16
	// allowed frame sizes for Opus codec
	frameSizesMs := []float32{2.5, 5, 10, 20, 40, 60}

	for len(pcm) > 0 {
		// Find the largest frame size that fits the remaining PCM data
		var chosenFrameSize int
		for _, frameSizeMs := range frameSizesMs {
			samplesPerFrame := int(math.Round(float64(frameSizeMs) / 1000 * float64(sampleRate) * float64(channels)))
			if samplesPerFrame <= len(pcm) {
				chosenFrameSize = samplesPerFrame
			} else {
				break
			}
		}

		// If no valid frame size is found, break the loop
		if chosenFrameSize == 0 {
			break
		}

		// Append the chosen frame to the frames slice and update pcmData
		frames = append(frames, pcm[:chosenFrameSize])
		pcm = pcm[chosenFrameSize:]
	}

	if len(pcm) > 0 {
		// TODO: what to do with the remaining PCM data?
		fmt.Println("remaining PCM data:", len(pcm))
	}

	return frames, nil
}

func GetFrameSizeInMS(frame []int16, channels, sampleRate int) (frameSizeMs float32, err error) {
	if len(frame) == 0 {
		return 0, fmt.Errorf("empty frame")
	}

	frameSize := len(frame)
	frameSizeMs = float32(frameSize) / float32(channels) * 1000 / float32(sampleRate)
	switch frameSizeMs {
	// allowed frame sizes in opus
	case 2.5, 5, 10, 20, 40, 60:
		return frameSizeMs, nil
	default:
		return 0, fmt.Errorf("illegal frame size: %d bytes (%f ms)", frameSize, frameSizeMs)
	}
}
