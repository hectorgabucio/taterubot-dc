package mp3decoder

import (
	"github.com/tcolgate/mp3"
	"io"
	"log"
	"os"
)

type MP3Decoder struct {
}

func (M MP3Decoder) GetDuration(file *os.File) float64 {

	d := mp3.NewDecoder(file)
	var f mp3.Frame
	skipped := 0

	var t float64
	for {

		if err := d.Decode(&f, &skipped); err != nil {
			if err == io.EOF {
				break
			}
			log.Println(err)
			return 0
		}

		t = t + f.Duration().Seconds()
	}

	return t
}

func NewMP3Decoder() *MP3Decoder {
	return &MP3Decoder{}
}
