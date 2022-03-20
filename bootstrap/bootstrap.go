package bootstrap

import (
	"context"
	"github.com/bwmarrin/discordgo"
	"github.com/hectorgabucio/taterubot-dc/application"
	"github.com/hectorgabucio/taterubot-dc/config"
	"github.com/hectorgabucio/taterubot-dc/domain"
	"github.com/hectorgabucio/taterubot-dc/infrastructure/inmemory"
	"github.com/hectorgabucio/taterubot-dc/infrastructure/localfs"
	"github.com/hectorgabucio/taterubot-dc/infrastructure/server"
	"github.com/hectorgabucio/taterubot-dc/localizations"
	"github.com/kelseyhightower/envconfig"
)

func createServerAndDependencies() (error, context.Context, *server.Server) {
	// CONFIG
	var cfg config.Config
	err := envconfig.Process("", &cfg)
	if err != nil {
		return err, nil, nil
	}
	l := localizations.New(cfg.Language, "en")

	// INFRASTRUCTURE
	s, err := discordgo.New("Bot " + cfg.BotToken)
	if err != nil {
		return err, nil, nil
	}
	// We only really care about receiving voice state updates.
	s.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildVoiceStates)

	eventBus := inmemory.NewEventBus()
	commandBus := inmemory.NewCommandBus()
	lockedUserRepo := inmemory.NewLockedUserRepository()
	fsRepo := localfs.NewRepository(cfg.BasePath)

	// APPLICATION LAYER
	greeting := application.NewGreetingMessageCreator(s, l, cfg.ChannelName)
	voice := application.NewVoiceRecorder(s, cfg.ChannelName, lockedUserRepo, eventBus, fsRepo)
	embedAudioData := application.NewAddMetadataOnAudioSent(s, l.GetWithLocale(cfg.Language, "texts.duration"), fsRepo)

	// EVENT SUBSCRIPTIONS
	eventBus.Subscribe(domain.AudioSentEventType, embedAudioData)

	// COMMAND HANDLING
	greetingCommandHandler := application.NewGreetingCommandHandler(greeting)
	commandBus.Register(application.GreetingCommandType, greetingCommandHandler)

	// TODO code smell, the server would need command bus, and would trigger gretting and voice command
	ctx, srv := server.NewServer(context.Background(), s, commandBus, voice)
	return nil, ctx, &srv
}

func Run() error {
	err, ctx, srv := createServerAndDependencies()
	if err != nil {
		return err
	}
	return srv.Run(ctx)
}
