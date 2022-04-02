package application

import (
	"context"
	"errors"
	"fmt"
	"github.com/hectorgabucio/taterubot-dc/domain/discord"
	"github.com/hectorgabucio/taterubot-dc/kit/command"
	"github.com/hectorgabucio/taterubot-dc/localizations"
	"log"
)

const GreetingCommandType command.Type = "command.greeting"

type GreetingCommand struct {
}

func NewGreetingCommand() GreetingCommand {
	return GreetingCommand{}
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
	_, ok := cmd.(GreetingCommand)
	if !ok {
		return errors.New("unexpected command")
	}
	return h.service.send()
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

func (service *GreetingMessageCreator) send() error {
	guilds, err := service.discordClient.GetGuilds()
	if err != nil {
		log.Println(err)
		return err
	}
	botUsername := service.discordClient.GetBotUsername()
	for _, guild := range guilds {
		channels, err := service.discordClient.GetGuildChannels(guild.Id)
		if err != nil {
			log.Println(err)
			return err
		}

		chosenChannelIdToSendGreeting := ""
		voiceChannelId := ""
		for _, channel := range channels {
			if channel.Type == discord.ChannelTypeGuildText && chosenChannelIdToSendGreeting == "" {
				chosenChannelIdToSendGreeting = channel.Id
			}
			if channel.Type == discord.ChannelTypeGuildVoice && channel.Name == service.channelName {
				voiceChannelId = channel.Id
			}
		}

		// if no voice channel found, try to create it if possible
		if voiceChannelId == "" {
			createdChannel, err := service.discordClient.CreateChannel(guild.Id, service.channelName, discord.ChannelTypeGuildVoice, 2)
			if err == nil {
				voiceChannelId = createdChannel.Id
			}
		}

		voiceChannelReplacement := fmt.Sprintf("<#%s>", voiceChannelId)
		if voiceChannelId == "" {
			voiceChannelReplacement = service.channelName
		}
		greetingMessage := service.localization.Get("texts.hello", &localizations.Replacements{"voiceChannel": voiceChannelReplacement, "botName": botUsername})
		err = service.discordClient.SendTextMessage(chosenChannelIdToSendGreeting, greetingMessage)
		if err != nil {
			log.Println(err)
			return err
		}
		break
	}
	return nil
}
