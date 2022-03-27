package mp3decoder

import (
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	_ "github.com/faiface/beep/mp3"
	"log"
	"os"
	"time"
)

type MP3Decoder struct {
}

func (M MP3Decoder) GetDuration(file *os.File) float64 {
	streamer, format, err := mp3.Decode(file)
	if err != nil {
		log.Fatal(err)
	}
	defer func(streamer beep.StreamSeekCloser) {
		err := streamer.Close()
		if err != nil {
			log.Println("err closing audio file", err)
		}
	}(streamer)
	length := format.SampleRate.D(streamer.Len())
	nTime := length.Round(time.Millisecond).Seconds()
	return nTime
}

func NewMP3Decoder() *MP3Decoder {
	return &MP3Decoder{}
}
