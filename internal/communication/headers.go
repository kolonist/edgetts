package communication

import "net/http"

const (
	BaseURL              = "api.msedgeservices.com/tts/cognitiveservices"
	ChromiumMajorVersion = "140"
	chromiumFullVersion  = ChromiumMajorVersion + ".0.3485.14"
	secMSGECVersion      = "1-" + chromiumFullVersion
	trustedClientToken   = "6A5AA1D4EAFF4E9FB37E23D68491D6F4"
)

var (
	baseHeaders map[string]string = map[string]string{
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36" +
			" (KHTML, like Gecko) Chrome/" + ChromiumMajorVersion + ".0.0.0 Safari/537.36" +
			" Edg/" + ChromiumMajorVersion + ".0.0.0",
		"Accept-Encoding": "gzip, deflate, br",
		"Accept-Language": "en-US,en;q=0.9",
	}
)

func SetHeaders(dest *http.Header, src map[string]string) {
	for k, v := range baseHeaders {
		dest.Set(k, v)
	}

	for k, v := range src {
		dest.Set(k, v)
	}
}
