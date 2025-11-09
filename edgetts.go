package edgetts

import (
	"github.com/kolonist/edgetts/internal/tts"
)

type Args tts.Args

func Transcribe(args Args) error {
	return tts.Transcribe(tts.Args(args))
}

func ListVoices() ([]tts.Voice, error) {
	return tts.ListVoices()
}
