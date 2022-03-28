package pion

import (
	"github.com/hectorgabucio/taterubot-dc/domain/discord"
	"github.com/pion/rtp"
	"github.com/pion/webrtc/v3/pkg/media/oggwriter"
	"io"
	"log"
)

type PionWriter struct {
}

func (w *PionWriter) NewWriter(path string) (io.Closer, error) {
	return oggwriter.New(path, 48000, 2)
}

func (w *PionWriter) WriteVoice(writer io.Closer, packet *discord.Packet) error {
	pionOggWriter, ok := writer.(*oggwriter.OggWriter)
	if !ok {
		log.Fatal("Could not cast to pion oggwriter")
	}
	err := pionOggWriter.WriteRTP(createPionRTPPacket(packet))
	return err
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
