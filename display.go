package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

type Display struct {
	chip  *Chip8
	scale int
}

// keymap for keypad.go
// TODO: implemet scancodes for different keyboard layouts
var keyMap = map[ebiten.Key]uint8{
	ebiten.Key1: 0x1,
	ebiten.Key2: 0x2,
	ebiten.Key3: 0x3,
	ebiten.Key4: 0xC,

	ebiten.KeyQ: 0x4,
	ebiten.KeyW: 0x5,
	ebiten.KeyE: 0x6,
	ebiten.KeyR: 0xD,

	ebiten.KeyA: 0x7,
	ebiten.KeyS: 0x8,
	ebiten.KeyD: 0x9,
	ebiten.KeyF: 0xE,

	ebiten.KeyZ: 0xA,
	ebiten.KeyX: 0x0,
	ebiten.KeyC: 0xB,
	ebiten.KeyV: 0xF,
}

func NewDisplay(chip *Chip8, scale int) *Display {
	return &Display{
		chip:  chip,
		scale: scale,
	}
}

func (display *Display) Update() error {
	for key, chipKey := range keyMap {
		if ebiten.IsKeyPressed(key) {
			display.chip.keypad.Press(chipKey)
		} else {
			display.chip.keypad.Release(chipKey)
		}
	}
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
