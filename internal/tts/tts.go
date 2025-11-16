package tts

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"iter"
	"net/http"

	"github.com/gorilla/websocket"

	"github.com/kolonist/edgetts/internal/communication"
)

type speechParams struct {
	text   string
	voice  string
	rate   string
	volume string
	format string
}

type responseChunkType int8

const (
	ChunkTypeAudio responseChunkType = iota
	ChunkTypeWordBoundary
	ChunkTypeSessionEnd
	ChunkTypeEnd
)

type responseChunk struct {
	ChunkType responseChunkType
	Data      []byte
	Metadata  SpeechMetadata
}

type SpeechMetadata struct {
	// Start time of word in generated sound in milliseconds
	Offset int `json:"offset"`

	// Duration of word pronunciation in milliseconds
	Duration int `json:"duration"`

	// Separate word
	Text string `json:"text"`
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

// SendRequest sends request to Edge TTS server and returns connection to read result
func SendRequest(ctx context.Context, text string, args Args, format OutputFormat) (*websocket.Conn, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// verify args and get TTS params
	speechParams, err := getSpeechParams(text, args, format)
	if err != nil {
		return nil, err
	}
	// send request to Edge server
	conn, err := openWebsocket(ctx)
	if err != nil {
		return nil, err
	}

	if err := sendWebsocketRequest(conn, speechParams); err != nil {
		return nil, err
	}

	return conn, nil
}

// ReadResponse read response from Edge TTS server
func ReadResponse(conn *websocket.Conn) iter.Seq2[responseChunk, error] {
	return func(yield func(responseChunk, error) bool) {
		// indicate that we are downloading audio data
		downloadAudio := false

		for {
			// read message
			messageType, data, err := conn.ReadMessage()
			if err != nil {
				yield(responseChunk{}, err)
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
						ChunkType: ChunkTypeEnd,
					}
					if !yield(chunk, nil) {
						return
					}

					return

				// receive metadata
				case "audio.metadata":
					audioMetadataJSON := audioMetadataJSON{}
					if err := json.Unmarshal(data, &audioMetadataJSON); err != nil {
						yield(responseChunk{}, err)
						return
					}

					for _, metadata := range audioMetadataJSON.Metadata {
						switch metadata.Type {
						case "WordBoundary":
							chunk := responseChunk{
								ChunkType: ChunkTypeWordBoundary,
								Metadata: SpeechMetadata{
									Offset:   metadataDurationToMilliseconds(metadata.Data.Offset),
									Duration: metadataDurationToMilliseconds(metadata.Data.Duration),
									Text:     metadata.Data.Text.Text,
								},
							}
							if !yield(chunk, nil) {
								return
							}
						case "SessionEnd":
							continue
						default:
							err = fmt.Errorf("unknown metadata type: %s", metadata.Type)
							yield(responseChunk{}, err)
							return
						}
					}
				case "response":
				default:
					err = fmt.Errorf("response from Edge TTS server not recognized: %s", data)
					yield(responseChunk{}, err)
					return
				}
			case websocket.BinaryMessage:
				if !downloadAudio {
					err = fmt.Errorf("unexpected binary message received")
					yield(responseChunk{}, err)
					return
				}

				if len(data) < 2 {
					err = fmt.Errorf("binary message lacks header length")
					yield(responseChunk{}, err)
					return
				}

				headerLength := int(binary.BigEndian.Uint16(data[:2]))
				if len(data) < headerLength+2 {
					err = fmt.Errorf("binary message lacks audio data")
					yield(responseChunk{}, err)
					return
				}

				chunk := responseChunk{
					ChunkType: ChunkTypeAudio,
					Data:      data[headerLength+2:],
				}
				if !yield(chunk, nil) {
					return
				}
			}
		}
	}
}

func getSpeechParams(text string, args Args, format OutputFormat) (speechParams, error) {
	if text == "" {
		return speechParams{}, fmt.Errorf("text not specified")
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

	params := speechParams{
		text:   text,
		voice:  voice,
		rate:   rate,
		volume: volume,
		format: format.String(),
	}

	return params, nil
}

func openWebsocket(ctx context.Context) (*websocket.Conn, error) {
	headers := http.Header{}

	communication.SetHeaders(&headers, wssHeaders)

	dialer := websocket.Dialer{}
	conn, _, err := dialer.DialContext(
		ctx,
		communication.GenerateSecURL(wssURL)+"&ConnectionId="+uuidWithoutDashes(),
		headers,
	)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func sendWebsocketRequest(conn *websocket.Conn, params speechParams) error {
	date := getCurrentTime()

	err := conn.WriteMessage(
		websocket.TextMessage,
		[]byte(
			"X-Timestamp:"+date+"\r\n"+
				"Content-Type:application/json; charset=utf-8\r\n"+
				"Path:speech.config\r\n\r\n"+
				`{"context":{"synthesis":{"audio":{"metadataoptions":{"sentenceBoundaryEnabled":false,"wordBoundaryEnabled":true},`+
				`"outputFormat":"`+params.format+`"}}}}`+"\r\n",
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

func metadataDurationToMilliseconds(time int) int {
	return time / 10_000
}
