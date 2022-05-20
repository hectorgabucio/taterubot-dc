package application

import (
	"context"
	"testing"

	"github.com/hectorgabucio/taterubot-dc/domain"
	"github.com/hectorgabucio/taterubot-dc/domain/mocks"
	"github.com/hectorgabucio/taterubot-dc/kit/event"
	"github.com/stretchr/testify/mock"
)

func TestRemoveFilesWhenNotNeeded_Handle(t *testing.T) {
	type fields struct {
		fsRepo *mocks.FileRepository
	}
	type args struct {
		ctx context.Context
		evt event.Event
	}
	tests := []struct {
		name        string
		fields      fields
		args        args
		wantErr     bool
		on          func(*fields)
		assertMocks func(t *testing.T, f *fields)
	}{
		{
			name:    "unexpected event should return err",
			fields:  fields{},
			args:    args{},
			wantErr: true,
		},
		{
			name:    "correct event should trigger 3 files removal",
			fields:  fields{&mocks.FileRepository{}},
			args:    args{evt: domain.NewDoneProcessingFilesEvent("1")},
			wantErr: false,
			on: func(fields *fields) {
				fields.fsRepo.On("DeleteAll", mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return()
			},
			assertMocks: func(t *testing.T, f *fields) {
				f.fsRepo.AssertNumberOfCalls(t, "DeleteAll", 1)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &RemoveFilesWhenNotNeeded{
				fsRepo: tt.fields.fsRepo,
			}
			if tt.on != nil {
				tt.on(&tt.fields)
			}
			if err := handler.Handle(tt.args.ctx, tt.args.evt); (err != nil) != tt.wantErr {
				t.Errorf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.assertMocks != nil {
				tt.assertMocks(t, &tt.fields)
			}
		})
	}
}
