package domain

import "github.com/hectorgabucio/taterubot-dc/kit/event"

const AudioSentEventType event.Type = "events.audio.sent"

type AudioSentEvent struct {
	id string
	event.BaseEvent
	channelId     string
	username      string
	userAvatarURL string
	mp3Fullname   string
	fileName      string
}

func NewAudioSentEvent(id string, channelId string, username string, userAvatarURL string, mp3Fullname string, fileName string) AudioSentEvent {
	return AudioSentEvent{
		BaseEvent:     event.NewBaseEvent(id),
		channelId:     channelId,
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

func (e AudioSentEvent) ChannelId() string {
	return e.channelId
}

func (e AudioSentEvent) FileName() string {
	return e.fileName
}
