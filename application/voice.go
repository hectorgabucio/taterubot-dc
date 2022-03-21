package application

import (
	"bufio"
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/hectorgabucio/taterubot-dc/domain"
	"github.com/hectorgabucio/taterubot-dc/kit/event"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3/pkg/media"
	"github.com/pion/webrtc/v3/pkg/media/oggwriter"
	"log"
	"os"
	"os/exec"
)

type VoiceRecorder struct {
	lockedUserRepository domain.LockedUserRepository
	eventBus             event.Bus
	session              *discordgo.Session
	configChannelName    string
	fsRepo               domain.FileRepository
}

func NewVoiceRecorder(session *discordgo.Session, configChannelName string, lockedUserRepository domain.LockedUserRepository, eventBus event.Bus, fsRepo domain.FileRepository) *VoiceRecorder {
	return &VoiceRecorder{
		lockedUserRepository: lockedUserRepository,
		eventBus:             eventBus,
		session:              session,
		configChannelName:    configChannelName,
		fsRepo:               fsRepo,
	}
}

func (usecase *VoiceRecorder) HandleVoiceRecording(userId string, nowChannelId string, guildID string, user *discordgo.User, done chan bool) {

	currentLockedUser := usecase.lockedUserRepository.GetCurrentLock(guildID)

	if nowChannelId == "" && currentLockedUser != userId {
		return
	}
	if currentLockedUser == userId {
		done <- true
		usecase.lockedUserRepository.ReleaseUserLock(guildID)
		return
	}

	channel, err := usecase.session.Channel(nowChannelId)
	if err != nil {
		log.Println(err)
		return
	}
	if channel.Name != usecase.configChannelName {
		return
	}

	usecase.lockedUserRepository.SetLock(guildID, userId)
	usecase.recordAndSend(guildID, nowChannelId, user, done)
}

func (usecase *VoiceRecorder) recordAndSend(guildId string, channelId string, user *discordgo.User, done chan bool) {
	v, err := usecase.session.ChannelVoiceJoin(guildId, channelId, true, false)

	if err != nil {
		log.Println("failed to join voice channel:", err)
		return
	}

	go func() {
		<-done
		log.Println("done recording")
		close(v.OpusRecv)
		v.Close()
		err := v.Disconnect()
		if err != nil {
			log.Println(err)
		}

	}()
	usecase.handleVoice(v.OpusRecv, user, guildId)
}

func (usecase *VoiceRecorder) handleVoice(c chan *discordgo.Packet, user *discordgo.User, guildId string) []string {
	files := make(map[string]media.Writer)
	for p := range c {
		name := user.Username + "-" + fmt.Sprintf("%d", p.SSRC)
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
		// Construct pion RTP packet from DiscordGo's type.
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
	usecase.sendAudioFiles(guildId, mp3Names, user)

	// TODO event finished with processing files
	defer usecase.deleteFiles(mp3Names)
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
		domain.NewAudioSentEvent(messageSent.ID, messageSent.ChannelID, user.Username, user.AvatarURL(""), mp3FullName, fileName),
	}

	err = usecase.eventBus.Publish(context.Background(), events)
	if err != nil {
		log.Println(err)
		return
	}

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
func (usecase *VoiceRecorder) deleteFiles(fileNames []string) {
	usecase.fsRepo.DeleteAll(fileNames...)

}
