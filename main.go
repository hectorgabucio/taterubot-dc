package main

import (
	"bufio"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/oggwriter"
	"github.com/tcolgate/mp3"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
)

func main() {

	Token := os.Getenv("BOT_TOKEN")
	if Token == "" {
		fmt.Println("Please set token on BOT_TOKEN env")
	}
	ChannelName := "TATERUV2-TEST"

	s, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session:", err)
		return
	}
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		fmt.Println("Bot is ready")

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
			fmt.Println("done recording")
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
		fmt.Println(r.ChannelID, r.UserID, user.Username)
		recordAndSend(s, r.GuildID, r.ChannelID, done)
		lockedUser = ""
		fmt.Println(lockedUser)

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

func recordAndSend(s *discordgo.Session, guildId string, channelId string, done chan bool) {
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

	handleVoice(v.OpusRecv)
	sendAudioFile(s, guildId)
}

func sendAudioFile(s *discordgo.Session, guildId string) {
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

	t := getDuration("file.mp3")

	file, err := os.Open("file.mp3")
	if err != nil {
		return
	}

	reader := bufio.NewReader(file)
	discFile := discordgo.File{
		Name:        "file.mp3",
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
	_, err = s.ChannelMessageSendComplex(chID, &discordgo.MessageSend{
		Content:         "uwu, duration of " + fmt.Sprintf("%f", t),
		TTS:             false,
		Files:           discFiles,
		AllowedMentions: nil,
		File:            nil,
		Embed:           nil,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

}

func getDuration(fileName string) float64 {
	file1, err := os.Open(fileName)
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

func handleVoice(c chan *discordgo.Packet) {
	files := make(map[uint32]media.Writer)
	for p := range c {
		file, ok := files[p.SSRC]
		if !ok {
			var err error
			file, err = oggwriter.New(fmt.Sprintf("%d.ogg", 1), 48000, 2)
			if err != nil {
				fmt.Printf("failed to create file %d.ogg, giving up on recording: %v\n", p.SSRC, err)
				return
			}
			files[p.SSRC] = file
		}
		// Construct pion RTP packet from DiscordGo's type.
		rtp := createPionRTPPacket(p)
		err := file.WriteRTP(rtp)
		if err != nil {
			fmt.Printf("failed to write to file %d.ogg, giving up on recording: %v\n", p.SSRC, err)
		}
	}

	// Once we made it here, we're done listening for packets. Close all files
	for _, f := range files {
		err := f.Close()
		if err != nil {
			return
		}

		err = convertToMp3("1.ogg", "file.mp3")
		if err != nil {
			fmt.Println(err)
			return
		}

	}

}

func convertToMp3(input string, output string) error {
	cmd := exec.Command("ffmpeg", "-y", "-i", input, output)

	fmt.Println(cmd.String())

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
