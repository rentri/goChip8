package main

import (
	"log"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

const scale = 20

func main() {
	keypad := &Keypad{}
	chip := NewChip(keypad)

	rom := os.Args[1]

	err := chip.LoadRom(rom)
	if err != nil {
		log.Fatal(err)
	}

	// create a ticker that ticks at 60Hz
	// run parrallel to our cpu cycle of 600Hz
	go func() {
		ticker := time.NewTicker(time.Second / 60)
		defer ticker.Stop()

		for range ticker.C {
			if chip.DT > 0 {
				chip.DT--
			}

			if chip.ST > 0 {
				chip.ST--
				Beep()
			}
		}
	}()

	display := NewDisplay(chip, scale, keypad)

	ebiten.SetWindowSize(ScreenWidth*scale, ScreenHeight*scale) // x 64, y 32
	ebiten.SetWindowTitle("Chip8")

	if err := ebiten.RunGame(display); err != nil {
		log.Fatal(err)
	}
}
