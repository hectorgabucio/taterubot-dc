package application

import (
	"context"
	"errors"
	"fmt"
	"github.com/hectorgabucio/taterubot-dc/domain"
	"github.com/hectorgabucio/taterubot-dc/kit/event"
	"log"
)

type RemoveFilesWhenNotNeeded struct {
	fsRepo domain.FileRepository
}

func NewRemoveFilesWhenNotNeeded(fsRepo domain.FileRepository) *RemoveFilesWhenNotNeeded {
	return &RemoveFilesWhenNotNeeded{fsRepo: fsRepo}
}

func (handler *RemoveFilesWhenNotNeeded) Handle(ctx context.Context, evt event.Event) error {
	doneProcessingEvt, ok := evt.(domain.DoneProcessingFilesEvent)
	if !ok {
		return errors.New("unexpected event")
	}
	log.Println("hello", doneProcessingEvt)
	handler.fsRepo.DeleteAll(
		fmt.Sprintf("%s.png", doneProcessingEvt.AggregateID()),
		fmt.Sprintf("%s.mp3", doneProcessingEvt.AggregateID()),
		fmt.Sprintf("%s.ogg", doneProcessingEvt.AggregateID()),
	)

	return nil
}
