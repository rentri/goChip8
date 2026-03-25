package main

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
)

const scale = 20

func main() {
	chip := NewChip()

	rom := filepath.Join("testRoms", "2-ibm-logo.ch8")
	err := chip.LoadRom(rom)
	if err != nil {
		log.Fatal(err)
	}

	// create a ticker that ticks at 60Hz
	// run parrallel to our cpu cycle of 700Hz
	go func() {
		ticker := time.NewTicker(time.Second / 60)
		defer ticker.Stop()

		for range ticker.C {
			if chip.DT > 0 {
				chip.DT--
			}

			if chip.ST > 0 {
				chip.ST--
				fmt.Println("BEEP")
			}
		}
	}()

	display := NewDisplay(chip, scale)

	ebiten.SetWindowSize(ScreenWidth*scale, ScreenHeight*scale) // x 64, y 32
	ebiten.SetWindowTitle("Chip8")

	if err := ebiten.RunGame(display); err != nil {
		log.Fatal(err)
	}
}
