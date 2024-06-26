package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	audio "github.com/deepankarm/pcm-to-opus/pkg"
	"github.com/livekit/protocol/auth"
)

var LivekitRoomName = "dummy-room"

var LivekitPublisherName = "publisher"
var LivekitPublisherIdentity = "publisher-identity"

var LivekitListenerName = "listener"
var LivekitListenerIdentity = "listener-identity"

func main() {
	if len(os.Args) < 5 {
		fmt.Println("Usage: go run cmd/main.go <livekit-url> <livekit-api-key> <livekit-api-secret> <directory>")
		os.Exit(1)
	}

	url := os.Args[1]
	apiKey := os.Args[2]
	secretKey := os.Args[3]
	directory := os.Args[4]

	pcm16SampleChan := make(chan audio.PCM16Sample)
	pcmTrack := audio.NewPCMTrack(
		url,
		apiKey,
		secretKey,
		LivekitRoomName,
		LivekitPublisherName,
		LivekitPublisherIdentity,
		24000,
		1,
		pcm16SampleChan,
	)
	ctx := context.Background()
	err := pcmTrack.Connect(ctx)
	if err != nil {
		panic(err)
	}

	err = pcmTrack.PublishTrack()
	if err != nil {
		panic(err)
	}

	userURL := GetLivekitUserURL(url, apiKey, secretKey, LivekitRoomName, LivekitListenerName, LivekitListenerIdentity)
	log.Printf("Listener URL: %s\n", userURL)

	SendAudio(directory, pcm16SampleChan)
}

func GetLivekitToken(livekitAPIKey, livekitSecretKey, room, name, identity string) (string, error) {
	at := auth.NewAccessToken(livekitAPIKey, livekitSecretKey)
	grant := &auth.VideoGrant{
		RoomJoin: true,
		Room:     room,
	}
	at.AddGrant(grant).SetName(name).SetIdentity(identity).SetValidFor(time.Hour)
	return at.ToJWT()
}

func GetLivekitUserURL(url, livekitAPIKey, livekitSecretKey, room, name, identity string) string {
	token, err := GetLivekitToken(livekitAPIKey, livekitSecretKey, room, name, identity)
	if err != nil {
		log.Fatalf("Error getting Livekit token: %v", err)
	}
	return fmt.Sprintf("https://meet.livekit.io/custom?liveKitUrl=%s&token=%s", url, token)
}

func SendAudio(directory string, pcm16SampleChan chan audio.PCM16Sample) {
	// get all the .pcm files
	files, err := os.ReadDir(directory)
	if err != nil {
		log.Fatalf("failed to read directory")
		return
	}

	log.Print("Sleeping for 10 seconds to allow the listener to join before sending the audio files")
	time.Sleep(10 * time.Second)

	// read each file and send the audio to the PCM16SampleChannel
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".pcm") {
			continue
		}

		log.Printf("Sending audio from file: %s\n", file.Name())

		data, err := os.ReadFile(directory + file.Name())
		if err != nil {
			log.Printf("failed to read file: %s", file.Name())
			continue
		}

		pcm, err := audio.ConvertToPCM(data)
		if err != nil {
			fmt.Println("failed to convert to PCM")
			// TODO: what to do with the error?
			continue
		}

		frames, err := audio.GetOpusCompatibleFrames(pcm, 1, 24000)
		if err != nil {
			log.Print("failed to get frames")
			// TODO: what to do with the error?
			continue
		}

		// send the frames to PCM16SampleChannel, so that the track can write them
		for _, frame := range frames {
			if len(frame) <= 0 {
				log.Print("empty frame")
				continue
			}
			pcm16SampleChan <- frame
		}
	}
}
