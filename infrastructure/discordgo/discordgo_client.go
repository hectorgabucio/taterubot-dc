package discordgo

import (
	"fmt"
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
		return nil, fmt.Errorf("err getting user guilds, %w", err)
	}

	guilds := make([]discord.Guild, len(infraGuilds))
	// TODO create a map function
	for i, infraGuild := range infraGuilds {
		newGuild := discord.Guild{
			ID:   infraGuild.ID,
			Name: infraGuild.Name,
		}
		guilds[i] = newGuild
	}

	return guilds, nil
}

func (c *Client) GetBotUsername() string {
	return c.session.State.User.Username
}

func (c *Client) GetGuildChannels(guildID string) ([]discord.Channel, error) {
	channels, err := c.session.GuildChannels(guildID)
	if err != nil {
		return nil, fmt.Errorf("err getting guild channels, %w", err)
	}
	mappedChannels := make([]discord.Channel, len(channels))
	for i, infraChannel := range channels {
		newChannel := discord.Channel{
			ID:   infraChannel.ID,
			Name: infraChannel.Name,
			Type: discord.ChannelType(infraChannel.Type),
		}
		mappedChannels[i] = newChannel
	}

	return mappedChannels, nil
}
func (c *Client) GetChannel(channelID string) (discord.Channel, error) {
	channel, err := c.session.Channel(channelID)
	if err != nil {
		return discord.Channel{}, fmt.Errorf("err getting channel, %w", err)
	}
	return discord.Channel{
		ID:   channel.ID,
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
		return discord.Channel{}, fmt.Errorf("err creating channel, %w", err)
	}
	return discord.Channel{
		ID:   createdChannel.ID,
		Name: createdChannel.Name,
		Type: discord.ChannelType(createdChannel.Type),
	}, nil
}
func (c *Client) SendTextMessage(channelID string, message string) error {
	_, err := c.session.ChannelMessageSend(channelID, message)
	return fmt.Errorf("err sending channel message, %w", err)
}

func (c *Client) SetEmbed(channelID string, messageID string, embed discord.MessageEmbed) error {
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

	_, err := c.session.ChannelMessageEditEmbed(channelID, messageID, dgEmbed)

	return fmt.Errorf("err editing embed, %w", err)
}

func (c *Client) JoinVoiceChannel(guildID, channelID string, mute, deaf bool) (voice *discord.VoiceConnection, err error) {
	conn, err := c.session.ChannelVoiceJoin(guildID, channelID, mute, deaf)
	if err != nil {
		return nil, fmt.Errorf("err joining voice channel, %w", err)
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
	return fmt.Errorf("err disconnecting discord voice conn, %w", err)
}

func (c *Client) SendFileMessage(channelID string, name, contentType string, readable io.Reader) (discord.Message, error) {
	sendComplex, err := c.session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Files: []*discordgo.File{
			{
				Name:        name,
				ContentType: contentType,
				Reader:      readable,
			},
		},
	})
	if err != nil {
		return discord.Message{}, fmt.Errorf("err sending complex message, %w", err)
	}
	return discord.Message{
		ID:        sendComplex.ID,
		ChannelID: sendComplex.ChannelID,
	}, nil
}
