package tts

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"iter"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/websocket"

	"github.com/kolonist/edgetts/internal/communication"
)

type speechParams struct {
	text   string
	voice  string
	rate   string
	volume string
}

type responseChunkType int8

const (
	chunkTypeAudio responseChunkType = iota
	chunkTypeWordBoundary
	chunkTypeSessionEnd
	chunkTypeEnd
)

type responseChunk struct {
	chunkType responseChunkType
	data      []byte
	metadata  SpeechMetadata
}

type SpeechMetadata struct {
	// Start time of word in generated sound in milliseconds
	Offset int `json:"offset"`

	// Duration of word pronunciation in milliseconds
	Duration int `json:"duration"`

	// Separate word
	Text string `json:"text"`
}

type readResponseResult struct {
	err error
}

type audioInfoText struct {
	Text         string `json:"Text"`
	Length       int    `json:"Length"`
	BoundaryType string `json:"BoundaryType"`
}

type audioInfo struct {
	Offset   int           `json:"Offset"`
	Duration int           `json:"Duration"`
	Text     audioInfoText `json:"text"`
}

type audioMetadata struct {
	Type string    `json:"Type"`
	Data audioInfo `json:"Data"`
}

type audioMetadataJSON struct {
	Metadata []audioMetadata `json:"Metadata"`
}

func Speak(args Args) error {
	// verify args and get TTS params
	speechParams, err := getSpeechParams(args)
	if err != nil {
		return err
	}

	// open destination files
	audioFile, err := createFile(args.AudioFile)
	if err != nil {
		return fmt.Errorf("can't create destination audio file: %v", err)
	}
	defer audioFile.Close()

	needWriteMeta := args.MetadataFile != ""

	var metadataFile *os.File
	if needWriteMeta {
		metadataFile, err = createFile(args.MetadataFile)
		if err != nil {
			return fmt.Errorf("can't create destination metadata file: %v", err)
		}
		defer metadataFile.Close()
	}

	// send request to Edge server
	conn, err := openWebsocket()
	if err != nil {
		return err
	}

	if err := sendRequest(conn, speechParams); err != nil {
		return err
	}

	// collect metadata to this var
	var metadata []SpeechMetadata
	if needWriteMeta {
		metadata = make([]SpeechMetadata, 0, 1024)
	}

	// read response from server
	readResult := readResponseResult{}

	for chunk := range readResponse(conn, &readResult) {
		switch chunk.chunkType {
		case chunkTypeAudio:
			// write sound data to audio file
			audioFile.Write(chunk.data)
		case chunkTypeWordBoundary:
			// collect metadata if need
			// todo: rewrite to json/v2 streaming when it come from beta
			if needWriteMeta {
				metadata = append(metadata, chunk.metadata)
			}
		}
	}

	if readResult.err != nil {
		return readResult.err
	}

	// write metadata to file
	if needWriteMeta {
		buf, err := json.Marshal(metadata)
		if err != nil {
			return fmt.Errorf("failed to marshal metadata JSON: %v", err)
		}

		if _, err := metadataFile.Write(buf); err != nil {
			return fmt.Errorf("failed to write to metadata file: %v", err)
		}
	}

	return nil
}

func getSpeechParams(args Args) (speechParams, error) {
	text, err := args.getText()
	if err != nil {
		return speechParams{}, err
	}

	voice, err := args.getVoice()
	if err != nil {
		return speechParams{}, err
	}

	rate, err := args.getRate()
	if err != nil {
		return speechParams{}, err
	}

	volume, err := args.getVolume()
	if err != nil {
		return speechParams{}, err
	}

	return speechParams{
		text:   text,
		voice:  voice,
		rate:   rate,
		volume: volume,
	}, nil
}

func openWebsocket() (*websocket.Conn, error) {
	headers := http.Header{}

	communication.SetHeaders(&headers, wssHeaders)

	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial(
		communication.GenerateSecURL(wssURL)+"&ConnectionId="+uuidWithoutDashes(),
		headers,
	)
	if err != nil {
		return nil, fmt.Errorf("dial error: %v", err)
	}
	return conn, nil
}

func sendRequest(conn *websocket.Conn, params speechParams) error {
	date := getCurrentTime()

	err := conn.WriteMessage(
		websocket.TextMessage,
		[]byte(
			"X-Timestamp:"+date+"\r\n"+
				"Content-Type:application/json; charset=utf-8\r\n"+
				"Path:speech.config\r\n\r\n"+
				`{"context":{"synthesis":{"audio":{"metadataoptions":{"sentenceBoundaryEnabled":false,"wordBoundaryEnabled":true},"outputFormat":"audio-24khz-48kbitrate-mono-mp3"}}}}`+"\r\n",
		),
	)
	if err != nil {
		return err
	}

	err = conn.WriteMessage(
		websocket.TextMessage,
		[]byte(
			ssmlHeadersPlusData(
				uuidWithoutDashes(),
				date,
				mkssml(params.text, params.voice, params.rate, params.volume),
			),
		),
	)
	if err != nil {
		return err
	}

	return nil
}

func readResponse(conn *websocket.Conn, result *readResponseResult) iter.Seq[responseChunk] {
	return func(yield func(responseChunk) bool) {
		// indicate that we are downloading audio data
		downloadAudio := false

		for {
			// read message
			messageType, data, err := conn.ReadMessage()
			if err != nil {
				result.err = err
				return
			}

			switch messageType {
			case websocket.TextMessage:
				path, data := getPathAndData(data)

				switch path {
				// start to receive audio binary
				case "turn.start":
					downloadAudio = true

				// end to receive data
				case "turn.end":
					chunk := responseChunk{
						chunkType: chunkTypeEnd,
					}
					if !yield(chunk) {
						result.err = nil
						return
					}

					result.err = nil
					return

				// receive metadata
				case "audio.metadata":
					audioMetadataJSON := audioMetadataJSON{}
					if err := json.Unmarshal(data, &audioMetadataJSON); err != nil {
						result.err = err
						return
					}

					for _, metadata := range audioMetadataJSON.Metadata {
						switch metadata.Type {
						case "WordBoundary":
							chunk := responseChunk{
								chunkType: chunkTypeWordBoundary,
								metadata: SpeechMetadata{
									Offset:   metadataDurationToMilliseconds(metadata.Data.Offset),
									Duration: metadataDurationToMilliseconds(metadata.Data.Duration),
									Text:     metadata.Data.Text.Text,
								},
							}
							if !yield(chunk) {
								result.err = nil
								return
							}
						case "SessionEnd":
							continue
						default:
							result.err = fmt.Errorf("unknown metadata type: %s", metadata.Type)
							return
						}
					}
				case "response":
				default:
					result.err = fmt.Errorf("response from Edge server not recognized: %s", data)
					return
				}
			case websocket.BinaryMessage:
				if !downloadAudio {
					result.err = fmt.Errorf("unexpected binary message received")
					return
				}

				if len(data) < 2 {
					result.err = fmt.Errorf("binary message lacks header length")
					return
				}

				headerLength := int(binary.BigEndian.Uint16(data[:2]))
				if len(data) < headerLength+2 {
					result.err = fmt.Errorf("binary message lacks audio data")
					return
				}

				chunk := responseChunk{
					chunkType: chunkTypeAudio,
					data:      data[headerLength+2:],
				}
				if !yield(chunk) {
					result.err = nil
					return
				}
			}
		}
	}
}

func metadataDurationToMilliseconds(time int) int {
	return time / 10_000
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
