package discordgo

import (
	"fmt"
	"io"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/hectorgabucio/taterubot-dc/domain/discord"
)

type Client struct {
	session *discordgo.Session
}

func (c *Client) GetGuildUsers(guildID string) ([]discord.User, error) {
	members, err := c.session.GuildMembers(guildID, "", 1000)
	if err != nil {
		return nil, fmt.Errorf("err.discordgo.listusers:%w", err)
	}
	users := make([]discord.User, len(members))
	for i, member := range members {
		users[i] = discord.User{
			ID:          member.User.ID,
			Username:    member.User.Username,
			AvatarURL:   member.User.AvatarURL(""),
			AccentColor: member.User.AccentColor,
		}
	}
	return users, nil
}

func (c *Client) GetUser(userID string) (discord.User, error) {
	user, err := c.session.User(userID)
	if err != nil {
		return discord.User{}, fmt.Errorf("err.discordgo.user:%w", err)
	}
	return discord.User{
		ID:          user.ID,
		Username:    user.Username,
		AvatarURL:   user.AvatarURL(""),
		AccentColor: user.AccentColor,
	}, nil
}

func (c *Client) EditInteractionComplex(token string, edit discord.ComplexInteractionEdit) error {
	embeds := make([]*discordgo.MessageEmbed, len(edit.Embeds))
	for i, embed := range edit.Embeds {
		fields := make([]*discordgo.MessageEmbedField, len(embed.Fields))
		for j, field := range embed.Fields {
			fields[j] = &discordgo.MessageEmbedField{
				Name:   field.Name,
				Value:  field.Value,
				Inline: false,
			}
		}
		embeds[i] = &discordgo.MessageEmbed{
			Title:       embed.Title,
			Description: embed.Description,
			Color:       embed.Color,
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: embed.Thumbnail,
			},
			Author: &discordgo.MessageEmbedAuthor{
				Name:    embed.Author.Name,
				IconURL: embed.Author.IconURL,
			},
			Fields: fields,
		}
	}
	_, err := c.session.InteractionResponseEdit(&discordgo.Interaction{Token: token, AppID: c.session.State.User.ID}, &discordgo.WebhookEdit{
		Content: &edit.Content,
		Embeds:  &embeds,
	})
	if err != nil {
		return fmt.Errorf("discordgo.interaction.edit.complex: %w", err)
	}
	return nil

}

func NewClient(session *discordgo.Session) *Client {
	return &Client{session: session}
}

func (c *Client) EditInteraction(token string, message string) error {
	_, err := c.session.InteractionResponseEdit(&discordgo.Interaction{Token: token, AppID: c.session.State.User.ID}, &discordgo.WebhookEdit{
		Content: &message,
	})
	if err != nil {
		return fmt.Errorf("discordgo.interaction.edit: %w", err)
	}
	return nil
}

func (c *Client) GetGuilds() ([]discord.Guild, error) {
	infraGuilds, err := c.session.UserGuilds(100, "", "")
	if err != nil {
		return nil, fmt.Errorf("err getting user guilds, %w", err)
	}

	guilds := make([]discord.Guild, len(infraGuilds))
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
	if _, err := c.session.ChannelMessageSend(channelID, message); err != nil {
		return fmt.Errorf("err sending channel message, %w", err)
	}
	return nil
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

	if _, err := c.session.ChannelMessageEditEmbed(channelID, messageID, dgEmbed); err != nil {
		return fmt.Errorf("err editing embed, %w", err)
	}
	return nil
}

func (c *Client) EstablishVoiceConnection(guildID, channelID string, mute, deaf bool, done chan bool) (voice *discord.VoiceConnection, err error) {
	conn, err := c.session.ChannelVoiceJoin(guildID, channelID, mute, deaf)
	if err != nil {
		return nil, fmt.Errorf("err joining voice channel, %w", err)
	}

	voiceRecv := make(chan *discord.Packet)
	go func(voice chan *discordgo.Packet) {
		for {
			if conn.Ready == false || conn.OpusRecv == nil {
				log.Printf("Discordgo not to receive opus packets. %+v : %+v", conn.Ready, conn.OpusSend)
				return
			}

			select {
			case packet, ok := <-voice:
				if !ok {
					return
				}
				voiceRecv <- &discord.Packet{
					SSRC:      packet.SSRC,
					Sequence:  packet.Sequence,
					Timestamp: packet.Timestamp,
					Type:      packet.Type,
					Opus:      packet.Opus,
					PCM:       packet.PCM,
				}
			case <-done: // done reading
				close(voiceRecv)
				//close(voice)

				err := conn.Disconnect()
				if err != nil {
					log.Printf("err disconnecting discord voice conn, %v", err)
				}
				return
			}
		}
	}(conn.OpusRecv)
	domainConn := discord.NewVoiceConnection(conn, voiceRecv)

	return domainConn, nil
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
		ID:           sendComplex.ID,
		ChannelID:    sendComplex.ChannelID,
		AttachmentId: sendComplex.Attachments[0].ID,
	}, nil
}
