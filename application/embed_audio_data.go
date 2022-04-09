package application

// TODO create image infra service.
import (
	"context"
	"errors"
	"fmt"
	"github.com/EdlinOrg/prominentcolor"
	"github.com/hectorgabucio/taterubot-dc/domain"
	"github.com/hectorgabucio/taterubot-dc/domain/discord"
	"github.com/hectorgabucio/taterubot-dc/kit/event"
	"image"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"
)

type AddMetadataOnAudioSent struct {
	discord       discord.Client
	durationText  string
	fsRepo        domain.FileRepository
	voiceDataRepo domain.VoiceDataRepository
	decoder       domain.MP3Decoder
	bus           event.Bus
}

func NewAddMetadataOnAudioSent(discord discord.Client, durationText string, fsRepo domain.FileRepository, voiceDataRepo domain.VoiceDataRepository, decoder domain.MP3Decoder, bus event.Bus) *AddMetadataOnAudioSent {
	return &AddMetadataOnAudioSent{discord: discord, durationText: durationText, fsRepo: fsRepo, voiceDataRepo: voiceDataRepo, decoder: decoder, bus: bus}
}

func (handler *AddMetadataOnAudioSent) Handle(ctx context.Context, evt event.Event) error {
	audioSentEvt, ok := evt.(domain.AudioSentEvent)
	if !ok {
		return errors.New("unexpected event")
	}
	log.Println("Going to handle event", audioSentEvt.ID(), "aggregate id", audioSentEvt.AggregateID(), "file", audioSentEvt.Mp3Fullname)

	dominantColor := handler.getDominantAvatarColor(audioSentEvt.UserAvatarURL, audioSentEvt.FileName)
	t := handler.getDuration(audioSentEvt.Mp3Fullname)
	seconds := int(t)
	newEmbed := discord.MessageEmbed{
		Title:     audioSentEvt.Username,
		Timestamp: time.Now().Format(time.RFC3339),
		Color:     dominantColor,
		Thumbnail: audioSentEvt.UserAvatarURL,
		Fields: []*discord.MessageEmbedField{
			{
				Name:  handler.durationText,
				Value: formatSeconds(seconds),
			},
		},
	}

	err := handler.discord.SetEmbed(audioSentEvt.ChannelID, audioSentEvt.AggregateID(), newEmbed)
	if err != nil {
		log.Println(err)
		return fmt.Errorf("err setting embed in message, %w", err)
	}

	voiceData := domain.VoiceData{
		GuildID:   audioSentEvt.GuildID,
		ID:        audioSentEvt.ID(),
		Timestamp: audioSentEvt.MOccurredOn,
		Name:      audioSentEvt.FileName,
		UserID:    audioSentEvt.UserID,
		Duration:  seconds,
	}
	log.Println("saving voice data", voiceData)
	if err := handler.voiceDataRepo.Save(voiceData); err != nil {
		log.Println("err saving voice data", err)
	}
	go func() {
		err := handler.bus.Publish(ctx, []event.Event{domain.NewDoneProcessingFilesEvent(audioSentEvt.FileName)})
		if err != nil {
			log.Println(err)
		}
	}()
	return nil
}

func (handler *AddMetadataOnAudioSent) getDominantAvatarColor(url string, fileName string) int {
	err := handler.downloadFile(url, fmt.Sprintf("%s.png", fileName))
	if err != nil {
		log.Println(err)
		return 0
	}
	color, err := handler.prominentColor(fileName)
	if err != nil {
		fmt.Printf("couldnt get prominent color: %v", err)
		return 0
	}
	return color
}

func (handler *AddMetadataOnAudioSent) prominentColor(fileName string) (int, error) {
	// Step 1: Load the image
	img, err := handler.loadImage(fmt.Sprintf("%s.png", fileName))
	if err != nil {
		return 0, fmt.Errorf("failed to load image: %w", err)
	}

	// Step 2: Process it
	colours, err := prominentcolor.Kmeans(img)
	if err != nil {
		return 0, fmt.Errorf("failed to process image: %w", err)
	}

	for _, colour := range colours {
		value, err := strconv.ParseInt(colour.AsString(), 16, 64)
		if err != nil {
			return 0, fmt.Errorf("error parsing color string to int, %w", err)
		}
		if value > 0 && value <= math.MaxInt {
			return int(value), nil
		}
		return 0, nil
	}
	return 0, errors.New("couldnt get any dominant color")
}

func (handler *AddMetadataOnAudioSent) loadImage(fileInput string) (image.Image, error) {
	f, err := handler.fsRepo.Open(fileInput)
	if err != nil {
		return nil, fmt.Errorf("err opening image file, %w", err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Println("err closing image file", err)
		}
	}(f)
	img, _, err := image.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("err decoding image, %w", err)
	}
	return img, nil
}

func (handler *AddMetadataOnAudioSent) downloadFile(url, fileName string) error {
	response, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file, %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Println("error closing http response body", err)
		}
	}(response.Body)

	if response.StatusCode != 200 {
		return errors.New("received non 200 response code")
	}
	file, err := handler.fsRepo.CreateEmpty(fileName)
	if err != nil {
		return fmt.Errorf("failed to create empty file to write response, %w", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println("error closing file", err)
		}
	}(file)
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return fmt.Errorf("failed to copy response body to file, %w", err)
	}

	return nil
}

func (handler *AddMetadataOnAudioSent) getDuration(fileName string) float64 {
	file1, err := handler.fsRepo.Open(fileName)
	if err != nil {
		return 0
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file1)

	duration := handler.decoder.GetDuration(file1)
	return duration
}

func formatSeconds(inSeconds int) string {
	minutes := inSeconds / 60
	seconds := inSeconds % 60
	str := fmt.Sprintf("%dm:%02ds", minutes, seconds)
	return str
}
