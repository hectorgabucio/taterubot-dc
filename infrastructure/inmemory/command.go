package inmemory

import (
	"context"
	"fmt"
	"github.com/hectorgabucio/taterubot-dc/kit/command"
)

// CommandBus is an in-memory implementation of the command.Bus.
type CommandBus struct {
	handlers map[command.Type]command.Handler
}

// NewCommandBus initializes a new instance of CommandBus.
func NewCommandBus() *CommandBus {
	return &CommandBus{
		handlers: make(map[command.Type]command.Handler),
	}
}

// Dispatch implements the command.Bus interface.
func (b *CommandBus) Dispatch(ctx context.Context, cmd command.Command) error {
	handler, ok := b.handlers[cmd.Type()]
	if !ok {
		return nil
	}

	if err := handler.Handle(ctx, cmd); err != nil {
		return fmt.Errorf("err invoking command handler, %w", err)
	}
	return nil
}

// Register implements the command.Bus interface.
func (b *CommandBus) Register(cmdType command.Type, handler command.Handler) {
	b.handlers[cmdType] = handler
}
