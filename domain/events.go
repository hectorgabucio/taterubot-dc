package domain

import "github.com/hectorgabucio/taterubot-dc/kit/event"

const AudioSentEventType event.Type = "events.audio.sent"

type AudioSentEvent struct {
	id string
	event.BaseEvent
}

func NewAudioSentEvent(id string) AudioSentEvent {
	return AudioSentEvent{
		BaseEvent: event.NewBaseEvent(id),
	}
}

func (e AudioSentEvent) Type() event.Type {
	return AudioSentEventType
}
