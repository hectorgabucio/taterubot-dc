package application

import (
	"errors"
	"fmt"
	"github.com/hectorgabucio/taterubot-dc/domain/discord/discordmocks"
	"github.com/hectorgabucio/taterubot-dc/localizations"
	"github.com/stretchr/testify/assert"
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
		name        string
		fields      fields
		args        args
		expectedOut error
		on          func(*fields)
		assertMocks func(t *testing.T, f *fields)
	}{
		{
			name:        "when get guilds fail, return error",
			fields:      fields{discordClient: &discordmocks.Client{}},
			args:        args{interactionToken: "token"},
			expectedOut: fmt.Errorf("err getting guilds, %w", errors.New("err guilds")),
			on: func(fields *fields) {
				fields.discordClient.On("GetGuilds").Return(nil, errors.New("err guilds"))
			},
			assertMocks: func(t *testing.T, f *fields) {
				f.discordClient.AssertNumberOfCalls(t, "GetGuilds", 1)
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

			if !assert.EqualErrorf(t, err, tt.expectedOut.Error(), "Error should be: %v, got: %v", tt.expectedOut.Error(), err) {
				t.Fail()
			}
			if tt.assertMocks != nil {
				tt.assertMocks(t, &tt.fields)
			}
		})
	}
}
