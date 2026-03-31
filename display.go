package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

type Display struct {
	chip   *Chip8
	scale  int
	keypad *Keypad
}

func NewDisplay(chip *Chip8, scale int, keypad *Keypad) *Display {
	return &Display{
		chip:   chip,
		scale:  scale,
		keypad: keypad,
	}
}

func (display *Display) Update() error {
	display.chip.keypad.Update()

	// ebiten runs at 60 fps, cpu cycle will run at 600 (60*10) Hz
	// 600 instruction per second
	for i := 0; i < 10; i++ {
		display.chip.Cycle()

		// if DXYN (drw) instruction draws anything pause the cycle
		if display.chip.drawFlag {
			display.chip.drawFlag = false
			break
		}
	}
	return nil
}

// render chip8 display
func (display *Display) Draw(screen *ebiten.Image) {
	screen.Fill(color.Black)

	// 32x64 display
	for y := 0; y < ScreenHeight; y++ {
		for x := 0; x < ScreenWidth; x++ {
			// gfx is a 1D array
			// (x,y) -> index map : index = y * width + x
			if display.chip.gfx[y*64+x] == 1 {

				// scale each pixel
				for dy := 0; dy < display.scale; dy++ {
					for dx := 0; dx < display.scale; dx++ {
						screen.Set(
							x*display.scale+dx,
							y*display.scale+dy,
							color.White,
						)
					}
				}
			}
		}
	}
}

func (display *Display) Layout(outsideW, outsideH int) (int, int) {
	return ScreenWidth * display.scale, ScreenHeight * display.scale
}
