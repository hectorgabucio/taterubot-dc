package pion

import (
	"fmt"
	"github.com/hectorgabucio/taterubot-dc/domain/discord"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3/pkg/media/oggwriter"
	"io"
	"log"
)

type Writer struct {
}

func (w *Writer) NewWriter(path string) (io.Closer, error) {
	writer, err := oggwriter.New(path, 48000, 2)
	if err != nil {
		return nil, fmt.Errorf("error getting new ogg writer: %w", err)
	}
	return writer, nil
}

func (w *Writer) WriteVoice(writer io.Closer, packet *discord.Packet) error {
	pionOggWriter, ok := writer.(*oggwriter.OggWriter)
	if !ok {
		log.Fatal("Could not cast to pion oggwriter")
	}
	err := pionOggWriter.WriteRTP(createPionRTPPacket(packet))
	return fmt.Errorf("err writing rtp packet, %w", err)
}

func createPionRTPPacket(p *discord.Packet) *rtp.Packet {
	return &rtp.Packet{
		Header: rtp.Header{
			Version: 2,
			// Taken from Discord voice docs
			PayloadType:    0x78,
			SequenceNumber: p.Sequence,
			Timestamp:      p.Timestamp,
			SSRC:           p.SSRC,
		},
		Payload: p.Opus,
	}
}
