package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/EdlinOrg/prominentcolor"
	"github.com/bwmarrin/discordgo"
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
	"os/signal"
	"path/filepath"
	"strconv"
	"time"
)

func getBasePath() string {
	var baseFilePath = os.Getenv("BASE_PATH")
	if baseFilePath == "" {
		baseFilePath = "./tmp"
	}
	return baseFilePath

}

func main() {

	Token := os.Getenv("BOT_TOKEN")
	if Token == "" {
		fmt.Println("Please set token on BOT_TOKEN env")
		return
	}
	ChannelName := "TATERU"

	s, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session:", err)
		return
	}
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		fmt.Println("Bot is ready")

		baseFilePath := getBasePath()
		if _, err := os.Stat(baseFilePath); os.IsNotExist(err) {
			err := os.Mkdir(baseFilePath, 0750)
			if err != nil {
				fmt.Println(err)
				return
			}
		}

	})

	var lockedUser string
	done := make(chan bool)
	defer close(done)
	s.AddHandler(func(s *discordgo.Session, r *discordgo.VoiceStateUpdate) {
		user, err := s.User(r.UserID)
		if err != nil {
			return
		}
		if user.Bot {
			return
		}
		if r.ChannelID == "" && lockedUser != r.UserID {
			return
		}
		if lockedUser == r.UserID {
			done <- true
			lockedUser = ""
			return
		}

		channel, err := s.Channel(r.ChannelID)
		if err != nil {
			fmt.Println(err)
			return
		}
		if channel.Name != ChannelName {
			return
		}

		lockedUser = r.UserID
		recordAndSend(s, r.GuildID, r.ChannelID, user, done)
		lockedUser = ""
	})

	// We only really care about receiving voice state updates.
	s.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildVoiceStates)

	err = s.Open()
	if err != nil {
		log.Fatalf("Cannot open the session: %v", err)
	}
	defer func(s *discordgo.Session) {
		err := s.Close()
		if err != nil {
			fmt.Println("err closing session", err)
		}
	}(s)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
	log.Println("Graceful shutdown")

}

func prominentColor(fileName string) int {
	// Step 1: Load the image
	img, err := loadImage(resolveFullPath(fmt.Sprintf("%s.png", fileName)))
	if err != nil {
		log.Fatal("Failed to load image", err)
	}

	// Step 2: Process it
	colours, err := prominentcolor.Kmeans(img)
	if err != nil {
		log.Fatal("Failed to process image", err)
	}

	for _, colour := range colours {
		value, _ := strconv.ParseInt(colour.AsString(), 16, 64)
		return int(value)
	}
	return 0
}

func loadImage(fileInput string) (image.Image, error) {
	f, err := os.Open(filepath.Clean(fileInput))
	if err != nil {
		return nil, err
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Println("err closing image file", err)
		}
	}(f)
	img, _, err := image.Decode(f)
	return img, err
}

func recordAndSend(s *discordgo.Session, guildId string, channelId string, user *discordgo.User, done chan bool) {
	v, err := s.ChannelVoiceJoin(guildId, channelId, true, false)

	if err != nil {
		fmt.Println("failed to join voice channel:", err)
		return
	}

	go func() {
		<-done
		close(v.OpusRecv)
		v.Close()
		err := v.Disconnect()
		if err != nil {
			fmt.Println(err)
		}
	}()

	fileNames := handleVoice(v.OpusRecv, user)
	defer deleteFiles(fileNames)
	sendAudioFiles(s, guildId, fileNames, user)

}

func deleteFiles(fileNames []string) {
	for _, fileName := range fileNames {
		_ = os.Remove(resolveFullPath(fmt.Sprintf("%s.png", fileName)))
		_ = os.Remove(resolveFullPath(fmt.Sprintf("%s.ogg", fileName)))
		_ = os.Remove(resolveFullPath(fmt.Sprintf("%s.mp3", fileName)))
	}

}

func sendAudioFiles(s *discordgo.Session, guildId string, fileNames []string, user *discordgo.User) {
	channels, err := s.GuildChannels(guildId)
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
		sendAudioFile(s, chID, fileName, user)
	}

}

func formatSeconds(inSeconds int) string {
	minutes := inSeconds / 60
	seconds := inSeconds % 60
	str := fmt.Sprintf("%dm:%02ds", minutes, seconds)
	return str
}

func getDominantAvatarColor(url string, fileName string) int {
	err := downloadFile(url, resolveFullPath(fmt.Sprintf("%s.png", fileName)))
	if err != nil {
		log.Fatal(err)
	}
	return prominentColor(fileName)

}

func resolveFullPath(fileName string) string {
	baseFilePath := getBasePath()
	return fmt.Sprintf("%s/%s", baseFilePath, fileName)
}

func sendAudioFile(s *discordgo.Session, chID string, fileName string, user *discordgo.User) {
	mp3FullName := resolveFullPath(fmt.Sprintf("%s", fileName) + ".mp3")
	t := getDuration(mp3FullName)

	file, err := os.Open(filepath.Clean(mp3FullName))
	if err != nil {
		fmt.Println(err)
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
			fmt.Println(err)
		}
	}(file)

	var discFiles []*discordgo.File
	discFiles = append(discFiles, &discFile)

	dominantColor := getDominantAvatarColor(user.AvatarURL(""), fileName)

	embed := &discordgo.MessageEmbed{
		Title:     user.Username,
		Timestamp: time.Now().Format(time.RFC3339),
		Color:     dominantColor,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: user.AvatarURL(""),
		},
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "Duration",
				Value:  formatSeconds(int(t)),
				Inline: false,
			},
		},
	}
	_, err = s.ChannelMessageSendComplex(chID, &discordgo.MessageSend{
		Embed: embed,
		Files: discFiles,
	})
	if err != nil {
		fmt.Println(err)
		return
	}
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
			fmt.Println("error closing http response body", err)
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
			fmt.Println("error closing file", err)
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
			fmt.Println(err)
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
			fmt.Println(err)
			return 0
		}

		t = t + f.Duration().Seconds()
	}

	return t

}

func handleVoice(c chan *discordgo.Packet, user *discordgo.User) []string {
	files := make(map[string]media.Writer)
	for p := range c {
		name := user.Username + "-" + fmt.Sprintf("%d", p.SSRC)
		file, ok := files[name]
		if !ok {
			var err error
			file, err = oggwriter.New(resolveFullPath(fmt.Sprintf("%s.ogg", name)), 48000, 2)
			if err != nil {
				fmt.Printf("failed to create file %d.ogg, giving up on recording: %v\n", p.SSRC, err)
				return nil
			}
			files[name] = file
		}
		// Construct pion RTP packet from DiscordGo's type.
		rtpPacket := createPionRTPPacket(p)
		err := file.WriteRTP(rtpPacket)
		if err != nil {
			fmt.Printf("failed to write to file %d.ogg, giving up on recording: %v\n", p.SSRC, err)
		}
	}

	// Once we made it here, we're done listening for packets. Close all files
	var mp3Names []string
	for fileName, f := range files {
		err := f.Close()
		if err != nil {
			return nil
		}

		err = convertToMp3(resolveFullPath(fmt.Sprintf("%s.ogg", fileName)), resolveFullPath(fmt.Sprintf("%s.mp3", fileName)))
		if err != nil {
			fmt.Println(err)
			return nil
		}
		mp3Names = append(mp3Names, fileName)
	}
	return mp3Names

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
