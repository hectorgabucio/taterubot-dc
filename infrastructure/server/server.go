package server

import (
	"context"
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

func NewServer(ctx context.Context, session *discordgo.Session, commandBus command.Bus) (context.Context, *Server) {
	log.Println("Bot server running")

	srv := Server{session, commandBus}
	srv.registerHandlers()

	return serverContext(ctx), &srv
}

func (server *Server) Close() {
	if err := server.session.Close(); err != nil {
		log.Println("err closing session", err)
	}
}

func (server *Server) afterReady() {

	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "taterubot",
			Description: "I will say hi!",
		},
	}
	commandHandlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"taterubot": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: ":hand_splayed:",
				},
			}); err != nil {
				log.Println(err)
			}
			go func() {
				err := server.commandBus.Dispatch(context.Background(), application.NewGreetingCommand())
				if err != nil {
					log.Println("err greeting command", err)
				}
			}()
		},
	}
	guilds, err := server.session.UserGuilds(100, "", "")
	if err != nil {
		return
	}

	log.Println("Adding commands...")
	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for _, guild := range guilds {
		for i, v := range commands {
			cmd, err := server.session.ApplicationCommandCreate(server.session.State.User.ID, guild.ID, v)
			if err != nil {
				log.Panicf("Cannot create '%v' command: %v", v.Name, err)
			}
			registeredCommands[i] = cmd
		}
	}

	// server.session.ApplicationCommandDelete(server.session.State.User.ID, "", applicationCommands[0].ID)

	server.session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

func (server *Server) registerHandlers() {
	server.session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is ready")
		go func() {
			err := server.commandBus.Dispatch(context.Background(), application.NewGreetingCommand())
			if err != nil {
				log.Println("err greeting command", err)
			}
		}()

		server.afterReady()
	})

	server.session.AddHandler(func(s *discordgo.Session, r *discordgo.VoiceStateUpdate) {
		user, err := s.User(r.UserID)
		if err != nil {
			return
		}
		if user.Bot {
			return
		}

		if r.BeforeUpdate != nil {
			// to avoid messing with muting and unmuting
			if (r.SelfMute == true && r.BeforeUpdate.SelfMute == false) || (r.SelfMute == false && r.BeforeUpdate.SelfMute == true) {
				return
			}
		}

		go func() {
			err := server.commandBus.Dispatch(context.Background(), application.NewRecordingCommand(r.UserID, r.ChannelID, r.GuildID, user.Username, user.AvatarURL("")))
			if err != nil {
				log.Println("err recording command", err)
			}
		}()
	})
}

func (server *Server) Run(ctx context.Context) error {
	if err := server.session.Open(); err != nil {
		return fmt.Errorf("Cannot open the session: %w", err)
	}
	<-ctx.Done()
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("reason why context canceled, %w", err)
	}
	return nil
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
