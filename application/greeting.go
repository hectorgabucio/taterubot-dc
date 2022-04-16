package application

import (
	"context"
	"errors"
	"fmt"
	"github.com/hectorgabucio/taterubot-dc/domain/discord"
	"github.com/hectorgabucio/taterubot-dc/kit/command"
	"github.com/hectorgabucio/taterubot-dc/localizations"
)

const GreetingCommandType command.Type = "command.greeting"

type GreetingCommand struct {
	InteractionToken string
}

func NewGreetingCommand(interactionToken string) GreetingCommand {
	return GreetingCommand{InteractionToken: interactionToken}
}

func (c GreetingCommand) Type() command.Type {
	return GreetingCommandType
}

type GreetingCommandHandler struct {
	service *GreetingMessageCreator
}

// NewGreetingCommandHandler initializes a new GreetingCommandHandler.
func NewGreetingCommandHandler(service *GreetingMessageCreator) GreetingCommandHandler {
	return GreetingCommandHandler{
		service: service,
	}
}

// Handle implements the command.Handler interface.
func (h GreetingCommandHandler) Handle(ctx context.Context, cmd command.Command) error {
	greetingCmd, ok :=
		cmd.(GreetingCommand)
	if !ok {
		return errors.New("unexpected command")
	}
	return h.service.send(greetingCmd.InteractionToken)
}

type GreetingMessageCreator struct {
	discordClient discord.Client
	localization  *localizations.Localizer
	channelName   string
}

func NewGreetingMessageCreator(discord discord.Client, localization *localizations.Localizer, channelName string) *GreetingMessageCreator {
	return &GreetingMessageCreator{
		discord,
		localization,
		channelName,
	}
}

func (service *GreetingMessageCreator) send(interactionToken string) error {
	guilds, err := service.discordClient.GetGuilds()
	if err != nil {
		return fmt.Errorf("err getting guilds, %w", err)
	}
	botUsername := service.discordClient.GetBotUsername()
	for _, guild := range guilds {
		channels, err := service.discordClient.GetGuildChannels(guild.ID)
		if err != nil {
			return fmt.Errorf("err getting guild channels, %w", err)
		}

		chosenChannelIDToSendGreeting := ""
		voiceChannelID := ""
		for _, channel := range channels {
			if channel.Type == discord.ChannelTypeGuildText && chosenChannelIDToSendGreeting == "" {
				chosenChannelIDToSendGreeting = channel.ID
			}
			if channel.Type == discord.ChannelTypeGuildVoice && channel.Name == service.channelName {
				voiceChannelID = channel.ID
			}
		}

		if voiceChannelID == "" {
			createdChannel, err := service.discordClient.CreateChannel(guild.ID, service.channelName, discord.ChannelTypeGuildVoice, 2)
			if err == nil {
				voiceChannelID = createdChannel.ID
			}
		}

		voiceChannelReplacement := fmt.Sprintf("<#%s>", voiceChannelID)
		if voiceChannelID == "" {
			voiceChannelReplacement = service.channelName
		}
		greetingMessage := service.localization.Get("texts.hello", &localizations.Replacements{"voiceChannel": voiceChannelReplacement, "botName": botUsername})
		if err := service.discordClient.EditInteraction(interactionToken, greetingMessage); err != nil {
			return fmt.Errorf("err sending interaction response, %w", err)
		}
	}
	return nil
}
