package voices

import (
	"github.com/kolonist/edgetts/internal/communication"
)

const (
	voiceListUrl = "https://" + communication.BaseURL + "/voices/list"
)

var (
	voiceHeaders map[string]string = map[string]string{
		"Authority": "speech.platform.bing.com",
		"Sec-CH-UA": `" Not;A Brand";v="99", "Microsoft Edge";v="` + communication.ChromiumMajorVersion + `",` +
			` "Chromium";v="` + communication.ChromiumMajorVersion + `"`,
		"Sec-CH-UA-Mobile": "?0",
		"Accept":           "*/*",
		"Sec-Fetch-Site":   "none",
		"Sec-Fetch-Mode":   "cors",
		"Sec-Fetch-Dest":   "empty",
	}
)
