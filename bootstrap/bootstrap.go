package bootstrap

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/hectorgabucio/taterubot-dc/application"
	"github.com/hectorgabucio/taterubot-dc/config"
	"github.com/hectorgabucio/taterubot-dc/domain"
	mp3decoder "github.com/hectorgabucio/taterubot-dc/infrastructure/decoder"
	discordwrapper "github.com/hectorgabucio/taterubot-dc/infrastructure/discordgo"
	"github.com/hectorgabucio/taterubot-dc/infrastructure/inmemory"
	"github.com/hectorgabucio/taterubot-dc/infrastructure/localfs"
	"github.com/hectorgabucio/taterubot-dc/infrastructure/pion"
	"github.com/hectorgabucio/taterubot-dc/infrastructure/server"
	sqlrepo "github.com/hectorgabucio/taterubot-dc/infrastructure/sql"
	"github.com/hectorgabucio/taterubot-dc/localizations"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"

	// side effect to add file support for db migrations
	_ "github.com/golang-migrate/migrate/v4/source/file"

	// side effect to load env file if any
	_ "github.com/joho/godotenv/autoload"
)

type Closer interface {
	Close() error
}

func setupSQLConnection(databaseURL string) *sqlx.DB {
	driverName := "postgres"
	
	db, err := sql.Open(driverName, databaseURL)
	if err != nil {
		log.Fatalln(err)
	}
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalln(err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://infrastructure/sql/migrations",
		driverName, driver)
	if err != nil {
		log.Fatalln("err preparing migrations", err)
	}
	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalln("err executing migrations", err)
	}
	log.Println("sql: all migrations run successfully")

	dbSQLX := sqlx.NewDb(db, driverName)
	if err := dbSQLX.Ping(); err != nil {
		log.Fatalln("err pinging conn", err)
	}
	return dbSQLX
}

func createServerAndDependencies() (context.Context, *server.Server, []Closer, error) {
	// CONFIG

	viper.SetDefault("LANGUAGE", "en")
	viper.SetDefault("CHANNEL_NAME", "TATERU")

	viper.SetConfigFile(`config.json`)
	viper.SetConfigType("json")
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
	cfg.DatabaseURL = viper.GetString("DATABASE_URL")

	// LOCALIZATION
	l := localizations.New(cfg.Language, "en")

	// INFRASTRUCTURE
	s, err := discordgo.New("Bot " + cfg.BotToken)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error getting new bot client, %w", err)
	}
	s.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildVoiceStates)

	eventBus := inmemory.NewEventBus()
	commandBus := inmemory.NewCommandBus()

	lockedUserRepo := inmemory.NewLockedUserRepository()
	fsRepo := localfs.NewRepository(cfg.BasePath)
	decoder := mp3decoder.NewMP3Decoder()

	discordClient := discordwrapper.NewClient(s)
	oggWriter := &pion.Writer{}

	db := setupSQLConnection(cfg.DatabaseURL)
	voiceDataRepo := sqlrepo.NewVoiceDataRepository(db)

	// APPLICATION LAYER
	greeting := application.NewGreetingMessageCreator(discordClient, l, cfg.ChannelName)
	stats := application.NewStatsMessageCreator(discordClient, l, voiceDataRepo)
	voice := application.NewVoiceRecorder(discordClient, cfg.ChannelName, lockedUserRepo, eventBus, fsRepo, oggWriter)
	embedAudioData := application.NewAddMetadataOnAudioSent(discordClient, l, fsRepo, voiceDataRepo, decoder, eventBus)
	removeFiles := application.NewRemoveFilesWhenNotNeeded(fsRepo)

	// EVENT SUBSCRIPTIONS
	eventBus.Subscribe(domain.AudioSentEventType, embedAudioData)
	eventBus.Subscribe(domain.DoneProcessingFilesEventType, removeFiles)

	// COMMAND HANDLING
	greetingCommandHandler := application.NewGreetingCommandHandler(greeting)
	commandBus.Register(application.GreetingCommandType, greetingCommandHandler)

	voiceCommandHandler := application.NewRecordingCommandHandler(voice)
	commandBus.Register(application.RecordingCommandType, voiceCommandHandler)

	statsCommandHandler := application.NewStatsCommandHandler(stats)
	commandBus.Register(application.StatsCommandType, statsCommandHandler)

	ctx, srv := server.NewServer(context.Background(), s, commandBus)
	return ctx, srv, []Closer{
		commandBus, eventBus, srv, db,
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
	log.Println("closing all resources...")
	for _, closer := range closers {
		err := closer.Close()
		if err != nil {
			log.Println("err closing resource:", err)
		}
	}
}
