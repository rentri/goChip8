package main

import "github.com/hajimehoshi/ebiten/v2"

type Keypad struct {
	keys [16]bool
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

func (keypad *Keypad) Update() {
	for i := 0; i < 16; i++ {
		keypad.keys[i] = false
	}

	for key, chipKey := range keyMap {
		if ebiten.IsKeyPressed(key) {
			keypad.keys[chipKey] = true
		}
	}
}

func (keypad *Keypad) IsPressed(key uint8) bool {
	return keypad.keys[key]
}
