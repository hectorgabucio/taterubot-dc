package application

// TODO create image infra service
// TODO create mp3 infra service
import (
	"context"
	"errors"
	"fmt"
	"github.com/EdlinOrg/prominentcolor"
	"github.com/bwmarrin/discordgo"
	"github.com/hectorgabucio/taterubot-dc/domain"
	"github.com/hectorgabucio/taterubot-dc/kit/event"
	"github.com/tcolgate/mp3"
	"image"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type AddMetadataOnAudioSent struct {
	session      *discordgo.Session
	durationText string
	fsRepo       domain.FileRepository
}

func NewAddMetadataOnAudioSent(session *discordgo.Session, durationText string, fsRepo domain.FileRepository) *AddMetadataOnAudioSent {
	return &AddMetadataOnAudioSent{session: session, durationText: durationText, fsRepo: fsRepo}
}

func (handler *AddMetadataOnAudioSent) Handle(_ context.Context, evt event.Event) error {
	audioSentEvt, ok := evt.(domain.AudioSentEvent)
	if !ok {
		return errors.New("unexpected event")
	}
	log.Println("Going to handle event", audioSentEvt.ID(), "aggregate id", audioSentEvt.AggregateID(), "file", audioSentEvt.Mp3Fullname())

	dominantColor := handler.getDominantAvatarColor(audioSentEvt.UserAvatarURL(), audioSentEvt.FileName())
	t := handler.getDuration(audioSentEvt.Mp3Fullname())

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
	err := handler.downloadFile(url, fmt.Sprintf("%s.png", fileName))
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
	img, err := handler.loadImage(fmt.Sprintf("%s.png", fileName))
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

func (handler *AddMetadataOnAudioSent) loadImage(fileInput string) (image.Image, error) {
	f, err := handler.fsRepo.Open(fileInput)
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Println("err closing image file", err)
		}
	}(f)
	img, _, err := image.Decode(f)
	return img, err
}

func (handler *AddMetadataOnAudioSent) downloadFile(URL, fileName string) error {
	//Get the response bytes from the url
	response, err := http.Get(URL)
	if err != nil {
		return err
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
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println("error closing file", err)
		}
	}(file)

	//Write the bytes to the field
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return err
	}

	return nil
}

func (handler *AddMetadataOnAudioSent) getDuration(fileName string) float64 {
	file1, err := handler.fsRepo.Open(fileName)
	if err != nil {
		return 0
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println(err)
		}
	}(file1)

	d := mp3.NewDecoder(file1)
	var f mp3.Frame
	skipped := 0

	var t float64
	for {

		if err := d.Decode(&f, &skipped); err != nil {
			if err == io.EOF {
				break
			}
			log.Println(err)
			return 0
		}

		t = t + f.Duration().Seconds()
	}

	return t

}

func formatSeconds(inSeconds int) string {
	minutes := inSeconds / 60
	seconds := inSeconds % 60
	str := fmt.Sprintf("%dm:%02ds", minutes, seconds)
	return str
}
