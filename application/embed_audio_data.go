package application

import (
	"context"
	"errors"
	"github.com/hectorgabucio/taterubot-dc/domain"
	"github.com/hectorgabucio/taterubot-dc/kit/event"
	"log"
)

type AddMetadataOnAudioSent struct {
}

func NewAddMetadataOnAudioSent() *AddMetadataOnAudioSent {
	return &AddMetadataOnAudioSent{}
}

func (e *AddMetadataOnAudioSent) Handle(_ context.Context, evt event.Event) error {
	audioSentEvt, ok := evt.(domain.AudioSentEvent)
	if !ok {
		return errors.New("unexpected event")
	}
	log.Println("Going to handle event", audioSentEvt.ID())
	return nil
}
