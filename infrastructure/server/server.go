package infrastructure

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/EdlinOrg/prominentcolor"
	"github.com/bwmarrin/discordgo"
	"github.com/hectorgabucio/taterubot-dc/config"
	"github.com/hectorgabucio/taterubot-dc/localizations"
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

type Server struct {
	config       config.Config
	localization *localizations.Localizer
	session      *discordgo.Session
}

func NewServer(ctx context.Context, l *localizations.Localizer, cfg config.Config) (context.Context, Server) {
	log.Println("Bot server running")

	s, err := discordgo.New("Bot " + cfg.BotToken)
	if err != nil {
		log.Fatal("Error initializing bot: " + err.Error())
	}
	srv := Server{cfg, l, s}

	return serverContext(ctx), srv
}

func setupHandlers() {
	session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is ready")

		baseFilePath := getBasePath()
		if _, err := os.Stat(baseFilePath); os.IsNotExist(err) {
			err := os.Mkdir(baseFilePath, 0750)
			if err != nil {
				log.Println(err)
				return
			}
		}

		guilds, err := s.UserGuilds(100, "", "")
		if err != nil {
			log.Println(err)
			return
		}
		for _, guild := range guilds {
			channels, err := s.GuildChannels(guild.ID)
			if err != nil {
				log.Println(err)
				return
			}

			for _, channel := range channels {
				if channel.Type == discordgo.ChannelTypeGuildText {

					_, _ = s.ChannelMessageSend(channel.ID, server.localization.Get("texts.hello", &localizations.Replacements{"voiceChannel": server.config.ChannelName, "botName": r.User.Username}))
					break
				}
			}

		}

	})
}

func (server *Server) Run(ctx context.Context) error {

	server.session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is ready")

		baseFilePath := getBasePath()
		if _, err := os.Stat(baseFilePath); os.IsNotExist(err) {
			err := os.Mkdir(baseFilePath, 0750)
			if err != nil {
				log.Println(err)
				return
			}
		}

		guilds, err := s.UserGuilds(100, "", "")
		if err != nil {
			log.Println(err)
			return
		}
		for _, guild := range guilds {
			channels, err := s.GuildChannels(guild.ID)
			if err != nil {
				log.Println(err)
				return
			}

			for _, channel := range channels {
				if channel.Type == discordgo.ChannelTypeGuildText {

					_, _ = s.ChannelMessageSend(channel.ID, server.localization.Get("texts.hello", &localizations.Replacements{"voiceChannel": server.config.ChannelName, "botName": r.User.Username}))
					break
				}
			}

		}

	})

	var lockedUser string
	done := make(chan bool)
	defer close(done)
	server.session.AddHandler(func(s *discordgo.Session, r *discordgo.VoiceStateUpdate) {
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
			log.Println(err)
			return
		}
		if channel.Name != server.config.ChannelName {
			return
		}

		lockedUser = r.UserID
		server.recordAndSend(r.GuildID, r.ChannelID, user, done)
		lockedUser = ""
	})

	// We only really care about receiving voice state updates.
	server.session.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildVoiceStates)

	err := server.session.Open()
	if err != nil {
		return errors.New(fmt.Sprintf("Cannot open the session: %v", err))
	}
	defer func(s *discordgo.Session) {
		err := s.Close()
		if err != nil {
			log.Println("err closing session", err)
		}
	}(server.session)

	<-ctx.Done()
	return ctx.Err()

}

func serverContext(ctx context.Context) context.Context {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		<-c
		cancel()
	}()

	return ctx
}

func prominentColor(fileName string) (int, error) {
	// Step 1: Load the image
	img, err := loadImage(resolveFullPath(fmt.Sprintf("%s.png", fileName)))
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

func (server *Server) recordAndSend(guildId string, channelId string, user *discordgo.User, done chan bool) {
	v, err := server.session.ChannelVoiceJoin(guildId, channelId, true, false)

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

	fileNames := handleVoice(v.OpusRecv, user)
	defer deleteFiles(fileNames)
	server.sendAudioFiles(guildId, fileNames, user)

}

func deleteFiles(fileNames []string) {
	for _, fileName := range fileNames {
		_ = os.Remove(resolveFullPath(fmt.Sprintf("%s.png", fileName)))
		_ = os.Remove(resolveFullPath(fmt.Sprintf("%s.ogg", fileName)))
		_ = os.Remove(resolveFullPath(fmt.Sprintf("%s.mp3", fileName)))
	}

}

func (server *Server) sendAudioFiles(guildId string, fileNames []string, user *discordgo.User) {
	channels, err := server.session.GuildChannels(guildId)
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
		server.sendAudioFile(chID, fileName, user)
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
		log.Println(err)
		return 0
	}
	color, err := prominentColor(fileName)
	if err != nil {
		return 0
	}
	return color

}

func resolveFullPath(fileName string) string {
	baseFilePath := getBasePath()
	return fmt.Sprintf("%s/%s", baseFilePath, fileName)
}

func (server *Server) sendAudioFile(chID string, fileName string, user *discordgo.User) {
	mp3FullName := resolveFullPath(fmt.Sprintf("%s", fileName) + ".mp3")
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
				Name:   server.localization.Get("texts.duration"),
				Value:  formatSeconds(int(t)),
				Inline: false,
			},
		},
	}
	_, err = server.session.ChannelMessageSendComplex(chID, &discordgo.MessageSend{
		Embed: embed,
		Files: discFiles,
	})
	if err != nil {
		log.Println(err)
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

func handleVoice(c chan *discordgo.Packet, user *discordgo.User) []string {
	files := make(map[string]media.Writer)
	for p := range c {
		name := user.Username + "-" + fmt.Sprintf("%d", p.SSRC)
		file, ok := files[name]
		if !ok {
			var err error
			file, err = oggwriter.New(resolveFullPath(fmt.Sprintf("%s.ogg", name)), 48000, 2)
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

		err = convertToMp3(resolveFullPath(fmt.Sprintf("%s.ogg", fileName)), resolveFullPath(fmt.Sprintf("%s.mp3", fileName)))
		if err != nil {
			log.Println(err)
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
