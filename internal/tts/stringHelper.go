package tts

import (
	"bytes"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
)

var (
	eol        []byte = []byte("\r\n")
	doubleEol  []byte = []byte("\r\n\r\n")
	pathHeader []byte = []byte("Path:")
)

func uuidWithoutDashes() string {
	id := uuid.New()
	return hex.EncodeToString(id[:])
}

func mkssml(text string, voice string, rate string, volume string) string {
	return "<speak version='1.0' xmlns='http://www.w3.org/2001/10/synthesis' xml:lang='en-US'>" +
		"<voice name='" + voice + "'>" +
		"<prosody pitch='+0Hz' rate='" + rate + "' volume='" + volume + "'>" + text + "</prosody>" +
		"</voice></speak>"
}

func ssmlHeadersPlusData(requestID string, timestamp string, ssml string) string {
	return "X-RequestId:" + requestID + "\r\n" +
		"Content-Type:application/ssml+xml\r\n" +
		"X-Timestamp:" + timestamp + "Z\r\n" +
		"Path:ssml\r\n\r\n" +
		ssml
}

func getCurrentTime() string {
	return time.Now().UTC().Format("Mon Jan 02 2006 15:04:05 GMT-0700 (Coordinated Universal Time)")
}

func getPathAndData(data []byte) (string, []byte) {
	var path string

	lines := bytes.SplitSeq(data[:bytes.Index(data, doubleEol)], eol)

	for line := range lines {
		if bytes.Index(line, pathHeader) == 0 {
			path = string(bytes.TrimSpace(line[len(pathHeader):]))
			break
		}
	}

	return path, data[bytes.Index(data, doubleEol)+4:]
}
