package application

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/EdlinOrg/prominentcolor"
	"github.com/bwmarrin/discordgo"
	"github.com/hectorgabucio/taterubot-dc/domain"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/oggwriter"
	"github.com/tcolgate/mp3"
	"image"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

type VoiceRecorder struct {
	lockedUserRepository domain.LockedUserRepository
	session              *discordgo.Session
	configChannelName    string
	basePath             string
	durationText         string
}

func NewVoiceRecorder(session *discordgo.Session, configChannelName string, lockedUserRepository domain.LockedUserRepository, basePath string, durationText string) *VoiceRecorder {
	return &VoiceRecorder{
		lockedUserRepository: lockedUserRepository,
		session:              session,
		configChannelName:    configChannelName,
		basePath:             basePath,
		durationText:         durationText,
	}
}

func (usecase *VoiceRecorder) HandleVoiceRecording(userId string, channelId string, guildID string, user *discordgo.User, done chan bool) {
	currentLockedUser := usecase.lockedUserRepository.GetCurrentLock()

	if channelId == "" && currentLockedUser != userId {
		return
	}
	if currentLockedUser == userId {
		done <- true
		usecase.lockedUserRepository.ReleaseUserLock()
		return
	}

	channel, err := usecase.session.Channel(channelId)
	if err != nil {
		log.Println(err)
		return
	}
	if channel.Name != usecase.configChannelName {
		return
	}

	usecase.lockedUserRepository.SetLock(userId)
	usecase.recordAndSend(guildID, channelId, user, done)
	usecase.lockedUserRepository.ReleaseUserLock()
}

func (usecase *VoiceRecorder) recordAndSend(guildId string, channelId string, user *discordgo.User, done chan bool) {
	v, err := usecase.session.ChannelVoiceJoin(guildId, channelId, true, false)

	if err != nil {
		log.Println("failed to join voice channel:", err)
		return
	}

	go func() {
		<-done
		close(v.OpusRecv)
		v.Close()
		err := v.Disconnect()
		if err != nil {
			log.Println(err)
		}
	}()

	fileNames := usecase.handleVoice(v.OpusRecv, user)
	defer usecase.deleteFiles(fileNames)
	usecase.sendAudioFiles(guildId, fileNames, user)

}

func (usecase *VoiceRecorder) handleVoice(c chan *discordgo.Packet, user *discordgo.User) []string {
	files := make(map[string]media.Writer)
	for p := range c {
		name := user.Username + "-" + fmt.Sprintf("%d", p.SSRC)
		file, ok := files[name]
		if !ok {
			var err error
			file, err = oggwriter.New(usecase.resolveFullPath(fmt.Sprintf("%s.ogg", name)), 48000, 2)
			if err != nil {
				log.Printf("failed to create file %d.ogg, giving up on recording: %v\n", p.SSRC, err)
				return nil
			}
			files[name] = file
		}
		// Construct pion RTP packet from DiscordGo's type.
		rtpPacket := createPionRTPPacket(p)
		err := file.WriteRTP(rtpPacket)
		if err != nil {
			log.Printf("failed to write to file %d.ogg, giving up on recording: %v\n", p.SSRC, err)
		}
	}

	// Once we made it here, we're done listening for packets. Close all files
	var mp3Names []string
	for fileName, f := range files {
		err := f.Close()
		if err != nil {
			return nil
		}

		err = convertToMp3(usecase.resolveFullPath(fmt.Sprintf("%s.ogg", fileName)), usecase.resolveFullPath(fmt.Sprintf("%s.mp3", fileName)))
		if err != nil {
			log.Println(err)
			return nil
		}
		mp3Names = append(mp3Names, fileName)
	}
	return mp3Names

}

func (usecase *VoiceRecorder) sendAudioFiles(guildId string, fileNames []string, user *discordgo.User) {
	channels, err := usecase.session.GuildChannels(guildId)
	if err != nil {
		return
	}

	var chID string
	for _, ch := range channels {
		if ch.Type == discordgo.ChannelTypeGuildText {
			chID = ch.ID
			break

		}
	}

	if chID == "" {
		return
	}

	for _, fileName := range fileNames {
		usecase.sendAudioFile(chID, fileName, user)
	}

}

func (usecase *VoiceRecorder) sendAudioFile(chID string, fileName string, user *discordgo.User) {
	mp3FullName := usecase.resolveFullPath(fmt.Sprintf("%s", fileName) + ".mp3")
	t := getDuration(mp3FullName)

	file, err := os.Open(filepath.Clean(mp3FullName))
	if err != nil {
		log.Println(err)
		return
	}

	reader := bufio.NewReader(file)
	discFile := discordgo.File{
		Name:        mp3FullName,
		ContentType: "audio/mpeg",
		Reader:      reader,
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println(err)
		}
	}(file)

	var discFiles []*discordgo.File
	discFiles = append(discFiles, &discFile)

	dominantColor := usecase.getDominantAvatarColor(user.AvatarURL(""), fileName)

	embed := &discordgo.MessageEmbed{
		Title:     user.Username,
		Timestamp: time.Now().Format(time.RFC3339),
		Color:     dominantColor,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: user.AvatarURL(""),
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   usecase.durationText,
				Value:  formatSeconds(int(t)),
				Inline: false,
			},
		},
	}
	_, err = usecase.session.ChannelMessageSendComplex(chID, &discordgo.MessageSend{
		Embed: embed,
		Files: discFiles,
	})
	if err != nil {
		log.Println(err)
		return
	}
}

func (usecase *VoiceRecorder) getDominantAvatarColor(url string, fileName string) int {
	err := downloadFile(url, usecase.resolveFullPath(fmt.Sprintf("%s.png", fileName)))
	if err != nil {
		log.Println(err)
		return 0
	}
	color, err := usecase.prominentColor(fileName)
	if err != nil {
		return 0
	}
	return color

}

func (usecase *VoiceRecorder) prominentColor(fileName string) (int, error) {
	// Step 1: Load the image
	img, err := loadImage(usecase.resolveFullPath(fmt.Sprintf("%s.png", fileName)))
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

func (usecase *VoiceRecorder) resolveFullPath(fileName string) string {
	baseFilePath := usecase.basePath
	return fmt.Sprintf("%s/%s", baseFilePath, fileName)
}

func convertToMp3(input string, output string) error {
	cmd := exec.Command("ffmpeg", "-y", "-i", input, output)

	err := cmd.Run()

	return err
}

func createPionRTPPacket(p *discordgo.Packet) *rtp.Packet {
	return &rtp.Packet{
		Header: rtp.Header{
			Version: 2,
			// Taken from Discord voice docs
			PayloadType:    0x78,
			SequenceNumber: p.Sequence,
			Timestamp:      p.Timestamp,
			SSRC:           p.SSRC,
		},
		Payload: p.Opus,
	}
}

func formatSeconds(inSeconds int) string {
	minutes := inSeconds / 60
	seconds := inSeconds % 60
	str := fmt.Sprintf("%dm:%02ds", minutes, seconds)
	return str
}

func loadImage(fileInput string) (image.Image, error) {
	f, err := os.Open(filepath.Clean(fileInput))
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

func downloadFile(URL, fileName string) error {
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
		return errors.New("Received non 200 response code")
	}
	//Create a empty file
	file, err := os.Create(filepath.Clean(fileName))
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

func getDuration(fileName string) float64 {
	file1, err := os.Open(filepath.Clean(fileName))
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

func (usecase *VoiceRecorder) deleteFiles(fileNames []string) {
	for _, fileName := range fileNames {
		_ = os.Remove(usecase.resolveFullPath(fmt.Sprintf("%s.png", fileName)))
		_ = os.Remove(usecase.resolveFullPath(fmt.Sprintf("%s.ogg", fileName)))
		_ = os.Remove(usecase.resolveFullPath(fmt.Sprintf("%s.mp3", fileName)))
	}

}
