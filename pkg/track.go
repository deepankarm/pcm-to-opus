package audio

import (
	"context"

	lksdk "github.com/livekit/server-sdk-go/v2"
	"github.com/pion/webrtc/v3"
	"go.uber.org/zap"
)

type PCMTrack struct {
	URL                 string
	RoomName            string
	ParticipantName     string
	ParticipantIdentity string
	LivekitAPIKey       string
	LivekitSecretKey    string

	SampleRate         int
	Channels           int
	Logger             *zap.Logger
	Room               *lksdk.Room
	Track              *lksdk.LocalTrack
	PCM16SampleChannel chan PCM16Sample
}

func (ln *PCMTrack) Connect(ctx context.Context) error {
	connInfo := lksdk.ConnectInfo{
		APIKey:              ln.LivekitAPIKey,
		APISecret:           ln.LivekitSecretKey,
		RoomName:            ln.RoomName,
		ParticipantName:     ln.ParticipantName,
		ParticipantIdentity: ln.ParticipantIdentity,
	}

	room, err := lksdk.ConnectToRoom(ln.URL, connInfo, nil)
	if err != nil {
		ln.Logger.Error("Failed to connect to room", zap.Error(err))
		return err
	}
	ln.Logger.Debug("Connected to room", zap.String("room", ln.RoomName), zap.String("participant", ln.ParticipantName))
	ln.Room = room

	return nil
}

func (t *PCMTrack) PublishTrack() error {
	track, err := lksdk.NewLocalSampleTrack(webrtc.RTPCodecCapability{
		MimeType:  webrtc.MimeTypeOpus,
		ClockRate: uint32(t.SampleRate),
		Channels:  uint16(t.Channels),
	})
	if err != nil {
		t.Logger.Error("failed to create new track", zap.Error(err))
		return err
	}

	// Create a sample provider to provide the audio samples to the track
	provider, err := NewSpeakerSampleProvider(t.PCM16SampleChannel, t.SampleRate, t.Channels, t.Logger)
	if err != nil {
		t.Logger.Error("Failed to create sample provider", zap.Error(err))
		return err
	}

	// StartWrite is necessary so packetizer is attached to the track
	if err = track.StartWrite(provider, func() {
		t.Logger.Info("Track write stopped")
	}); err != nil {
		t.Logger.Error("Failed to start writing to track", zap.Error(err))
		return err
	}

	if _, err = t.Room.LocalParticipant.PublishTrack(
		track,
		&lksdk.TrackPublicationOptions{
			Name: t.Room.LocalParticipant.Identity(),
		},
	); err != nil {
		t.Logger.Error("failed to publish track", zap.Error(err))
		return err
	}

	t.Track = track
	t.Logger.Info("Track published", zap.String("trackID", t.Track.ID()))
	return nil
}

func NewPCMTrack(
	url, livekitAPIKey, livekitSecretKey,
	roomName, participantName, participantIdentity string,
	sampleRate, channels int,
	pcmSampleChannel chan PCM16Sample,
) *PCMTrack {
	return &PCMTrack{
		URL:                 url,
		RoomName:            roomName,
		ParticipantName:     participantName,
		ParticipantIdentity: participantIdentity,
		LivekitAPIKey:       livekitAPIKey,
		LivekitSecretKey:    livekitSecretKey,
		SampleRate:          sampleRate,
		Channels:            channels,
		Logger:              logger,
		PCM16SampleChannel:  pcmSampleChannel,
	}
}
