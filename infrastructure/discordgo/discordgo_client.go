package discordgo

import (
	"github.com/bwmarrin/discordgo"
	"github.com/hectorgabucio/taterubot-dc/domain/discord"
)

type Client struct {
	session *discordgo.Session
}

func NewClient(session *discordgo.Session) *Client {
	return &Client{session: session}
}

func (c *Client) GetGuilds() ([]discord.Guild, error) {
	infraGuilds, err := c.session.UserGuilds(100, "", "")
	if err != nil {
		return nil, err
	}
	var guilds []discord.Guild

	// TODO create a map function
	for _, infraGuild := range infraGuilds {
		newGuild := discord.Guild{
			Id:   infraGuild.ID,
			Name: infraGuild.Name,
		}
		guilds = append(guilds, newGuild)
	}
	return guilds, nil

}

func (c *Client) GetBotUsername() string {
	return c.session.State.User.Username
}

func (c *Client) GetGuildChannels(guildID string) ([]discord.Channel, error) {
	channels, err := c.session.GuildChannels(guildID)
	if err != nil {
		return nil, err
	}
	var mappedChannels []discord.Channel
	for _, infraChannel := range channels {
		newChannel := discord.Channel{
			Id:   infraChannel.ID,
			Name: infraChannel.Name,
			Type: discord.ChannelType(infraChannel.Type),
		}
		mappedChannels = append(mappedChannels, newChannel)
	}

	return mappedChannels, nil
}
func (c *Client) CreateChannel(guildID string, name string, channelType discord.ChannelType, maxUsers int) (discord.Channel, error) {
	createdChannel, err := c.session.GuildChannelCreateComplex(guildID, discordgo.GuildChannelCreateData{
		Name:      name,
		Type:      discordgo.ChannelType(channelType),
		UserLimit: maxUsers,
	})
	if err != nil {
		return discord.Channel{}, err
	}
	return discord.Channel{
		Id:   createdChannel.ID,
		Name: createdChannel.Name,
		Type: discord.ChannelType(createdChannel.Type),
	}, nil
}
func (c *Client) SendTextMessage(channelId string, message string) error {
	_, err := c.session.ChannelMessageSend(channelId, message)
	return err
}
