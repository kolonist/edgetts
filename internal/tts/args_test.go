package tts

import (
	"testing"
)

func Test_getVoice(t *testing.T) {
	tests := []struct {
		name    string
		args    Args
		want    string
		wantErr bool
	}{
		{
			name: "voice with region",
			args: Args{
				Voice: "zh-CN-henan-YundengNeural",
			},
			want:    "Microsoft Server Speech Text to Speech Voice (zh-CN-henan, YundengNeural)",
			wantErr: false,
		},
		{
			name: "voice without region",
			args: Args{
				Voice: "pt-PT-DuarteNeural",
			},
			want:    "Microsoft Server Speech Text to Speech Voice (pt-PT, DuarteNeural)",
			wantErr: false,
		},
		{
			name: "wrong voice 1",
			args: Args{
				Voice: "pt-PT-Duarte-Neural",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "wrong voice 2",
			args: Args{
				Voice: "DuarteNeural",
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "wrong voice 3",
			args: Args{
				Voice: "ro-RO-Emil",
			},
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			voice, err := tt.args.getVoice()
			t.Logf("\nvoice: '%v' \nerr: '%v'\n", voice, err)

			gotErr := err != nil
			if gotErr != tt.wantErr {
				if tt.wantErr {
					t.Error("expected to get error, but got nothing")
				} else {
					t.Errorf("expected error to be 'nil', but got '%v", err)
				}
			}

			if voice != tt.want {
				t.Errorf("expected voice to be '%v', but got '%v'", tt.want, voice)
			}
		})
	}
}
