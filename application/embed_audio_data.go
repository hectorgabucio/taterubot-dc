package application

import (
	"context"
	"errors"
	"fmt"
	"github.com/EdlinOrg/prominentcolor"
	"github.com/bwmarrin/discordgo"
	"github.com/hectorgabucio/taterubot-dc/domain"
	"github.com/hectorgabucio/taterubot-dc/kit/event"
	"log"
	"strconv"
	"time"
)

type AddMetadataOnAudioSent struct {
	session      *discordgo.Session
	durationText string
	basePath     string
}

func NewAddMetadataOnAudioSent(session *discordgo.Session, durationText string, basePath string) *AddMetadataOnAudioSent {
	return &AddMetadataOnAudioSent{session: session, durationText: durationText, basePath: basePath}
}

func (handler *AddMetadataOnAudioSent) Handle(_ context.Context, evt event.Event) error {
	audioSentEvt, ok := evt.(domain.AudioSentEvent)
	if !ok {
		return errors.New("unexpected event")
	}
	log.Println("Going to handle event", audioSentEvt.ID(), "aggregate id", audioSentEvt.AggregateID(), "file", audioSentEvt.Mp3Fullname())

	dominantColor := handler.getDominantAvatarColor(audioSentEvt.UserAvatarURL(), audioSentEvt.FileName())
	t := getDuration(audioSentEvt.Mp3Fullname())

	embed := &discordgo.MessageEmbed{
		Title:     audioSentEvt.Username(),
		Timestamp: time.Now().Format(time.RFC3339),
		Color:     dominantColor,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: audioSentEvt.UserAvatarURL(),
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   handler.durationText,
				Value:  formatSeconds(int(t)),
				Inline: false,
			},
		},
	}
	_, err := handler.session.ChannelMessageEditEmbed(audioSentEvt.ChannelId(), audioSentEvt.AggregateID(), embed)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (handler *AddMetadataOnAudioSent) getDominantAvatarColor(url string, fileName string) int {
	err := downloadFile(url, handler.resolveFullPath(fmt.Sprintf("%s.png", fileName)))
	if err != nil {
		log.Println(err)
		return 0
	}
	color, err := handler.prominentColor(fileName)
	if err != nil {
		return 0
	}
	return color

}

func (handler *AddMetadataOnAudioSent) prominentColor(fileName string) (int, error) {
	// Step 1: Load the image
	img, err := loadImage(handler.resolveFullPath(fmt.Sprintf("%s.png", fileName)))
	if err != nil {
		return 0, errors.New(fmt.Sprintf("Failed to load image: %v", err))
	}

	// Step 2: Process it
	colours, err := prominentcolor.Kmeans(img)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("Failed to process image: %v", err))
	}

	for _, colour := range colours {
		value, _ := strconv.ParseInt(colour.AsString(), 16, 64)
		return int(value), nil
	}
	return 0, errors.New("couldnt get any dominant color")
}

// TODO make a file handler repository that knows full path and can open files
func (handler *AddMetadataOnAudioSent) resolveFullPath(fileName string) string {
	baseFilePath := handler.basePath
	return fmt.Sprintf("%s/%s", baseFilePath, fileName)
}
