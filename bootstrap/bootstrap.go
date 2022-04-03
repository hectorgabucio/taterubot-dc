package bootstrap

import (
	"context"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/hectorgabucio/taterubot-dc/application"
	"github.com/hectorgabucio/taterubot-dc/config"
	"github.com/hectorgabucio/taterubot-dc/domain"
	mp3decoder "github.com/hectorgabucio/taterubot-dc/infrastructure/decoder"
	discordwrapper "github.com/hectorgabucio/taterubot-dc/infrastructure/discordgo"
	"github.com/hectorgabucio/taterubot-dc/infrastructure/inmemory"
	"github.com/hectorgabucio/taterubot-dc/infrastructure/localfs"
	"github.com/hectorgabucio/taterubot-dc/infrastructure/pion"
	"github.com/hectorgabucio/taterubot-dc/infrastructure/rabbitmq"
	"github.com/hectorgabucio/taterubot-dc/infrastructure/server"
	"github.com/hectorgabucio/taterubot-dc/localizations"
	"github.com/spf13/viper"
	"log"
)

type Closer interface {
	Close()
}

func createServerAndDependencies() (context.Context, *server.Server, []Closer, error) {
	// CONFIG

	viper.SetDefault("LANGUAGE", "en")
	viper.SetDefault("CHANNEL_NAME", "TATERU")

	viper.SetConfigFile(`config.json`)
	viper.SetConfigType("json") // Look for specific type
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalln(err)
	}
	viper.AutomaticEnv()

	var cfg config.Config
	cfg.BotToken = viper.GetString("BOT_TOKEN")
	cfg.Language = viper.GetString("LANGUAGE")
	cfg.BasePath = viper.GetString("BASE_PATH")
	cfg.ChannelName = viper.GetString("CHANNEL_NAME")
	cfg.CloudAMQPUrl = viper.GetString("CLOUDAMQP_URL")

	// LOCALIZATION
	l := localizations.New(cfg.Language, "en")

	// INFRASTRUCTURE
	s, err := discordgo.New("Bot " + cfg.BotToken)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error getting new bot client, %w", err)
	}
	// We only really care about receiving voice state updates.
	s.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildVoiceStates)

	// eventBus := inmemory.NewEventBus()
	// commandBus := inmemory.NewCommandBus()

	eventBus, err := rabbitmq.NewEventBus(cfg.CloudAMQPUrl)
	if err != nil {
		log.Fatalln(err)
	}
	commandBus, err := rabbitmq.NewCommandBus(cfg.CloudAMQPUrl)
	if err != nil {
		log.Fatalln(err)
	}
	lockedUserRepo := inmemory.NewLockedUserRepository()
	fsRepo := localfs.NewRepository(cfg.BasePath)
	decoder := mp3decoder.NewMP3Decoder()

	discordClient := discordwrapper.NewClient(s)
	oggWriter := &pion.Writer{}

	// APPLICATION LAYER
	greeting := application.NewGreetingMessageCreator(discordClient, l, cfg.ChannelName)
	voice := application.NewVoiceRecorder(discordClient, cfg.ChannelName, lockedUserRepo, eventBus, fsRepo, oggWriter)
	embedAudioData := application.NewAddMetadataOnAudioSent(discordClient, l.GetWithLocale(cfg.Language, "texts.duration"), fsRepo, decoder, eventBus)
	removeFiles := application.NewRemoveFilesWhenNotNeeded(fsRepo)

	// EVENT SUBSCRIPTIONS
	eventBus.Subscribe(domain.AudioSentEventType, embedAudioData)
	eventBus.Subscribe(domain.DoneProcessingFilesEventType, removeFiles)

	// COMMAND HANDLING
	greetingCommandHandler := application.NewGreetingCommandHandler(greeting)
	commandBus.Register(application.GreetingCommandType, greetingCommandHandler)

	voiceCommandHandler := application.NewRecordingCommandHandler(voice)
	commandBus.Register(application.RecordingCommandType, voiceCommandHandler)

	ctx, srv := server.NewServer(context.Background(), s, commandBus)
	return ctx, srv, []Closer{
		commandBus, eventBus, srv,
	}, nil
}

func Run() error {
	ctx, srv, closers, err := createServerAndDependencies()
	defer closeResources(closers)
	if err != nil {
		return err
	}
	if err = srv.Run(ctx); err != nil {
		return fmt.Errorf("err running server, %w", err)
	}
	return nil
}

func closeResources(closers []Closer) {
	for _, closer := range closers {
		closer.Close()
	}
}
