package discord

import "io"

type ChannelType int

const (
	ChannelTypeGuildText  ChannelType = 0
	ChannelTypeGuildVoice ChannelType = 2
)

type Client interface {
	GetGuilds() ([]Guild, error)
	GetGuildUsers(guildID string) ([]User, error)
	GetUser(userID string) (User, error)
	GetBotUsername() string
	GetGuildChannels(guildID string) ([]Channel, error)
	GetChannel(channelID string) (Channel, error)
	CreateChannel(guildID string, name string, channelType ChannelType, maxUsers int) (Channel, error)
	SendTextMessage(channelID string, message string) error
	SendFileMessage(channelID string, name, contentType string, readable io.Reader) (Message, error)
	SetEmbed(channelID string, messageID string, embed MessageEmbed) error
	JoinVoiceChannel(guildID, channelID string, mute, deaf bool) (voice *VoiceConnection, err error)
	EndVoiceConnection(voice *VoiceConnection) error
	EditInteraction(token string, message string) error
	EditInteractionComplex(token string, edit ComplexInteractionEdit) error
}

type ComplexInteractionEdit struct {
	Content string
	Embeds  []*MessageEmbed
}

type User struct {
	ID          string
	Username    string
	AvatarURL   string
	AccentColor int
}

type Guild struct {
	ID   string
	Name string
}

type Channel struct {
	ID   string
	Name string
	Type ChannelType
}

type Message struct {
	ID        string
	ChannelID string
}

type MessageEmbed struct {
	URL         string
	Title       string
	Description string
	Timestamp   string
	Color       int
	Thumbnail   string
	Fields      []*MessageEmbedField
	Author      *MessageEmbedAuthor
}

type MessageEmbedAuthor struct {
	Name    string
	IconURL string
}

type MessageEmbedField struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

type Packet struct {
	SSRC      uint32
	Sequence  uint16
	Timestamp uint32
	Type      []byte
	Opus      []byte
	PCM       []int16
}

type VoiceConnection struct {
	Internals     any
	VoiceReceiver chan *Packet
}

func NewVoiceConnection(internals any, voice chan *Packet) *VoiceConnection {
	return &VoiceConnection{
		Internals:     internals,
		VoiceReceiver: voice,
	}
}
