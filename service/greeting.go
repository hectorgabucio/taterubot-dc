package service

import (
	"github.com/bwmarrin/discordgo"
	"github.com/hectorgabucio/taterubot-dc/localizations"
	"log"
)

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

func (service *GreetingMessageCreator) Send() {
	guilds, err := service.session.UserGuilds(100, "", "")
	if err != nil {
		log.Println(err)
		return
	}
	botUsername := service.session.State.User.Username
	greetingMessage := service.localization.Get("texts.hello", &localizations.Replacements{"voiceChannel": service.channelName, "botName": botUsername})
	for _, guild := range guilds {
		channels, err := service.session.GuildChannels(guild.ID)
		if err != nil {
			log.Println(err)
			return
		}

		for _, channel := range channels {
			if channel.Type == discordgo.ChannelTypeGuildText {
				_, _ = service.session.ChannelMessageSend(channel.ID, greetingMessage)
				break
			}
		}

	}
}
