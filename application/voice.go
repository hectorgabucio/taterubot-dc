package application

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/hectorgabucio/taterubot-dc/domain"
	"github.com/hectorgabucio/taterubot-dc/domain/discord"
	"github.com/hectorgabucio/taterubot-dc/kit/command"
	"github.com/hectorgabucio/taterubot-dc/kit/event"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/oggwriter"
	"log"
	"os"
	"os/exec"
)

const RecordingCommandType command.Type = "command.recording"

type RecordingCommand struct {
	UserId           string
	CurrentChannelId string
	GuildId          string
	Username         string
	AvatarUrl        string
}

func NewRecordingCommand(userId string, channelId string, guildId string, username string, avatarUrl string) RecordingCommand {
	return RecordingCommand{
		UserId:           userId,
		CurrentChannelId: channelId,
		GuildId:          guildId,
		Username:         username,
		AvatarUrl:        avatarUrl,
	}
}

func (c RecordingCommand) Type() command.Type {
	return RecordingCommandType
}

type RecordingCommandHandler struct {
	service *VoiceRecorder
}

// NewRecordingCommandHandler initializes a new RecordingCommandHandler.
func NewRecordingCommandHandler(service *VoiceRecorder) RecordingCommandHandler {
	return RecordingCommandHandler{
		service: service,
	}
}

// Handle implements the command.Handler interface.
func (h RecordingCommandHandler) Handle(ctx context.Context, cmd command.Command) error {
	recordingCmd, ok := cmd.(RecordingCommand)
	if !ok {
		return errors.New("unexpected command")
	}
	return h.service.HandleVoiceRecording(recordingCmd.UserId, recordingCmd.CurrentChannelId, recordingCmd.GuildId, recordingCmd.Username, recordingCmd.AvatarUrl)
}

type VoiceRecorder struct {
	lockedUserRepository domain.LockedUserRepository
	eventBus             event.Bus
	discord              discord.Client
	configChannelName    string
	fsRepo               domain.FileRepository
	session              *discordgo.Session
}

func NewVoiceRecorder(discord discord.Client, configChannelName string, lockedUserRepository domain.LockedUserRepository, eventBus event.Bus, fsRepo domain.FileRepository, session *discordgo.Session) *VoiceRecorder {
	return &VoiceRecorder{
		lockedUserRepository: lockedUserRepository,
		eventBus:             eventBus,
		discord:              discord,
		configChannelName:    configChannelName,
		fsRepo:               fsRepo,
		session:              session,
	}
}

func (usecase *VoiceRecorder) HandleVoiceRecording(userId string, nowChannelId string, guildID string, username string, avatarUrl string) error {

	currentLockedUser, done := usecase.lockedUserRepository.GetCurrentLock(guildID)

	if nowChannelId == "" && currentLockedUser != userId {
		return nil
	}
	if currentLockedUser == userId {
		done <- true
		usecase.lockedUserRepository.ReleaseUserLock(guildID)
		return nil
	}

	channel, err := usecase.discord.GetChannel(nowChannelId)
	if err != nil {
		log.Println(err)
		return err
	}
	if channel.Name != usecase.configChannelName {
		return nil
	}

	usecase.lockedUserRepository.SetLock(guildID, userId)
	return usecase.recordAndSend(guildID, nowChannelId, username, avatarUrl, done)
}

func (usecase *VoiceRecorder) recordAndSend(guildId string, channelId string, username string, avatarUrl string, done chan bool) error {
	v, err := usecase.discord.JoinVoiceChannel(guildId, channelId, true, false)

	if err != nil {
		log.Println("failed to join voice channel:", err)
		return err
	}

	go func() {
		<-done
		log.Println("done recording")
		err := usecase.discord.EndVoiceConnection(v)
		if err != nil {
			log.Println(err)
		}

	}()
	usecase.handleVoice(v.VoiceReceiver, guildId, username, avatarUrl)
	return nil
}

func (usecase *VoiceRecorder) handleVoice(c chan *discord.Packet, guildId string, username string, avatarUrl string) []string {
	files := make(map[string]media.Writer)
	for p := range c {
		name := username + "-" + fmt.Sprintf("%d", p.SSRC)
		file, ok := files[name]
		if !ok {
			var err error
			file, err = oggwriter.New(usecase.fsRepo.GetFullPath(fmt.Sprintf("%s.ogg", name)), 48000, 2)
			if err != nil {
				log.Printf("failed to create file %d.ogg, giving up on recording: %v\n", p.SSRC, err)
				return nil
			}
			files[name] = file
		}
		// Construct pion RTP packet from discord's type.
		rtpPacket := createPionRTPPacket(p)
		err := file.WriteRTP(rtpPacket)
		if err != nil {
			log.Printf("failed to write to file %d.ogg, giving up on recording: %v\n", p.SSRC, err)
		}
	}

	log.Println("done listening voice")

	// Once we made it here, we're done listening for packets. Close all files
	var mp3Names []string
	for fileName, f := range files {
		err := f.Close()
		if err != nil {
			return nil
		}

		err = convertToMp3(usecase.fsRepo.GetFullPath(fmt.Sprintf("%s.ogg", fileName)), usecase.fsRepo.GetFullPath(fmt.Sprintf("%s.mp3", fileName)))
		if err != nil {
			log.Println(err)
			return nil
		}
		mp3Names = append(mp3Names, fileName)
	}

	// TODO event recording file created
	usecase.sendAudioFiles(guildId, mp3Names, username, avatarUrl)

	return mp3Names

}

func (usecase *VoiceRecorder) sendAudioFiles(guildId string, fileNames []string, username string, avatarUrl string) {
	channels, err := usecase.discord.GetGuildChannels(guildId)
	if err != nil {
		return
	}

	var chID string
	for _, ch := range channels {
		if ch.Type == discord.ChannelTypeGuildText {
			chID = ch.Id
			break

		}
	}

	if chID == "" {
		return
	}

	for _, fileName := range fileNames {
		usecase.sendAudioFile(chID, fileName, username, avatarUrl)
	}

}

func (usecase *VoiceRecorder) sendAudioFile(chID string, fileName string, username string, avatarUrl string) {
	mp3FullName := usecase.fsRepo.GetFullPath(fmt.Sprintf("%s", fileName) + ".mp3")
	file, err := usecase.fsRepo.Open(mp3FullName)
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
	messageSent, err := usecase.session.ChannelMessageSendComplex(chID, &discordgo.MessageSend{
		Files: discFiles,
	})
	if err != nil {
		log.Println(err)
		return
	}

	events := []event.Event{
		domain.NewAudioSentEvent(messageSent.ID, messageSent.ChannelID, username, avatarUrl, mp3FullName, fileName),
	}

	go func() {
		err := usecase.eventBus.Publish(context.Background(), events)
		if err != nil {
			log.Println("err publishing audio sent event", err)
		}
	}()

}

func convertToMp3(input string, output string) error {
	cmd := exec.Command("ffmpeg", "-y", "-i", input, output)

	err := cmd.Run()

	return err
}

func createPionRTPPacket(p *discord.Packet) *rtp.Packet {
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
