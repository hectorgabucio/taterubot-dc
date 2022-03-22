package domain

import "os"

type MP3Decoder interface {
	GetDuration(file *os.File) float64
}
