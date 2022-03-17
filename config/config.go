package config

type Config struct {
	BotToken    string `default:"token" split_words:"true"`
	ChannelName string `default:"TATERU" split_words:"true"`
	Language    string `default:"en" split_words:"true"`
	BasePath    string `default:"./tmp" split_words:"true"`
}
