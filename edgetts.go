// Package edgetts contains structs and functions to generate speech from text
package edgetts

import (
	"github.com/kolonist/edgetts/internal/tts"
)

// Args contains parametsrs for text to speech synthesys
type Args = tts.Args

// EdgeTTS used to generate speech from text
type EdgeTTS struct {
	args Args
}

// New creates EdgeTTS struct with arguments to generate speech.
//
// Parameters:
//
//	args - parameters of speech generation like voice, rate, volume
//
// Returns:
//
//	New EdgeTTS struct
func New(args Args) *EdgeTTS {
	return &EdgeTTS{
		args: args,
	}
}

// Speak define text you need to convert to speech
//
// Parameters:
//
//	text - text to speak
//
// Returns:
//
//	speaker struct to use get synthesized sound
func (etts *EdgeTTS) Speak(text string) *Speaker {
	return &Speaker{
		text:     text,
		args:     etts.args,
		ready:    false,
		metadata: nil,
	}
}

// Speak define text you need to convert to speech
//
// Parameters:
//
//	text - text to speak
//	voice - voice os speaker to read text
//
// Returns:
//
//	speaker struct to use get synthesized sound
func (etts *EdgeTTS) SpeakWithVoice(text string, voice string) *Speaker {
	speaker := &Speaker{
		text:     text,
		args:     etts.args,
		ready:    false,
		metadata: nil,
	}

	speaker.args.Voice = voice

	return speaker
}
