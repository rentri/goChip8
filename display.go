package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

type Display struct {
	chip  *Chip8
	scale int
}

func NewDisplay(chip *Chip8, scale int) *Display {
	return &Display{
		chip:  chip,
		scale: scale,
	}
}

func (display *Display) Update() error {
	// ebiten runs at 60 fps, cpu cycle will run at 720 (60*12) Hz
	// roughly 700 instruction per second
	for i := 0; i < 12; i++ {
		display.chip.Cycle()
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
