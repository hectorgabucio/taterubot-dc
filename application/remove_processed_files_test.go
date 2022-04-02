package application

import (
	"context"
	"github.com/hectorgabucio/taterubot-dc/domain"
	"github.com/hectorgabucio/taterubot-dc/kit/event"
	"testing"
)

func TestRemoveFilesWhenNotNeeded_Handle(t *testing.T) {
	type fields struct {
		fsRepo domain.FileRepository
	}
	type args struct {
		ctx context.Context
		evt event.Event
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "",
			fields:  fields{},
			args:    args{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &RemoveFilesWhenNotNeeded{
				fsRepo: tt.fields.fsRepo,
			}
			if err := handler.Handle(tt.args.ctx, tt.args.evt); (err != nil) != tt.wantErr {
				t.Errorf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
