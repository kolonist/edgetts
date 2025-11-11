package edgetts

import (
	"context"

	"github.com/kolonist/edgetts/internal/tts"
	"github.com/kolonist/edgetts/internal/voices"
)

// Arguments for TTS
type Args = tts.Args

// Voice description
type Voice = voices.Voice

func Speak(args Args) error {
	return tts.Speak(args)
}

func ListVoices(ctx context.Context) ([]Voice, error) {
	return voices.ListVoices(ctx)
}
