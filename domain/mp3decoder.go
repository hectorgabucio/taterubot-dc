package domain

import "os"

type MP3Decoder interface {
	// GetDuration returns the duration of the mp3 file in seconds
	GetDuration(file *os.File) float64
}
