package main

import (
	"fmt"

	"github.com/kolonist/edgetts"
)

func main() {
	fmt.Println("Trying to get voices list...")
	voices, err := edgetts.ListVoices()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	fmt.Println("Voices:")

	voice := ""
	for i, v := range voices {
		fmt.Printf(
			"    %d: locale: %s, gender: %s, short name: %s\n",
			i,
			v.Locale,
			v.Gender,
			v.ShortName,
		)

		if voice == "" && v.Locale == "en-US" && v.Gender == "Male" {
			voice = v.ShortName
		}
	}

	fmt.Println("")

	filename := "./sample.mp3"
	text := "edgetts is a golang module that allows you to use Microsoft Edge's online text-to-speech service directly from your golang code"
	fmt.Printf(
		"Speak '%s' to audio file '%s' using voice '%s'...\n",
		text,
		filename,
		voice,
	)

	args := edgetts.Args{
		Voice:        voice,
		Text:         text,
		Rate:         "+15%",
		AudioFile:    filename,
		MetadataFile: "./subtitles.json",
	}

	err = edgetts.Transcribe(args)
	if err != nil {
		fmt.Printf("Error trying to convert text to speach:\n%s\n", err.Error())
		return
	}

	fmt.Printf("Success! Listen transcription in '%s'\n", filename)
}
