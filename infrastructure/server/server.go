package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/hectorgabucio/taterubot-dc/application"
	"github.com/hectorgabucio/taterubot-dc/config"
	"github.com/hectorgabucio/taterubot-dc/infrastructure/inmemory"
	"github.com/hectorgabucio/taterubot-dc/localizations"
	"log"
	"os"
	"os/signal"
)

type Server struct {
	config          config.Config
	localization    *localizations.Localizer
	session         *discordgo.Session
	greetingService *application.GreetingMessageCreator
	voiceService    *application.VoiceRecorder
}

func NewServer(ctx context.Context, l *localizations.Localizer, cfg config.Config) (context.Context, Server) {
	log.Println("Bot server running")

	s, err := discordgo.New("Bot " + cfg.BotToken)
	if err != nil {
		log.Fatal("Error initializing bot: " + err.Error())
	}
	// We only really care about receiving voice state updates.
	s.Identify.Intents = discordgo.MakeIntent(discordgo.IntentsGuildVoiceStates)

	lockedUserRepo := inmemory.New()

	greeting := application.NewGreetingMessageCreator(s, l, cfg.ChannelName)

	voice := application.NewVoiceRecorder(s, cfg.ChannelName, lockedUserRepo, cfg.BasePath, l.GetWithLocale(cfg.Language, "texts.duration"))

	srv := Server{cfg, l, s, greeting, voice}
	srv.registerHandlers()

	return serverContext(ctx), srv
}

func (server *Server) registerHandlers() {
	server.session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is ready")

		server.greetingService.Send()

	})

	done := make(chan bool)
	server.session.AddHandler(func(s *discordgo.Session, r *discordgo.VoiceStateUpdate) {
		user, err := s.User(r.UserID)
		if err != nil {
			return
		}
		if user.Bot {
			return
		}

		server.voiceService.Algo(r.UserID, r.ChannelID, r.GuildID, user, done)

	})
}

func (server *Server) Run(ctx context.Context) error {

	err := server.session.Open()
	if err != nil {
		return errors.New(fmt.Sprintf("Cannot open the session: %v", err))
	}
	defer func(s *discordgo.Session) {
		err := s.Close()
		if err != nil {
			log.Println("err closing session", err)
		}
	}(server.session)

	<-ctx.Done()
	return ctx.Err()

}

func serverContext(ctx context.Context) context.Context {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		<-c
		cancel()
	}()

	return ctx
}
