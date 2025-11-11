package tts

import (
	"github.com/kolonist/edgetts/internal/communication"
)

const (
	wssURL = "wss://" + communication.BaseURL + "/websocket/v1"
)

var (
	wssHeaders map[string]string = map[string]string{
		"Pragma":                 "no-cache",
		"Cache-Control":          "no-cache",
		"Origin":                 "chrome-extension://jdiccldimpdaibmpdkjnbmckianbfold",
		"Sec-WebSocket-Protocol": "synthesize",
	}
)
