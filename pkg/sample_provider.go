package audio

import (
	"context"
	"fmt"
	"time"

	"github.com/pion/webrtc/v3/pkg/media"
	"go.uber.org/zap"
	"gopkg.in/hraban/opus.v2"
)

type PCM16Sample []int16

type SpeakerSampleProvider struct {
	framesCh   chan PCM16Sample
	encoder    *opus.Encoder
	bufferSize int
	channels   int
	sampleRate int
	logger     *zap.Logger
}

func (p *SpeakerSampleProvider) OnBind() error {
	return nil
}

func (p *SpeakerSampleProvider) OnUnbind() error {
	return nil
}

func (p *SpeakerSampleProvider) Close() error {
	return nil
}

func (p *SpeakerSampleProvider) NextSample(ctx context.Context) (media.Sample, error) {
	for {
		select {
		case <-ctx.Done():
			return media.Sample{}, nil
		case frame := <-p.framesCh:
			if len(frame) <= 0 {
				continue
			}
			encodedAudio := make([]byte, p.bufferSize)
			n, err := p.encoder.Encode(frame, encodedAudio)
			if err != nil {
				p.logger.Error("Failed to encode audio data", zap.String("error", err.Error()))
				continue
			}

			frameSizeMs, err := GetFrameSizeInMS(frame, p.channels, p.sampleRate)
			if err != nil {
				p.logger.Debug("Failed to get frame size", zap.String("error", err.Error()))
				continue
			}

			return media.Sample{
				Data:     encodedAudio[:n],
				Duration: time.Duration(frameSizeMs) * time.Millisecond,
			}, nil
		}
	}
}

func NewSpeakerSampleProvider(
	framesCh chan PCM16Sample,
	sampleRate, channels int,
	logger *zap.Logger,
) (*SpeakerSampleProvider, error) {
	enc, err := opus.NewEncoder(sampleRate, channels, opus.AppVoIP)
	if err != nil {
		return nil, fmt.Errorf("failed to create opus encoder: %v", err)
	}

	return &SpeakerSampleProvider{
		framesCh:   framesCh,
		encoder:    enc,
		sampleRate: sampleRate,
		bufferSize: 10000, // TODO: check this
		channels:   channels,
		logger:     logger,
	}, nil
}
