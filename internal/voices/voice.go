package voices

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/kolonist/edgetts/internal/communication"
)

type Voice struct {
	// Voice full name
	Name string `json:"Name"`

	// Voice short name. You should use this value when specifying voice in this library functions
	ShortName string `json:"ShortName"`

	// Speaker gender
	Gender string `json:"Gender"`

	// Locale
	Locale string `json:"Locale"`

	// Is almost always an empty string
	SuggestedCodec string `json:"SuggestedCodec"`

	// Voice friendly name, is almost always an empty string
	FriendlyName string `json:"FriendlyName"`

	// GA for General Availability or Preview
	Status string `json:"Status"`

	// Additional information
	VoiceTag VoiceTag `json:"VoiceTag"`
}

type VoiceTag struct {
	// Is almost always an empty slice
	ContentCategories []string `json:"ContentCategories"`

	// Vocal characteristics of voice, is often an empty slice
	VoicePersonalities []string `json:"VoicePersonalities"`
}

func ListVoices(ctx context.Context) ([]Voice, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	client := http.Client{}
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		communication.GenerateSecURL(voiceListUrl),
		nil,
	)
	if err != nil {
		return nil, err
	}

	communication.SetHeaders(&req.Header, voiceHeaders)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var voices []Voice
	err = json.Unmarshal(body, &voices)
	if err != nil {
		return nil, err
	}

	return voices, nil
}
