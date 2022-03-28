package ogg

import (
	"github.com/hectorgabucio/taterubot-dc/domain/discord"
	"io"
)

type Writer interface {
	NewWriter(path string) (io.Closer, error)
	WriteVoice(writer io.Closer, packet *discord.Packet) error
}
