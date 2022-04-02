package domain

import "github.com/hectorgabucio/taterubot-dc/kit/event"

const AudioSentEventType event.Type = "events.audio.sent"

type AudioSentEvent struct {
	id string
	event.BaseEvent
	channelID     string
	username      string
	userAvatarURL string
	mp3Fullname   string
	fileName      string
}

func NewAudioSentEvent(id string, channelID string, username string, userAvatarURL string, mp3Fullname string, fileName string) AudioSentEvent {
	return AudioSentEvent{
		BaseEvent:     event.NewBaseEvent(id),
		channelID:     channelID,
		username:      username,
		userAvatarURL: userAvatarURL,
		mp3Fullname:   mp3Fullname,
		fileName:      fileName,
	}
}

func (e AudioSentEvent) Type() event.Type {
	return AudioSentEventType
}

func (e AudioSentEvent) Username() string {
	return e.username
}

func (e AudioSentEvent) UserAvatarURL() string {
	return e.userAvatarURL
}

func (e AudioSentEvent) Mp3Fullname() string {
	return e.mp3Fullname
}

func (e AudioSentEvent) ChannelID() string {
	return e.channelID
}

func (e AudioSentEvent) FileName() string {
	return e.fileName
}

const DoneProcessingFilesEventType event.Type = "events.files.processed"

type DoneProcessingFilesEvent struct {
	id string
	event.BaseEvent
}

func NewDoneProcessingFilesEvent(id string) DoneProcessingFilesEvent {
	return DoneProcessingFilesEvent{
		BaseEvent: event.NewBaseEvent(id),
	}
}

func (e DoneProcessingFilesEvent) Type() event.Type {
	return DoneProcessingFilesEventType
}
