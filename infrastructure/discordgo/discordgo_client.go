package discordgo

import (
	"github.com/bwmarrin/discordgo"
	"github.com/hectorgabucio/taterubot-dc/domain/discord"
	"io"
	"log"
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
func (c *Client) GetChannel(channelId string) (discord.Channel, error) {
	channel, err := c.session.Channel(channelId)
	if err != nil {
		return discord.Channel{}, err
	}
	return discord.Channel{
		Id:   channel.ID,
		Name: channel.Name,
		Type: discord.ChannelType(channel.Type),
	}, nil
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

func (c *Client) SetEmbed(channelId string, messageId string, embed discord.MessageEmbed) error {
	dgEmbed := &discordgo.MessageEmbed{

		Title:     embed.Title,
		Timestamp: embed.Timestamp,
		Color:     embed.Color,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: embed.Thumbnail,
		},
		Fields: []*discordgo.MessageEmbedField{},
	}

	for _, field := range embed.Fields {
		dgEmbed.Fields = append(dgEmbed.Fields, &discordgo.MessageEmbedField{
			Name:   field.Name,
			Value:  field.Value,
			Inline: false,
		})
	}

	_, err := c.session.ChannelMessageEditEmbed(channelId, messageId, dgEmbed)

	return err
}

func (c *Client) JoinVoiceChannel(guildId, channelId string, mute, deaf bool) (voice *discord.VoiceConnection, err error) {
	conn, err := c.session.ChannelVoiceJoin(guildId, channelId, mute, deaf)
	if err != nil {
		return nil, err
	}

	voiceRecv := make(chan *discord.Packet)
	go func(voice chan *discordgo.Packet) {
		for packet := range voice {
			voiceRecv <- &discord.Packet{
				SSRC:      packet.SSRC,
				Sequence:  packet.Sequence,
				Timestamp: packet.Timestamp,
				Type:      packet.Type,
				Opus:      packet.Opus,
				PCM:       packet.PCM,
			}
		}
	}(conn.OpusRecv)
	domainConn := discord.NewVoiceConnection(conn, voiceRecv)

	return domainConn, nil

}

func (c *Client) EndVoiceConnection(voice *discord.VoiceConnection) error {
	discordGoConn, ok := voice.Internals.(*discordgo.VoiceConnection)
	if !ok {
		log.Fatalln("couldnt cast to discordgo conn")
	}
	close(discordGoConn.OpusRecv)
	discordGoConn.Close()
	err := discordGoConn.Disconnect()
	close(voice.VoiceReceiver)
	return err
}

func (c *Client) SendFileMessage(channelId string, name, contentType string, readable io.Reader) (discord.Message, error) {
	sendComplex, err := c.session.ChannelMessageSendComplex(channelId, &discordgo.MessageSend{
		Files: []*discordgo.File{
			{
				Name:        name,
				ContentType: contentType,
				Reader:      readable,
			},
		},
	})
	if err != nil {
		return discord.Message{}, err
	}
	return discord.Message{
		Id:        sendComplex.ID,
		ChannelId: sendComplex.ChannelID,
	}, nil
}
