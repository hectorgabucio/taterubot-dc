package application

import (
	"context"
	"errors"
	"fmt"
	"github.com/hectorgabucio/taterubot-dc/domain"
	"github.com/hectorgabucio/taterubot-dc/domain/discord"
	"github.com/hectorgabucio/taterubot-dc/kit/command"
	"github.com/hectorgabucio/taterubot-dc/localizations"
	"log"
	"math/rand"
	"time"
)

const StatsCommandType command.Type = "command.stats"

type StatsCommand struct {
	InteractionToken string
	GuildID          string
}

func NewStatsCommand(interactionToken string, guildID string) StatsCommand {
	return StatsCommand{InteractionToken: interactionToken, GuildID: guildID}
}

func (c StatsCommand) Type() command.Type {
	return StatsCommandType
}

type StatsCommandHandler struct {
	service *StatsMessageCreator
}

// NewStatsCommandHandler initializes a new StatsCommandHandler.
func NewStatsCommandHandler(service *StatsMessageCreator) StatsCommandHandler {
	return StatsCommandHandler{
		service: service,
	}
}

// Handle implements the command.Handler interface.
func (h StatsCommandHandler) Handle(ctx context.Context, cmd command.Command) error {
	statsCmd, ok :=
		cmd.(StatsCommand)
	if !ok {
		return errors.New("unexpected command")
	}
	return h.service.send(statsCmd.InteractionToken, statsCmd.GuildID)
}

type StatsMessageCreator struct {
	discordClient discord.Client
	localization  *localizations.Localizer
	voiceDataRepo domain.VoiceDataRepository
}

func NewStatsMessageCreator(discord discord.Client, localization *localizations.Localizer, voiceDataRepo domain.VoiceDataRepository) *StatsMessageCreator {
	return &StatsMessageCreator{
		discord,
		localization,
		voiceDataRepo,
	}
}

func (service *StatsMessageCreator) send(interactionToken string, guildID string) error {
	log.Println("received stats command", interactionToken, guildID)
	now := time.Now()
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	lastOfMonth := firstOfMonth.AddDate(0, 1, -1)
	onRange, err := service.voiceDataRepo.GetOnRange(guildID, firstOfMonth, lastOfMonth)
	if err != nil {
		return fmt.Errorf("err getting voice range, %w", err)
	}
	if len(onRange) == 0 {
		return service.sendEmptyInteraction(interactionToken)
	}
	message, err := service.buildStatsMessage(onRange, guildID)
	if err != nil {
		return err
	}
	if err := service.discordClient.EditInteractionComplex(interactionToken, message); err != nil {
		log.Println(err)
		return fmt.Errorf("err sending interaction response, %w", err)
	}
	return nil
}

func (service *StatsMessageCreator) buildStatsMessage(voiceStats []domain.VoiceData, guildID string) (discord.ComplexInteractionEdit, error) {
	globalDuration := 0

	type userData struct {
		audiosSent        int
		totalAudioSeconds int

		longestAudioDuration int
	}
	usersData := make(map[string]*userData)

	for _, voiceData := range voiceStats {
		globalDuration += voiceData.Duration
		_, ok := usersData[voiceData.UserID]
		if !ok {
			usersData[voiceData.UserID] = &userData{
				audiosSent:           0,
				totalAudioSeconds:    0,
				longestAudioDuration: 0,
			}
		}
		user := usersData[voiceData.UserID]
		user.audiosSent++
		user.totalAudioSeconds += voiceData.Duration
		if voiceData.Duration > user.longestAudioDuration {
			user.longestAudioDuration = voiceData.Duration
		}
	}

	var longestAudioUser string
	var mostAudiosSentUser string
	for userID, userData := range usersData {
		if longestAudioUser == "" {
			longestAudioUser = userID
		}
		if mostAudiosSentUser == "" {
			mostAudiosSentUser = userID
		}
		if userData.audiosSent > usersData[mostAudiosSentUser].audiosSent {
			mostAudiosSentUser = userID
		}
	}

	embeds := make([]*discord.MessageEmbed, 0)

	user, err := service.discordClient.GetUser(longestAudioUser)
	if err != nil {
		return discord.ComplexInteractionEdit{}, fmt.Errorf("err.stats.get.user:%w", err)
	}
	embeds = append(embeds, service.buildAchievementEmbed(user,
		service.localization.Get("texts.achievement_longest_audio_title"),
		"https://images.emojiterra.com/google/android-11/512px/1fac1.png",
		service.localization.Get("texts.achievement_longest_audio_description", &localizations.Replacements{"seconds": usersData[longestAudioUser].longestAudioDuration})))

	user, err = service.discordClient.GetUser(mostAudiosSentUser)
	if err != nil {
		return discord.ComplexInteractionEdit{}, fmt.Errorf("err.stats.get.user:%w", err)
	}
	embeds = append(embeds, service.buildAchievementEmbed(user,
		service.localization.Get("texts.achievement_most_audios_sent_title"),
		"https://www.emojirequest.com/images/TalkingTooMuchEmoji.jpg",
		service.localization.Get("texts.achievement_most_audios_sent_description", &localizations.Replacements{"audios": usersData[user.ID].audiosSent})))

	guildUsers, err := service.discordClient.GetGuildUsers(guildID)
	if err != nil {
		return discord.ComplexInteractionEdit{}, fmt.Errorf("err.stats.get.guild.users:%w", err)
	}
	rand.Seed(time.Now().Unix()) // initialize global pseudo random generator
	randomUser := guildUsers[rand.Intn(len(guildUsers))]
	log.Println(randomUser)

	randomDescriptions := []string{
		"texts.achievement_random_description_1",
		"texts.achievement_random_description_2",
		"texts.achievement_random_description_3",
		"texts.achievement_random_description_4",
		"texts.achievement_random_description_5",
	}
	randomDescription := randomDescriptions[rand.Intn(len(randomDescriptions))]
	embeds = append(embeds, service.buildAchievementEmbed(user,
		service.localization.Get("texts.achievement_random_title"),
		"https://images.emojiterra.com/twitter/v13.1/512px/1f3b2.png",
		service.localization.Get(randomDescription)))

	medianDuration := globalDuration / len(voiceStats)
	message := service.localization.Get("texts.stats",
		&localizations.Replacements{"globalDuration": globalDuration, "globalAmount": len(voiceStats), "globalMedianDuration": medianDuration})

	return discord.ComplexInteractionEdit{
		Content: message,
		Embeds:  embeds,
	}, nil
}

func (service *StatsMessageCreator) buildAchievementEmbed(user discord.User, title, thumbnailURL, achievementText string) *discord.MessageEmbed {
	return &discord.MessageEmbed{
		Title:       title,
		Description: fmt.Sprintf("<@%s>", user.ID),
		Color:       user.AccentColor,
		Thumbnail:   thumbnailURL,
		Fields: []*discord.MessageEmbedField{
			{
				Name:  service.localization.Get("texts.achievement"),
				Value: achievementText,
			},
		},
		Author: &discord.MessageEmbedAuthor{
			Name:    user.Username,
			IconURL: user.AvatarURL,
		},
	}
}

func (service *StatsMessageCreator) sendEmptyInteraction(interactionToken string) error {
	if err := service.discordClient.EditInteraction(interactionToken, service.localization.Get("texts.stats-empty")); err != nil {
		return fmt.Errorf("err sending interaction response, %w", err)
	}
	return nil
}
