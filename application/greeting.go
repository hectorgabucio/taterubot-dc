package application

import (
	"context"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
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
	cmd, ok := cmd.(GreetingCommand)
	if !ok {
		return errors.New("unexpected command")
	}
	return h.service.Send()
}

type GreetingMessageCreator struct {
	session      *discordgo.Session
	localization *localizations.Localizer
	channelName  string
}

func NewGreetingMessageCreator(session *discordgo.Session, localization *localizations.Localizer, channelName string) *GreetingMessageCreator {
	return &GreetingMessageCreator{
		session,
		localization,
		channelName,
	}
}

func (service *GreetingMessageCreator) Send() error {
	guilds, err := service.session.UserGuilds(100, "", "")
	if err != nil {
		log.Println(err)
		return err
	}
	botUsername := service.session.State.User.Username
	for _, guild := range guilds {
		channels, err := service.session.GuildChannels(guild.ID)
		if err != nil {
			log.Println(err)
			return err
		}

		chosenChannelIdToSendGreeting := ""
		voiceChannelId := ""
		for _, channel := range channels {
			if channel.Type == discordgo.ChannelTypeGuildText && chosenChannelIdToSendGreeting == "" {
				chosenChannelIdToSendGreeting = channel.ID
			}
			if channel.Type == discordgo.ChannelTypeGuildVoice && channel.Name == service.channelName {
				voiceChannelId = channel.ID
			}
		}

		// if no voice channel found, try to create it if possible
		if voiceChannelId == "" {
			createdChannel, err := service.session.GuildChannelCreateComplex(guild.ID, discordgo.GuildChannelCreateData{
				Name:      service.channelName,
				Type:      discordgo.ChannelTypeGuildVoice,
				UserLimit: 2,
			})
			if err == nil {
				voiceChannelId = createdChannel.ID
			}
		}

		voiceChannelReplacement := fmt.Sprintf("<#%s>", voiceChannelId)
		if voiceChannelId == "" {
			voiceChannelReplacement = service.channelName
		}
		greetingMessage := service.localization.Get("texts.hello", &localizations.Replacements{"voiceChannel": voiceChannelReplacement, "botName": botUsername})
		_, err = service.session.ChannelMessageSend(chosenChannelIdToSendGreeting, greetingMessage)
		if err != nil {
			log.Println(err)
			return err
		}
		break

	}
	return nil
}
