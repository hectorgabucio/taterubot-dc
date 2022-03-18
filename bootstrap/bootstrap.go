package bootstrap

import (
	"context"
	"github.com/bwmarrin/discordgo"
	"github.com/hectorgabucio/taterubot-dc/application"
	"github.com/hectorgabucio/taterubot-dc/config"
	"github.com/hectorgabucio/taterubot-dc/domain"
	"github.com/hectorgabucio/taterubot-dc/infrastructure/inmemory"
	"github.com/hectorgabucio/taterubot-dc/infrastructure/server"
	"github.com/hectorgabucio/taterubot-dc/localizations"
	"github.com/kelseyhightower/envconfig"
	"os"
)

func createServerAndDependencies() (error, context.Context, *server.Server) {
	var cfg config.Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return err, nil, nil
	}
	l := localizations.New(cfg.Language, "en")

	baseFilePath := cfg.BasePath
	if _, err := os.Stat(baseFilePath); os.IsNotExist(err) {
		err := os.Mkdir(baseFilePath, 0750)
		if err != nil {
			return err, nil, nil
		}
	}

	eventBus := inmemory.NewEventBus()

	lockedUserRepo := inmemory.NewLockedUserRepository()

	s, err := discordgo.New("Bot " + cfg.BotToken)
	if err != nil {
		return err, nil, nil
	}
	// We only really care about receiving voice state updates.
	s.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildVoiceStates)

	eventBus.Subscribe(
		domain.AudioSentEventType,
		application.NewAddMetadataOnAudioSent(s, l.GetWithLocale(cfg.Language, "texts.duration"), cfg.BasePath),
	)

	greeting := application.NewGreetingMessageCreator(s, l, cfg.ChannelName)

	voice := application.NewVoiceRecorder(s, cfg.ChannelName, lockedUserRepo, eventBus, cfg.BasePath)

	ctx, srv := server.NewServer(context.Background(), s, greeting, voice)
	return nil, ctx, &srv
}

func Run() error {
	err, ctx, srv := createServerAndDependencies()
	if err != nil {
		return err
	}
	return srv.Run(ctx)
}
