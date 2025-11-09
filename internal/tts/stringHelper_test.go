package tts

import (
	"bytes"
	"testing"
)

func Test_getPathAndData(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		args    args
		path    string
		data    []byte
		wantErr bool
	}{
		{
			name: "successful parse",
			args: args{
				data: []byte(
					"X-Timestamp:2022-01-01\r\n" +
						"Content-Type:application/json; charset=utf-8\r\n" +
						"Path:speech.config\r\n\r\n" +
						`{"context":{"synthesis":{"audio":{"metadataoptions":{"sentenceBoundaryEnabled":false,"wordBoundaryEnabled":true},"outputFormat":"audio-24khz-48kbitrate-mono-mp3"}}}}`,
				),
			},
			path:    "speech.config",
			data:    []byte(`{"context":{"synthesis":{"audio":{"metadataoptions":{"sentenceBoundaryEnabled":false,"wordBoundaryEnabled":true},"outputFormat":"audio-24khz-48kbitrate-mono-mp3"}}}}`),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, data := getPathAndData(tt.args.data)
			t.Logf("Path: %v\nData: %s\n", path, data)

			if path != tt.path {
				t.Errorf("expected path to be '%s' but got '%s'", tt.path, path)
			}

			if !bytes.Equal(data, tt.data) {
				t.Errorf("expected data to be '%s' but got '%s'", tt.data, data)
			}
		})
	}
}
