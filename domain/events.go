package domain

import "github.com/hectorgabucio/taterubot-dc/kit/event"

const AudioSentEventType event.Type = "events.audio.sent"

type AudioSentEvent struct {
	event.BaseEvent
	ChannelID     string
	Username      string
	UserAvatarURL string
	Mp3Fullname   string
	FileName      string
}

func NewAudioSentEvent(id string, channelID string, username string, userAvatarURL string, mp3Fullname string, fileName string) AudioSentEvent {
	return AudioSentEvent{
		BaseEvent:     event.NewBaseEvent(id),
		ChannelID:     channelID,
		Username:      username,
		UserAvatarURL: userAvatarURL,
		Mp3Fullname:   mp3Fullname,
		FileName:      fileName,
	}
}

func (e AudioSentEvent) Type() event.Type {
	return AudioSentEventType
}

const DoneProcessingFilesEventType event.Type = "events.files.processed"

type DoneProcessingFilesEvent struct {
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
