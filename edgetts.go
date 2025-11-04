package edgetts

import (
	"fmt"

	"github.com/kolonist/edgetts/internal/tts"
)

type Args = tts.TTSArgs

func Speak(args Args) error {
	if err := validateArgs(args); err != nil {
		return err
	}

	edgeTTS, err := tts.NewTTS(args)
	if err != nil {
		return err
	}

	if err := edgeTTS.Speak(); err != nil {
		return err
	}

	return nil
}

func ListVoices() ([]tts.Voice, error) {
	return tts.ListVoices()
}

func validateArgs(args Args) error {
	if args.Text == "" {
		return fmt.Errorf("'Args.Text' should contain text to speak but empty string set")
	}

	if args.Voice == "" {
		return fmt.Errorf("'Args.Voice' should contain 'Voice.ShortName' to speak with but empty string set")
	}

	if args.WriteMedia == "" {
		return fmt.Errorf("'Args.WriteMedia' should contain mp3 filename speach should be saved to but empty string set")
	}

	return nil
}
