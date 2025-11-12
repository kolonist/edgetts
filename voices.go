package edgetts

import (
	"context"

	"github.com/kolonist/edgetts/internal/voices"
)

// Voice description
type Voice = voices.Voice

// ListVoices gets list of all available voices to use in speech generation.
// Parameters:
//
//	ctx - context to stop request before it finished
//
// Returns:
//
//	slice of voices
func ListVoices(ctx context.Context) ([]Voice, error) {
	return voices.ListVoices(ctx)
}
