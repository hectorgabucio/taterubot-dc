package application

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/hectorgabucio/taterubot-dc/domain"
	"github.com/hectorgabucio/taterubot-dc/domain/discord"
	"github.com/hectorgabucio/taterubot-dc/domain/ogg"
	"github.com/hectorgabucio/taterubot-dc/kit/command"
	"github.com/hectorgabucio/taterubot-dc/kit/event"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

const RecordingCommandType command.Type = "command.recording"

type RecordingCommand struct {
	UserID           string
	CurrentChannelID string
	GuildID          string
	Username         string
	AvatarURL        string
}

func NewRecordingCommand(userID string, channelID string, guildID string, username string, avatarURL string) RecordingCommand {
	return RecordingCommand{
		UserID:           userID,
		CurrentChannelID: channelID,
		GuildID:          guildID,
		Username:         username,
		AvatarURL:        avatarURL,
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
	return h.service.handleVoiceRecording(recordingCmd.UserID, recordingCmd.CurrentChannelID, recordingCmd.GuildID, recordingCmd.Username, recordingCmd.AvatarURL)
}

type VoiceRecorder struct {
	lockedUserRepository domain.LockedUserRepository
	eventBus             event.Bus
	discord              discord.Client
	configChannelName    string
	fsRepo               domain.FileRepository
	oggWriter            ogg.Writer
}

func NewVoiceRecorder(discord discord.Client, configChannelName string, lockedUserRepository domain.LockedUserRepository, eventBus event.Bus, fsRepo domain.FileRepository, writer ogg.Writer) *VoiceRecorder {
	return &VoiceRecorder{
		lockedUserRepository: lockedUserRepository,
		eventBus:             eventBus,
		discord:              discord,
		configChannelName:    configChannelName,
		fsRepo:               fsRepo,
		oggWriter:            writer,
	}
}
func (usecase *VoiceRecorder) handleVoiceRecording(userID string, nowChannelID string, guildID string, username string, avatarURL string) error {
	currentLockedUser, done := usecase.lockedUserRepository.GetCurrentLock(guildID)

	if nowChannelID == "" && currentLockedUser != userID {
		return nil
	}
	if currentLockedUser == userID {
		done <- true
		usecase.lockedUserRepository.ReleaseUserLock(guildID)
		return nil
	}

	channel, err := usecase.discord.GetChannel(nowChannelID)
	if err != nil {
		return fmt.Errorf("err getting channel, %w", err)
	}
	if channel.Name != usecase.configChannelName {
		return nil
	}

	usecase.lockedUserRepository.SetLock(guildID, userID)
	return usecase.recordAndSend(userID, guildID, nowChannelID, username, avatarURL, done)
}

func (usecase *VoiceRecorder) recordAndSend(userID string, guildID string, channelID string, username string, avatarURL string, done chan bool) error {
	closeVoiceConn := make(chan bool)
	v, err := usecase.discord.JoinVoiceChannel(guildID, channelID, true, false, done, closeVoiceConn)
	if err != nil {
		return fmt.Errorf("err joining voice channel, %w", err)
	}

	go func() {
		<-closeVoiceConn
		err := usecase.discord.EndVoiceConnection(v)
		if err != nil {
			log.Println(err)
		}
	}()
	usecase.handleVoice(v.VoiceReceiver, userID, guildID, username, avatarURL)
	return nil
}

func (usecase *VoiceRecorder) handleVoice(c chan *discord.Packet, userID string, guildID string, username string, avatarURL string) []string {
	files := make(map[string]io.Closer)
	for p := range c {
		name := username + "-" + fmt.Sprintf("%d", p.SSRC)
		file, ok := files[name]
		if !ok {
			var err error
			file, err = usecase.oggWriter.NewWriter(usecase.fsRepo.GetFullPath(fmt.Sprintf("%s.ogg", name)))
			if err != nil {
				log.Printf("failed to create file %d.ogg, giving up on recording: %v\n", p.SSRC, err)
				return nil
			}
			files[name] = file
		}
		err := usecase.oggWriter.WriteVoice(file, p)
		if err != nil {
			log.Printf("failed to write to file %d.ogg, giving up on recording: %v\n", p.SSRC, err)
		}
	}

	mp3Names := make([]string, len(files))
	i := 0
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
		mp3Names[i] = fileName
		i++
	}

	usecase.sendAudioFiles(guildID, userID, mp3Names, username, avatarURL)

	return mp3Names
}

func (usecase *VoiceRecorder) sendAudioFiles(guildID string, userID string, fileNames []string, username string, avatarURL string) {
	channels, err := usecase.discord.GetGuildChannels(guildID)
	if err != nil {
		return
	}

	var chID string
	for _, ch := range channels {
		if ch.Type == discord.ChannelTypeGuildText {
			chID = ch.ID

			break
		}
	}

	if chID == "" {
		return
	}

	for _, fileName := range fileNames {
		usecase.sendAudioFile(guildID, userID, chID, fileName, username, avatarURL)
	}
}

func (usecase *VoiceRecorder) sendAudioFile(guildID string, userID string, chID string, fileName string, username string, avatarURL string) {
	mp3FullName := usecase.fsRepo.GetFullPath(fileName + ".mp3")
	file, err := usecase.fsRepo.Open(mp3FullName)
	if err != nil {
		log.Println(err)
		return
	}

	reader := bufio.NewReader(file)

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Println(err)
		}
	}(file)

	messageSent, err := usecase.discord.SendFileMessage(chID, mp3FullName, "audio/mpeg", reader)
	if err != nil {
		log.Println(err)
		return
	}

	events := []event.Event{
		domain.NewAudioSentEvent(messageSent.ID, userID, guildID, messageSent.ChannelID, username, avatarURL, mp3FullName, fileName, messageSent.AttachmentId),
	}
	go func() {
		err := usecase.eventBus.Publish(context.Background(), events)
		if err != nil {
			log.Println("err publishing audio sent event", err)
		}
	}()
}
func convertToMp3(input string, output string) error {
	if err := ffmpeg.Input(input).
		Output(output, ffmpeg.KwArgs{"acodec": "libmp3lame", "b:a": "96k", "map": "a"}).
		OverWriteOutput().Run(); err != nil {
		return fmt.Errorf("failed to convert to mp3, %w", err)
	}
	return nil
}
