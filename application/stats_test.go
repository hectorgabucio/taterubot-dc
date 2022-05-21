package application

import (
	"errors"
	"testing"

	"github.com/hectorgabucio/taterubot-dc/domain"
	"github.com/hectorgabucio/taterubot-dc/domain/discord"
	discordmocks "github.com/hectorgabucio/taterubot-dc/domain/discord/mocks"
	domainmocks "github.com/hectorgabucio/taterubot-dc/domain/mocks"
	"github.com/hectorgabucio/taterubot-dc/localizations"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStatsMessageCreator_send(t *testing.T) {
	type fields struct {
		discordClient *discordmocks.Client
		localization  *localizations.Localizer
		voiceDataRepo *domainmocks.VoiceDataRepository
	}

	type args struct {
		interactionToken string
		guildID          string
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
			name:          "when get data on range fails, return error",
			expectedError: true,
			fields:        fields{discordClient: &discordmocks.Client{}, voiceDataRepo: &domainmocks.VoiceDataRepository{}},
			args:          args{interactionToken: "token", guildID: "1"},
			on: func(fields *fields) {
				fields.voiceDataRepo.On("GetOnRange", "1", mock.Anything, mock.Anything).Return(nil, errors.New("err range"))
			},
			assertMocks: func(t *testing.T, f *fields) {
				f.voiceDataRepo.AssertNumberOfCalls(t, "GetOnRange", 1)
			},
		},
		{
			name:          "when get data on range is empty, should send empty interaction",
			expectedError: false,
			fields:        fields{discordClient: &discordmocks.Client{}, voiceDataRepo: &domainmocks.VoiceDataRepository{}, localization: localizations.New("en", "en")},
			args:          args{interactionToken: "token", guildID: "1"},
			on: func(fields *fields) {
				fields.voiceDataRepo.On("GetOnRange", "1", mock.Anything, mock.Anything).Return([]domain.VoiceData{}, nil)
				fields.discordClient.On("EditInteraction", "token", mock.AnythingOfType("string")).Return(nil)
			},
			assertMocks: func(t *testing.T, f *fields) {
				f.voiceDataRepo.AssertNumberOfCalls(t, "GetOnRange", 1)
				f.discordClient.AssertNumberOfCalls(t, "EditInteraction", 1)

			},
		},
		{
			name:          "when building message, it should fail",
			expectedError: true,
			fields:        fields{discordClient: &discordmocks.Client{}, voiceDataRepo: &domainmocks.VoiceDataRepository{}, localization: localizations.New("en", "en")},
			args:          args{interactionToken: "token", guildID: "1"},
			on: func(fields *fields) {
				fields.voiceDataRepo.On("GetOnRange", "1", mock.Anything, mock.Anything).Return([]domain.VoiceData{{UserID: "1", Duration: 5}}, nil)
				fields.discordClient.On("GetUser", "1").Return(discord.User{}, errors.New("error"))
			},
			assertMocks: func(t *testing.T, f *fields) {
				f.voiceDataRepo.AssertNumberOfCalls(t, "GetOnRange", 1)
				f.discordClient.AssertNumberOfCalls(t, "GetUser", 1)

			},
		},
		{
			name:          "when building correct stats message, it should edit interaction complex",
			expectedError: false,
			fields:        fields{discordClient: &discordmocks.Client{}, voiceDataRepo: &domainmocks.VoiceDataRepository{}, localization: localizations.New("en", "en")},
			args:          args{interactionToken: "token", guildID: "1"},
			on: func(fields *fields) {
				fields.voiceDataRepo.On("GetOnRange", "1", mock.Anything, mock.Anything).Return([]domain.VoiceData{{UserID: "1", Duration: 5}}, nil)
				fields.discordClient.On("GetUser", "1").Return(discord.User{ID: "1"}, nil)
				fields.discordClient.On("GetGuildUsers", "1").Return([]discord.User{{ID: "1"}}, nil)
				fields.discordClient.On("EditInteractionComplex", "token", mock.AnythingOfType("discord.ComplexInteractionEdit")).Return(nil)

			},
			assertMocks: func(t *testing.T, f *fields) {
				f.voiceDataRepo.AssertNumberOfCalls(t, "GetOnRange", 1)
				f.discordClient.AssertNumberOfCalls(t, "GetUser", 2)
				f.discordClient.AssertNumberOfCalls(t, "GetGuildUsers", 1)
				f.discordClient.AssertNumberOfCalls(t, "EditInteractionComplex", 1)

			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &StatsMessageCreator{
				discordClient: tt.fields.discordClient,
				localization:  tt.fields.localization,
				voiceDataRepo: tt.fields.voiceDataRepo,
			}
			if tt.on != nil {
				tt.on(&tt.fields)
			}
			err := service.send(tt.args.interactionToken, tt.args.guildID)

			assert.Equal(t, tt.expectedError, err != nil)

			if tt.assertMocks != nil {
				tt.assertMocks(t, &tt.fields)
			}
		})
	}
}
