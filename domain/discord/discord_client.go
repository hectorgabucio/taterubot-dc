package discord

// ChannelType is the type of a Channel
type ChannelType int

// Block contains known ChannelType values
const (
	ChannelTypeGuildText  ChannelType = 0
	ChannelTypeGuildVoice ChannelType = 2
)

type Client interface {
	GetGuilds() ([]Guild, error)
	GetBotUsername() string
	GetGuildChannels(guildID string) ([]Channel, error)
	CreateChannel(guildID string, name string, channelType ChannelType, maxUsers int) (Channel, error)
	SendTextMessage(channelId string, message string) error
	SetEmbed(channelId string, messageId string, embed MessageEmbed) error
}

type Guild struct {
	Id   string
	Name string
}

type Channel struct {
	Id   string
	Name string
	Type ChannelType
}

type MessageEmbed struct {
	URL         string
	Title       string
	Description string
	Timestamp   string
	Color       int
	Thumbnail   string
	Fields      []*MessageEmbedField
}

type MessageEmbedField struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}
