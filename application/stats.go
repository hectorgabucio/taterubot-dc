package application

import (
	"context"
	"errors"
	"fmt"
	"github.com/hectorgabucio/taterubot-dc/domain"
	"github.com/hectorgabucio/taterubot-dc/domain/discord"
	"github.com/hectorgabucio/taterubot-dc/kit/command"
	"github.com/hectorgabucio/taterubot-dc/localizations"
	"log"
	"time"
)

const StatsCommandType command.Type = "command.stats"

type StatsCommand struct {
	InteractionToken string
	GuildID          string
}

func NewStatsCommand(interactionToken string, guildID string) StatsCommand {
	return StatsCommand{InteractionToken: interactionToken, GuildID: guildID}
}

func (c StatsCommand) Type() command.Type {
	return StatsCommandType
}

type StatsCommandHandler struct {
	service *StatsMessageCreator
}

// NewStatsCommandHandler initializes a new StatsCommandHandler.
func NewStatsCommandHandler(service *StatsMessageCreator) StatsCommandHandler {
	return StatsCommandHandler{
		service: service,
	}
}

// Handle implements the command.Handler interface.
func (h StatsCommandHandler) Handle(ctx context.Context, cmd command.Command) error {
	statsCmd, ok :=
		cmd.(StatsCommand)
	if !ok {
		return errors.New("unexpected command")
	}
	return h.service.send(statsCmd.InteractionToken, statsCmd.GuildID)
}

type StatsMessageCreator struct {
	discordClient discord.Client
	localization  *localizations.Localizer
	voiceDataRepo domain.VoiceDataRepository
}

func NewStatsMessageCreator(discord discord.Client, localization *localizations.Localizer, voiceDataRepo domain.VoiceDataRepository) *StatsMessageCreator {
	return &StatsMessageCreator{
		discord,
		localization,
		voiceDataRepo,
	}
}

func (service *StatsMessageCreator) send(interactionToken string, guildID string) error {
	log.Println("received stats command", interactionToken, guildID)

	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()

	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)

	onRange, err := service.voiceDataRepo.GetOnRange(guildID, firstOfMonth, lastOfMonth)
	if err != nil {
		return fmt.Errorf("err getting voice range, %w", err)
	}
	if len(onRange) == 0 {
		return service.sendEmptyInteraction(interactionToken)
	}

	globalDuration := 0
	for _, voiceData := range onRange {
		globalDuration += voiceData.Duration
	}
	medianDuration := globalDuration / len(onRange)
	message := service.localization.Get("texts.stats",
		&localizations.Replacements{"globalDuration": globalDuration, "globalAmount": len(onRange), "globalMedianDuration": medianDuration})
	if err := service.discordClient.EditInteraction(interactionToken, message); err != nil {
		log.Println(err)
		return fmt.Errorf("err sending interaction response, %w", err)
	}
	return nil
}

func (service *StatsMessageCreator) sendEmptyInteraction(interactionToken string) error {
	if err := service.discordClient.EditInteraction(interactionToken, "Start sending voice messages to have stats!"); err != nil {
		return fmt.Errorf("err sending interaction response, %w", err)
	}
	return nil
}
