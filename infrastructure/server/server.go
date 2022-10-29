package server

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bwmarrin/discordgo"
	"github.com/hectorgabucio/taterubot-dc/application"
	"github.com/hectorgabucio/taterubot-dc/kit/command"
	"github.com/robfig/cron/v3"
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

func (server *Server) Close() error {
	if err := server.session.Close(); err != nil {
		return fmt.Errorf("server.close:%w", err)
	}
	return nil
}

func (server *Server) installInteractions() {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "taterubot",
			Description: "I will say hi!",
			DescriptionLocalizations: &map[discordgo.Locale]string{
				discordgo.SpanishES: "Te explicaré de qué va esto!",
			},
		},
		{
			Name:        "stats",
			Description: "Let's see some cool stats about this discord server!",
			DescriptionLocalizations: &map[discordgo.Locale]string{
				discordgo.SpanishES: "Vamos a ver algunas estadísticas chulas de este servidor",
			},
		},
	}
	commandHandlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"taterubot": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if err := s.InteractionRespond(&discordgo.Interaction{ID: i.ID, Token: i.Token}, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "...",
				},
			}); err != nil {
				log.Println(err)
				return
			}
			go func() {
				err := server.commandBus.Dispatch(context.Background(), application.NewGreetingCommand(i.Token))
				if err != nil {
					log.Println("err greeting command", err)
				}
			}()
		},
		"stats": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			if err := s.InteractionRespond(&discordgo.Interaction{ID: i.ID, Token: i.Token}, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "...",
				},
			}); err != nil {
				log.Println(err)
				return
			}
			go func() {
				err := server.commandBus.Dispatch(context.Background(), application.NewStatsCommand(i.Token, i.GuildID))
				if err != nil {
					log.Println("err stats command", err)
				}
			}()

		},
	}
	guilds, err := server.session.UserGuilds(100, "", "")
	if err != nil {
		return
	}

	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for _, guild := range guilds {
		for i, v := range commands {
			cmd, err := server.session.ApplicationCommandCreate(server.session.State.User.ID, guild.ID, v)
			if err != nil {
				log.Fatalf("Cannot create '%v' command: %v", v.Name, err)
			}
			registeredCommands[i] = cmd
		}
	}

	server.session.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

func (server *Server) registerHandlers() {
	server.session.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is ready")
		server.installInteractions()
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
			if (r.SelfMute && !r.BeforeUpdate.SelfMute) || (!r.SelfMute && r.BeforeUpdate.SelfMute) {
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
	c := cron.New()

	c.AddFunc("0 10 1 * *", func() {
		server.session.ChannelMessageSend("673215160764334093", "GENRE HAY QUE GUARDAR DINERO PA JAPONNN")
	})

	c.Start()

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
