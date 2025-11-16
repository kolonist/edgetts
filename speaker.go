package edgetts

import (
	"context"
	"fmt"
	"iter"
	"os"
	"path/filepath"

	"github.com/kolonist/edgetts/internal/tts"
)

// Speaker used to get synthesized sound
type Speaker struct {
	text     string
	args     Args
	ready    bool
	metadata []SpeechMetadata
}

// SpeechMetadata contains time of word start and its pronunciation duration
type SpeechMetadata = tts.SpeechMetadata

// OutputFormat represents sound data output format
type OutputFormat = tts.OutputFormat

// Sound data output formats possible values
const (
	// mp3 24khz, 48k bitrate (default)
	OutputFormatMp3 = tts.OutputFormatMp3

	// webm 24khz, 16bit, 24k bitrate
	OutputFormatWebm = tts.OutputFormatWebm

	// ogg 24khz, 16bit
	OutputFormatOgg = tts.OutputFormatOgg

	// raw PCM 22050 hz, 16bit
	OutputFormatRaw22050 = tts.OutputFormatRaw22050

	// raw PCM 44100 hz, 16bit
	OutputFormatRaw44100 = tts.OutputFormatRaw44100
)

// GetSoundIter generate speech and return it in iterator with small byte buffers as they come from server.
//
// Parameters:
//
//	ctx - context to stop operation before it finished
//	format - format of sound data to write to file. Use one of OutputFormat* constants
//
// Returns:
//
//	error if file was not written for some reason
func (s *Speaker) GetSoundIter(ctx context.Context, format OutputFormat) iter.Seq2[[]byte, error] {
	return func(yield func([]byte, error) bool) {
		if err := ctx.Err(); err != nil {
			yield(nil, err)
			return
		}

		conn, err := tts.SendRequest(ctx, s.text, s.args, format)
		if err != nil {
			yield(nil, err)
			return
		}

		capacity := getWordsCount(s.text)
		metadata := make([]SpeechMetadata, 0, capacity)

		for chunk, err := range tts.ReadResponse(conn) {
			if err != nil {
				yield(nil, err)
				return
			}

			switch chunk.ChunkType {
			case tts.ChunkTypeAudio:
				if !yield(chunk.Data, nil) {
					return
				}
			case tts.ChunkTypeWordBoundary:
				metadata = append(metadata, chunk.Metadata)
			}
		}

		s.metadata = metadata
		s.ready = true
	}
}

// GetSound generate speech and returns it as bytes.
//
// Parameters:
//
//	ctx - context to stop operation before it finished
//	format - format of sound data to write to file. Use one of OutputFormat* constants
//
// Returns:
//
//   - buffer containing sound data of defined format
//   - error if file was not written for some reason
func (s *Speaker) GetSound(ctx context.Context, format OutputFormat) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	capacity := getBytesCount(s.text, format)
	result := make([]byte, 0, capacity)

	for data, err := range s.GetSoundIter(ctx, format) {
		if err != nil {
			return nil, err
		}

		result = append(result, data...)
	}

	return result, nil
}

// SaveToFile generate speach and save it to file.
//
// Parameters:
//
//	ctx - context to stop operation before it finished
//	filename - full path to file. Should be write accessible
//	format - format of sound data to write to file. Use one of OutputFormat* constants
//
// Returns:
//
//	error if file was not written for some reason
func (s *Speaker) SaveToFile(ctx context.Context, filename string, format OutputFormat) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	file, err := createFile(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	for data, err := range s.GetSoundIter(ctx, format) {
		if err != nil {
			return err
		}

		if _, err := file.Write(data); err != nil {
			return err
		}
	}

	return nil
}

// GetMetadata gets metadata of generated speech
//
// Returns:
//
//   - Slice with SpeechMetadata structs containing timings of each word in text
//   - Error
func (s *Speaker) GetMetadata() ([]SpeechMetadata, error) {
	if s.ready {
		return s.metadata, nil
	}

	return nil, fmt.Errorf("speech synthesys not finished, first call one of SaveToFile() or GetSpeech() methods")
}

// getWordsCount calculate approximate count of words to preallocate Metadata of right capacity
// and avoid lots of slice reallocations
func getWordsCount(text string) int {
	// median length of word in English is 4 + 1 space/punctuation character per word
	const wordLen = 5

	return len(text) / wordLen
}

// getBytesCount calculate approximate size of sound data to prealoocate audio buffer of right capacity
// and avoid lots of slice reallocations
func getBytesCount(text string, format OutputFormat) int {
	// duration of 1000 symbols is approximately 100 seconds so 10 symbols can be spoken in 1 second
	const symbolsPerSecond = 10

	durationSec := len(text) / symbolsPerSecond
	bytesPerSec := getBytesPerSecond(format)

	return durationSec * bytesPerSec
}

func getBytesPerSecond(format OutputFormat) int {
	switch format {
	case OutputFormatWebm:
		return 24_000 / 8 // 24k btrate
	case OutputFormatRaw22050:
		return 22_050 * 16 / 8 // 22050 hz, 16 bit
	case OutputFormatRaw44100:
		return 44_100 * 16 / 8 // 44100 hz, 16 bit
	case OutputFormatMp3:
		fallthrough
	case OutputFormatOgg:
		fallthrough
	default:
		return 48_000 / 8 // 48k btrate
	}
}

func createFile(path string) (*os.File, error) {
	// create directory
	if _, err := os.Stat(path); os.IsNotExist(err) {
		dir := filepath.Dir(path)

		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return nil, fmt.Errorf("failed to create dir '%s': %v", dir, err)
		}
	} else if err != nil {
		return nil, err
	}

	// open file
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("failed to open file '%s': %v", path, err)
	}

	return file, nil
}
