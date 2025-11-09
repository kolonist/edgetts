# edgetts

`edgetts` is a golang module that allows you to use Microsoft Edge's online text-to-speech service directly from your golang code.

## Installation

To install it, run the following command:

    $ go install github.com/kolonist/edgetts

## Usage

To transcribe a text you need to import package `github.com/kolonist/edgetts` and then just use `edgetts.Transcribe()` function:

```go
package main

import (
	"github.com/kolonist/edgetts"
)

func main() {
	args := edgetts.Args{
		Text:      "Text I need to transcribe now",
        Voice:     "en-US-AlloyTurboMultilingualNeural",
		AudioFile: "./sample.mp3",
	}

    // read text 'args.Text' with voice 'args.Voice' and save it to 'args.AudioFile' mp3 file
	err = edgetts.Transcribe(args)
}
```



## Thanks

* https://github.com/rany2/edge-tts
* https://github.com/surfaceyu/edge-tts-go
