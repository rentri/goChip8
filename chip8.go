package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"
)

// chip 8 programs start at location 0x200 (512)
const (
	startAddress     = 0x200
	ScreenWidth      = 64 // x ScreenWidth
	ScreenHeight     = 32 // y ScreenHeight
	fontStartAddress = 0x50
	fontEndAddress   = 0x9F
)

type Chip8 struct {
	memory [4096]byte                       // 4kb memory
	V      [16]byte                         // registers V0 to VF , VF is flag register and should not be used by any program
	I      uint16                           // 16 bit memory address register
	PC     uint16                           // 16 bit pseudo register "program counter"
	inst   uint16                           // each instruction is 2 bytes long
	stack  [16]uint16                       // array of 16, 16bit values, upto 16 levels of nested subroutines allowed
	SP     uint8                            // points to topmost level of stack
	gfx    [ScreenWidth * ScreenHeight]byte // 64x32 display
	DT     byte                             // delay timer register
	ST     byte                             // sound timer register
	keypad *Keypad                          // keypad has 16 keys
	rng    *rand.Rand                       // rng for each instance of Chip8 used for CXNN instruction
}

var chip8Font = [80]byte{
	0xF0, 0x90, 0x90, 0x90, 0xF0, // 0
	0x20, 0x60, 0x20, 0x20, 0x70, // 1
	0xF0, 0x10, 0xF0, 0x80, 0xF0, // 2
	0xF0, 0x10, 0xF0, 0x10, 0xF0, // 3
	0x90, 0x90, 0xF0, 0x10, 0x10, // 4
	0xF0, 0x80, 0xF0, 0x10, 0xF0, // 5
	0xF0, 0x80, 0xF0, 0x90, 0xF0, // 6
	0xF0, 0x10, 0x20, 0x40, 0x40, // 7
	0xF0, 0x90, 0xF0, 0x90, 0xF0, // 8
	0xF0, 0x90, 0xF0, 0x10, 0xF0, // 9
	0xF0, 0x90, 0xF0, 0x90, 0x90, // A
	0xE0, 0x90, 0xE0, 0x90, 0xE0, // B
	0xF0, 0x80, 0x80, 0x80, 0xF0, // C
	0xE0, 0x90, 0x90, 0x90, 0xE0, // D
	0xF0, 0x80, 0xF0, 0x80, 0xF0, // E
	0xF0, 0x80, 0xF0, 0x80, 0x80, // F
}

func NewChip(keypad *Keypad) (chip *Chip8) {
	chip = &Chip8{
		keypad: keypad,
	}
	chip.PC = startAddress

	// seed rng with current time , seed is time elapsed since unix posix time
	chip.rng = rand.New(rand.NewSource(time.Now().UnixNano()))

	// load fonts
	copy(chip.memory[fontStartAddress:fontEndAddress], chip8Font[:])
	return
}

// loads the rom into memory 0x200 to 0xFFF
// returns error
func (chip *Chip8) LoadRom(rom string) error {
	data, err := os.ReadFile(rom)
	if err != nil {
		return err
	}

	nbytes := len(data)

	if nbytes <= 0 {
		return fmt.Errorf("Empty ROM %v bytes\n", nbytes)
	} else if nbytes > 4096-startAddress {
		return fmt.Errorf("ROM size %v bytes exceeds maximum allowed %v bytes\n", nbytes, (4096 - startAddress))
	} else {
		// 0x200 is the start of chip8 programs
		copy(chip.memory[startAddress:], data)
	}

	return nil
}

func (chip *Chip8) Cycle() {
	chip.fetch()
	chip.decodeAndExecute()
}

// read and copy two bytes from memory and store as instruction
// immediately increment PC by 2 bytes
func (chip *Chip8) fetch() {
	chip.inst = uint16(chip.memory[chip.PC])<<8 | uint16(chip.memory[chip.PC+1])
	chip.PC += 0x2
}

func (chip *Chip8) decodeAndExecute() {
	// DECODE

	// 0xAXYN instruction format
	// A,X,Y,N are 4 nibbles making up the 16 bits
	// X: second nibble, used to lookup one of the registers VX from V0 to VF
	// Y: third nibble, used to lookup one of the registers VY from V0 to VF
	// N: fourth nibble, a 4 bit number
	// NN: second byte, third and fourth nibbles, an immediate number
	// NNN: second, third and fourth nibble, a 12bit immediate memory address

	inst := chip.inst

	// extract second nibble
	X := (inst & 0x0F00) >> 8
	// extract third nibble
	Y := (inst & 0x00F0) >> 4
	// extract foruth nibble
	N := inst & 0x000F

	// extract second byte
	NN := inst & 0x00FF
	// extract second nibble and byte
	NNN := inst & 0x0FFF

	// EXECUTE
	switch inst & 0xF000 {

	case 0x0000: // 0x00E0, 0x00EE

		switch NN {
		case 0xE0:
			// CLS
			chip.cls()
		case 0xEE:
			// RET
			chip.ret()
		}

	case 0x1000: // 1NNN
		// JP addr
		chip.jp(NNN)

	case 0x2000: // 2NNN
		// CALL addr
		chip.call(NNN)

	case 0x3000: // 3XNN
		// SE VX, byte
		chip.skipIfEqual(X, NN)

	case 0x4000: // 4XNN
		// SNE VX, byte
		chip.skipIfNotEqual(X, NN)

	case 0x5000: // 5XY0
		// SE VX, VY
		chip.skipIfEqualReg(X, Y)

	case 0x6000: // 6XNN
		// LD VX, byte
		chip.ldByte(X, NN)

	case 0x7000: // 7XNN
		// ADD VX, byte
		chip.addByte(X, NN)

	case 0x8000: // 8XY0, 8XY1, 8XY2, 8XY3, 8XY4,
		// 8XY5, 8XY6, 8XY7, 8XYE

		switch N {
		case 0x0:
			// LD VX, VY
			chip.ldReg(X, Y)
		case 0x1:
			// OR VX, VY
			chip.orReg(X, Y)
		case 0x2:
			// AND VX, VY
			chip.andReg(X, Y)
		case 0x3:
			// XOR VX, VY
			chip.xorReg(X, Y)
		case 0x4:
			// ADD VX, VY
			chip.addReg(X, Y)
		case 0x5:
			// SUB VX, VY
			chip.subReg(X, Y)
		case 0x6:
			// SHR VX
			chip.shiftRight(X)
		case 0x7:
			// SUBN VX, VY
			chip.subNReg(X, Y)
		case 0xE:
			// SHL VX
			chip.shiftLeftReg(X)
		}

	case 0x9000: // 9XY0
		// SNE VX, VY
		chip.skipIfNotEqualReg(X, Y)

	case 0xA000: // ANNN
		// LD I, addr
		chip.ldAddr(NNN)

	case 0xB000: // BNNN
		// JP V0, addr
		chip.jpWithOffset(NNN)

	case 0xC000: // CXNN
		// RND VX, byte
		chip.randAndByte(X, NN)

	case 0xD000: // DXYN
		// DRW VX, VY, nibble
		chip.drw(X, Y, N)

	case 0xE000: // EX9E, EXA1

		switch NN {
		case 0x9E:
			// SKP VX
			chip.skipIfKeyPressed(X)
		case 0xA1:
			// SKNP VX
			chip.skipIfKeyNotPressed(X)
		}

	case 0xF000: // FX

		switch NN {

		case 0x07: // FX07
			// LD VX, DT
			chip.storeDelayTime(X)

		case 0x0A: // FX0A
			// LD VX, K
			chip.waitKeyPress(X)

		case 0x15: // FX15
			// LD DT, VX
			chip.setDelayTime(X)

		case 0x18: // FX18
			// LD ST, VX
			chip.setSoundTimer(X)

		case 0x1E: // FX1E
			// ADD I, VX
			chip.addToIndex(X)

		case 0x29: // FX29
			// LD F, VX
			chip.loadFontToIndex(X)

		case 0x33: // FX33
			// LD B, VX
			chip.storeBCD(X)

		case 0x55: // FX55
			// LD [I], VX
			chip.storeAllRegToMem(X)

		case 0x65: // FX65
			// LD VX, [I]
			chip.loadRegFromMem(X)
		}
	}
}

// CLS clear the display
// zero gfx array
func (chip *Chip8) cls() {
	for i := range chip.gfx {
		chip.gfx[i] = 0
	}
}

// RET return from a subroutine
// sets PC to address at top of stack, subtracts 1 from SP
func (chip *Chip8) ret() {
	chip.PC = chip.stack[chip.SP]
	chip.SP--
}

// JP addr
// sets PC to NNN
func (chip *Chip8) jp(NNN uint16) {
	chip.PC = NNN
}

// CALL addr
// call subroutine at nnn
func (chip *Chip8) call(NNN uint16) {
	chip.SP++
	chip.stack[chip.SP] = chip.PC
	chip.PC = NNN
}

// SE VX, byte
// if VX = NN then increment PC
func (chip *Chip8) skipIfEqual(X, NN uint16) {
	if chip.V[X] == byte(NN) {
		chip.PC += 2
	}
}

// SNE VX, byte
// if VX != NN then increment PC
func (chip *Chip8) skipIfNotEqual(X, NN uint16) {
	if chip.V[X] != byte(NN) {
		chip.PC += 2
	}
}

// SE Vx, Vy
// if Vx = Vy then increment PC
func (chip *Chip8) skipIfEqualReg(X, Y uint16) {
	if chip.V[X] == chip.V[Y] {
		chip.PC += 2
	}
}

// LD VX, byte
// set VX = byte
func (chip *Chip8) ldByte(X, NN uint16) {
	chip.V[X] = byte(NN)
}

// ADD VX, byte
// set VX = VX + byte
func (chip *Chip8) addByte(X, NN uint16) {
	chip.V[X] += byte(NN)
}

// LD VX, VY
// set VX = VY
func (chip *Chip8) ldReg(X, Y uint16) {
	chip.V[X] = chip.V[Y]
}

// OR VX, VY
// set VX = VX OR VY
func (chip *Chip8) orReg(X, Y uint16) {
	chip.V[X] |= chip.V[Y]
}

// AND VX, VY
// set VX = VX AND VY
func (chip *Chip8) andReg(X, Y uint16) {
	chip.V[X] &= chip.V[Y]
}

// XOR VX, VY
// set VX = VX XOR VY
func (chip *Chip8) xorReg(X, Y uint16) {
	chip.V[X] ^= chip.V[Y]
}

// ADD VX, VY
// set VX = VX + VY, set VF = carry
func (chip *Chip8) addReg(X, Y uint16) {
	sum := uint16(chip.V[X]) + uint16(chip.V[Y])

	// flags test expects setting VX first, then VF
	chip.V[X] = uint8(sum)

	if sum > 255 {
		chip.V[0xF] = 1 // carry
	} else {
		chip.V[0xF] = 0
	}
}

// SUB VX, VY
// set VX = VX - VY, set VF = !borrow
func (chip *Chip8) subReg(X, Y uint16) {

	// we are using the byte type which is uint8 under the hood
	// uint8 IS expected to wrap around if VX - VY goes negative
	vx := chip.V[X]
	vy := chip.V[Y]

	// VX is expected to be set before VF
	chip.V[X] -= chip.V[Y]

	if vx >= vy {
		chip.V[0xF] = 1
	} else {
		chip.V[0xF] = 0 // borrow
	}
}

// SUBN VX, VY
// set VX = VY - VX
func (chip *Chip8) subNReg(X, Y uint16) {

	vy := chip.V[Y]
	vx := chip.V[X]

	chip.V[X] = vy - vx

	if vy >= vx {
		chip.V[0xF] = 1
	} else {
		chip.V[0xF] = 0
	}
}

// In the COSMIC VIP, the SHR and SHL instruction put value of VY
// in VX then shifted VX by 1 bit
// however starting with CHIP-48 and SUPER-CHIP these instruction
// shifted VX and ignored VY completely
// cowgod's CHIP-8 reference uses the modern behavior

// NOTE: Ambiguous instruction
// SHR VX
// division by 2
func (chip *Chip8) shiftRight(X uint16) {
	// set VF to LSB of VX
	chip.V[0xF] = chip.V[X] & 0x1

	// division by 2
	chip.V[X] >>= 1
}

// NOTE: Ambiguous instruction
// SHL VX
// multiplication by 2
func (chip *Chip8) shiftLeftReg(X uint16) {
	// set VF to MSB of VX
	chip.V[0xF] = (chip.V[X] & 0x80) >> 7

	// multiplication by 2
	chip.V[X] <<= 1
}

// SNE VX, VY
// if VX != VY then increment PC
func (chip *Chip8) skipIfNotEqualReg(X, Y uint16) {
	if chip.V[X] != chip.V[Y] {
		chip.PC += 2
	}
}

// LD I, addr
// set I = NNN
func (chip *Chip8) ldAddr(NNN uint16) {
	chip.I = NNN
}

// JP V0, addr
// set PC = NNN + V0
func (chip *Chip8) jpWithOffset(NNN uint16) {
	chip.PC = uint16(chip.V[0]) + uint16(NNN)
}

// helper function for mehtod randAndByte
func (chip *Chip8) randByte() byte {
	// could not find what type of PRNG the early chip8 systems used
	// each system such as the COSMIC VIP and the HP-48 calculators
	// used their own PRNG implementation
	// return a pseudo random number <= 255
	return byte(chip.rng.Intn(256))
}

// RND VX, byte
// set VX = random byte & NN
func (chip *Chip8) randAndByte(X, NN uint16) {
	chip.V[X] = chip.randByte() & byte(NN)
}

// DRW VX, VY, nibble
// display nbyte sprite starting at memory location I at (VX, VY)
// set VF = collision (1 or 0)
func (chip *Chip8) drw(X, Y, N uint16) {
	// coordinates are modulo the size of display
	// however the actual drawing of the sprite should not wrap
	x := int(chip.V[X] & 63) // set x to VX modulo 64
	y := int(chip.V[Y] & 31) // set y to VY modulo 31

	chip.V[0xF] = 0

	for i := uint16(0); i < N; i++ {
		spriteNbyte := chip.memory[chip.I+i]

		// 8 bit sprite
		for j := 0; j < 8; j++ {
			// bit extraction
			bit := (spriteNbyte >> (7 - j)) & 1

			if bit == 1 {
				screenX := int(x) + j
				screenY := int(y) + int(i)

				// reached edge of screen, stop
				if screenX >= 64 || screenY >= 32 {
					continue // do not warp pixels
				}

				index := screenY*64 + screenX

				// set VF to 1, pixel is being turned off
				if chip.gfx[index] == 1 {
					chip.V[0xF] = 1
				}

				// xor sprite into the screen
				chip.gfx[index] ^= 1
			}
		}
	}
}

// SKP VX
// if key with value VX is pressed then increment pc
func (chip *Chip8) skipIfKeyPressed(X uint16) {
	if chip.keypad.IsPressed(chip.V[X]) {
		chip.PC += 2
	}
}

// SKNP VX
// if key with value VX is NOT pressed then increment PC
func (chip *Chip8) skipIfKeyNotPressed(X uint16) {
	if !chip.keypad.IsPressed(chip.V[X]) {
		chip.PC += 2
	}
}

// LD VX, DT
// set VX = DT
func (chip *Chip8) storeDelayTime(X uint16) {
	chip.V[X] = chip.DT
}

// LD VX, K
// wait for keypress, store value in VX
func (chip *Chip8) waitKeyPress(X uint16) {
	// look for any pressed key
	for i := 0; i < 16; i++ {
		if chip.keypad.IsPressed(uint8(i)) {
			chip.V[X] = byte(i)
			return
		}
	}

	// loop until key press
	chip.PC -= 2
}

// LD DT, VX
// set DT = VX
func (chip *Chip8) setDelayTime(X uint16) {
	chip.DT = chip.V[X]
}

// LD ST, VX
// set ST = VX
func (chip *Chip8) setSoundTimer(X uint16) {
	chip.ST = chip.V[X]
}

// ADD I, VX
// set I = I = VX
func (chip *Chip8) addToIndex(X uint16) {
	chip.I += uint16(chip.V[X])
}

// LD F, VX
// set I = location of sprite for digit VX
func (chip *Chip8) loadFontToIndex(X uint16) {
	// each font character is 5 byte long
	// we can get address of first byte of any character
	// by taking offset from start address (0x50)
	chip.I = fontStartAddress + (5 * uint16(chip.V[X]))
}

// LD B, VX
// store BCD representation of VX in memory locations, I, I+1, I+2
// hundredth digit at I,
// tens digit at I+1,
// ones digit at I+2
func (chip *Chip8) storeBCD(X uint16) {
	bcd := chip.V[X]

	chip.memory[chip.I+2] = bcd % 10
	bcd /= 10
	chip.memory[chip.I+1] = bcd % 10
	bcd /= 10
	chip.memory[chip.I] = bcd % 10
}

// LD [I], VX
// store V0 to VX in memory starting at location I
func (chip *Chip8) storeAllRegToMem(X uint16) {
	for i := uint16(0); i <= X; i++ {
		chip.memory[chip.I+i] = chip.V[i]
	}
}

// LD VX, [I]
// read register V0 to VX from memory starting at location I
func (chip *Chip8) loadRegFromMem(X uint16) {
	for i := uint16(0); i <= X; i++ {
		chip.V[i] = chip.memory[chip.I+i]
	}
}
