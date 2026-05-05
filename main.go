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

	if len(os.Args) > 1 {
	    err := chip.LoadRom(os.Args[1])
	    if err != nil {
	        log.Fatal(err)
	    }
	}
	// ticker refrences chip from display
	display := NewDisplay(chip, scale, keypad)

	// create a ticker that ticks at 60Hz
	// run parrallel to our cpu cycle of 600Hz
	// ticker always uses currect chip
	go func() {
		ticker := time.NewTicker(time.Second / 60)
		defer ticker.Stop()

		for range ticker.C {
			if display.chip.DT > 0 {
				display.chip.DT--
			}

			if display.chip.ST > 0 {
				display.chip.ST--
				Beep()
			}
		}
	}()

	ebiten.SetWindowSize(ScreenWidth*scale, ScreenHeight*scale) // x 64, y 32
	ebiten.SetWindowTitle("Chip8")

	if err := ebiten.RunGame(display); err != nil {
		log.Fatal(err)
	}
}
