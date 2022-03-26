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
