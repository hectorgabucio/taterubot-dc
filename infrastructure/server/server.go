package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/hectorgabucio/taterubot-dc/application"
	"github.com/hectorgabucio/taterubot-dc/kit/command"
	"log"
	"os"
	"os/signal"
)

type Server struct {
	session    *discordgo.Session
	commandBus command.Bus
}

func NewServer(ctx context.Context, session *discordgo.Session, commandBus command.Bus) (context.Context, Server) {
	log.Println("Bot server running")

	srv := Server{session, commandBus}
	srv.registerHandlers()

	return serverContext(ctx), srv
}

func (server *Server) registerHandlers() {
	server.session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is ready")

		err := server.commandBus.Dispatch(context.Background(), application.NewGreetingCommand())
		if err != nil {
			log.Println("err greeting command", err)
			return
		}

	})

	server.session.AddHandler(func(s *discordgo.Session, r *discordgo.VoiceStateUpdate) {
		user, err := s.User(r.UserID)
		if err != nil {
			return
		}
		if user.Bot {
			return
		}
		err = server.commandBus.Dispatch(context.Background(), application.NewRecordingCommand(r.UserID, r.ChannelID, r.GuildID, user.Username, user.AvatarURL("")))
		if err != nil {
			log.Println("err recording command", err)
			return
		}

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
