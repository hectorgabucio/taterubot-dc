package application

import (
	"errors"
	"github.com/hectorgabucio/taterubot-dc/domain/discord"
	"github.com/hectorgabucio/taterubot-dc/domain/discord/discordmocks"
	"github.com/hectorgabucio/taterubot-dc/localizations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

func TestGreetingMessageCreator_send(t *testing.T) {
	type fields struct {
		discordClient *discordmocks.Client
		localization  *localizations.Localizer
		channelName   string
	}
	type args struct {
		interactionToken string
	}
	tests := []struct {
		name          string
		fields        fields
		args          args
		expectedError bool
		on            func(*fields)
		assertMocks   func(t *testing.T, f *fields)
	}{
		{
			name:          "when get guilds fail, return error",
			fields:        fields{discordClient: &discordmocks.Client{}},
			args:          args{interactionToken: "token"},
			expectedError: true,
			on: func(fields *fields) {
				fields.discordClient.On("GetGuilds").Return(nil, errors.New("err guilds"))
			},
			assertMocks: func(t *testing.T, f *fields) {
				f.discordClient.AssertNumberOfCalls(t, "GetGuilds", 1)
			},
		},
		{
			name:          "when get guild channels fails, return error",
			fields:        fields{discordClient: &discordmocks.Client{}},
			args:          args{interactionToken: "token"},
			expectedError: true,
			on: func(fields *fields) {
				fields.discordClient.On("GetGuilds").Return([]discord.Guild{
					{ID: "1", Name: "1"},
				}, nil)
				fields.discordClient.On("GetBotUsername").Return("botUsername")
				fields.discordClient.On("GetGuildChannels", "1").Return(nil, errors.New("guild channel errors"))
			},
			assertMocks: func(t *testing.T, f *fields) {
				f.discordClient.AssertNumberOfCalls(t, "GetGuilds", 1)
			},
		},
		{
			name:          "send greeting message on a new created voice channel",
			fields:        fields{discordClient: &discordmocks.Client{}, channelName: "channelName", localization: localizations.New("en", "en")},
			args:          args{interactionToken: "token"},
			expectedError: false,
			on: func(fields *fields) {
				fields.discordClient.On("GetGuilds").Return([]discord.Guild{
					{ID: "1", Name: "guild-1"},
				}, nil)
				fields.discordClient.On("GetBotUsername").Return("botUsername")
				fields.discordClient.On("GetGuildChannels", "1").Return([]discord.Channel{
					{ID: "1", Name: "channel-1", Type: discord.ChannelTypeGuildText},
				}, nil)
				fields.discordClient.On("CreateChannel", "1", "channelName", discord.ChannelTypeGuildVoice, 2).Return(discord.Channel{
					ID:   "created-channel-id",
					Name: "created-channel-name",
					Type: discord.ChannelTypeGuildVoice,
				}, nil)
				fields.discordClient.On("EditInteraction", "token", mock.AnythingOfType("string")).Return(nil)
			},
			assertMocks: func(t *testing.T, f *fields) {
				f.discordClient.AssertNumberOfCalls(t, "GetGuilds", 1)
				f.discordClient.AssertNumberOfCalls(t, "EditInteraction", 1)
			},
		},
		{
			name:          "send greeting message to multiple guilds",
			fields:        fields{discordClient: &discordmocks.Client{}, channelName: "channelName", localization: localizations.New("en", "en")},
			args:          args{interactionToken: "token"},
			expectedError: false,
			on: func(fields *fields) {
				fields.discordClient.On("GetGuilds").Return([]discord.Guild{
					{ID: "1", Name: "guild-1"},
					{ID: "2", Name: "guild-2"},
					{ID: "3", Name: "guild-2"},
				}, nil)
				fields.discordClient.On("GetBotUsername").Return("botUsername")
				fields.discordClient.On("GetGuildChannels", mock.AnythingOfType("string")).Return([]discord.Channel{
					{ID: "1", Name: "channel-1", Type: discord.ChannelTypeGuildText},
					{ID: "2", Name: "channelName", Type: discord.ChannelTypeGuildVoice},
				}, nil)
				fields.discordClient.On("EditInteraction", "token", mock.AnythingOfType("string")).Return(nil)
			},
			assertMocks: func(t *testing.T, f *fields) {
				f.discordClient.AssertNumberOfCalls(t, "GetGuilds", 1)
				f.discordClient.AssertNumberOfCalls(t, "EditInteraction", 3)

			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &GreetingMessageCreator{
				discordClient: tt.fields.discordClient,
				localization:  tt.fields.localization,
				channelName:   tt.fields.channelName,
			}
			if tt.on != nil {
				tt.on(&tt.fields)
			}
			err := service.send(tt.args.interactionToken)

			assert.Equal(t, tt.expectedError, err != nil)

			if tt.assertMocks != nil {
				tt.assertMocks(t, &tt.fields)
			}
		})
	}
}
