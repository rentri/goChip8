package main

import (
	"bytes"
	"encoding/binary"
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2/audio"
)

// ebiten audio reference https://ebitengine.org/en/examples/sinewave.html

const hertz = 44100 // 44.1KHz

var audioContext = audio.NewContext(hertz)
var player = generateBeep()

func Beep() {
	player.Rewind()
	player.Play()
}

func generateBeep() *audio.Player {
	sampleRate := hertz // 44100 samples per second
	duration := 0.2     // seconds
	freq := 330.0

	length := int(float64(sampleRate) * duration)

	// buffer for beep sound
	buf := make([]byte, length*2) // 16-bit mono

	// generate audio sample
	for i := 0; i < length; i++ {
		// phase in range [0,1)
		phase := math.Mod(float64(i)*freq/float64(sampleRate), 1.0)

		// triangle wave
		var v float64
		if phase < 0.5 {
			v = 4*phase - 1 // [-1, 1]
		} else {
			v = -4*phase + 3 // [1, -1]
		}

		fadeSamples := int(0.005 * float64(sampleRate)) // 5ms

		// fade in
		if i < fadeSamples {
			v *= float64(i) / float64(fadeSamples)
		}

		// fade out
		if i > length-fadeSamples {
			v *= float64(length-i) / float64(fadeSamples)
		}

		sample := int16(v * 20000) // convert to PCM amplitude

		// write to audio buffer
		binary.LittleEndian.PutUint16(buf[i*2:], uint16(sample))
	}

	// create a audio.Player to read the the byte slice buf which has now been converted to a stream
	player, err := audioContext.NewPlayer(bytes.NewReader(buf))
	if err != nil {
		log.Fatal(err)
	}
	return player
}
