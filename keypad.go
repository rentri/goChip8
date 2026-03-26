package main

type Keypad struct {
	keys [16]uint8
}

func (keypad *Keypad) Press(key uint8) {
	keypad.keys[key] = 1
}

func (keypad *Keypad) Release(key uint8) {
	keypad.keys[key] = 0
}

func (keypad *Keypad) IsPressed(key uint8) bool {
	return keypad.keys[key] == 1
}
