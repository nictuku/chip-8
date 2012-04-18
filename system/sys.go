// CHIP-8 emulator.
//
// nictuku@gamerom.org
//
// References:
// http://devernay.free.fr/hacks/chip8/C8TECH10.HTM
// http://en.wikipedia.org/wiki/CHIP-8
// http://www.multigesture.net/articles/how-to-write-an-emulator-chip-8-interpreter/
package system

import (
	"fmt"
	"log"
	"time"
)

func New() *Sys {
	mem := make(memory, ramCapacity)

	// XXX Should this be after 0x50? Find a program that uses it, then try.
	copy(mem, fontset)

	return &Sys{
		V:     make([]byte, numRegisters),
		PC:    programAreaStart,
		I:     programAreaStart,
		SP:    0,
		mem:   mem,
		gfx:   make([]byte, screenWidth*screenHeight),
		video: new(video),
	}
}

type Sys struct {
	// Exported fields will appear in the tracer.
	V          []byte `V`
	PC         uint16 `PC`
	SP         byte   `SP`
	mem        []byte
	I          uint16 `I`
	DelayTimer byte   `delayTimer`
	SoundTimer byte   `soundTimer`

	// The screen buffer. I could use the SDL pixels
	// directly, but the additional
	// code complexity isn't worth for saving 64*32 bytes.
	gfx []byte

	video *video
}

func (s *Sys) String() string {
	return CpuTracer(s)
}

func (s *Sys) Init() error {
	return s.video.init()
}

func (s *Sys) Quit() {
	s.video.quit()
}

func (s *Sys) LoadGame(rom []byte) {
	if len(rom) == 0 {
		log.Fatal("Tried to load zero-length ROM.")
	}
	romArea := s.mem[programAreaStart:]
	copy(romArea, rom)
}

func (s *Sys) runCycles(c int) error {
	tick := time.Tick(time.Second / cpuFrequency) // 60hz.
	for i := 0; c < 0 || i < c; i++ {
		<-tick
		if err := s.stepCycle(); err != nil {
			return err
		}
	}
	return nil
}

func (s *Sys) Run() error {
	return s.runCycles(-1)
}

func (s *Sys) stepCycle() error {

	draw := false
	// pc points to next opcode.
	opcode := uint16(s.mem[s.PC])<<8 | uint16(s.mem[s.PC+1])
	log.Printf("opcode 0x%04x", opcode)

	switch opcode & 0xF000 {
	case 0x0000:
		if opcode&0xFF00 != 0x0000 {
			// 0NNN	Calls RCA 1802 program at address NNN.
			//  => Only used by the original computers that implemented CHIP-8.
			goto NOTIMPLEMENTED
		}

		switch opcode & 0x000F {
		case 0x0000:
			// 00E0	Clears the screen.
			s.gfx = make([]byte, screenWidth*screenHeight)
			draw = true
		default:
			// 00EE	Returns from a subroutine.
			goto NOTIMPLEMENTED
		}

	case 0x1000:
		// 1NNN	Jumps to address NNN.
		s.PC = opcode & 0x0FFF
		goto SKIPINC

	// 2NNN	Calls subroutine at NNN.
	// 3XNN	Skips the next instruction if VX equals NN.

	case 0x4000:
		if s.V[(opcode&0x0F00)>>8] != byte(opcode&0x00FF) {
			// Skip next.
			s.PC += 2
		}

	// 4XNN	Skips the next instruction if VX doesn't equal NN.
	// 5XY0	Skips the next instruction if VX equals VY.

	case 0x6000:
		// 6XNN	Sets VX to NN.
		vx := (opcode & 0x0F00) >> 8
		s.V[vx] = byte(opcode & 0x00FF)
	case 0x7000:
		// 7XNN	Adds NN to VX.
		vx := (opcode & 0x0F00) >> 8
		s.V[vx] += byte(opcode & 0x00FF)

	// 8XY0	Sets VX to the value of VY.
	// 8XY1	Sets VX to VX or VY.
	// 8XY2	Sets VX to VX and VY.

	case 0x8000:
		switch opcode & 0x000F {
		case 0x0003:
			// 8XY3	Sets VX to VX xor VY.
			vx := (opcode & 0x0F00) >> 8
			vy := (opcode & 0x00F0) >> 4
			s.V[vx] = byte(vx ^ vy)
		case 0x0004:
			// 8XY4	Adds VY to VX. VF is set to 1 when there's a carry, and to 0 when there isn't.
			vx := (opcode & 0x0F00) >> 8
			vy := (opcode & 0x00F0) >> 4
			var add uint16 = uint16(s.V[vx]) + uint16(s.V[vy])
			s.V[vx] = byte(add & 0xFF)
			s.V[0xF] = byte(add>>8) & 0x1
		default:
			goto NOTIMPLEMENTED
		}
	// 8XY5	VY is subtracted from VX. VF is set to 0 when there's a borrow, and 1 when there isn't.
	// 8XY6	Shifts VX right by one. VF is set to the value of the least significant bit of VX before the shift.[2]
	// 8XY7	Sets VX to VY minus VX. VF is set to 0 when there's a borrow, and 1 when there isn't.
	// 8XYE	Shifts VX left by one. VF is set to the value of the most significant bit of VX before the shift.[2]
	// 9XY0	Skips the next instruction if VX doesn't equal VY.

	case 0xA000:
		// ANNN	Sets I to the address NNN.
		s.I = opcode & 0x0FFF

	// BNNN	Jumps to the address NNN plus V0.
	// CXNN	Sets VX to a random number and NN.

	case 0xD000:
		// DXYN	Draws a sprite at coordinate (VX, VY) that has a width of 8
		// pixels and a height of N pixels. Each row of 8 pixels is read as
		// bit-coded (with the most significant bit of each byte displayed on
		// the left) starting from memory location I; I value doesn't change
		// after the execution of this instruction.

		// Based on the implementation from:
		// http://www.multigesture.net
		var pixel byte
		x := uint16(s.V[(opcode&0x0F00)>>8])
		y := uint16(s.V[(opcode&0x00F0)>>4])
		height := uint16(opcode & 0xF)
		s.V[0xF] = 0

		for yline := uint16(0); yline < height; yline++ {
			pixel = s.mem[s.I+yline]
			for xline := uint16(0); xline < 8; xline++ {
				if (pixel & (0x80 >> xline)) != 0 {
					offset := (x + xline + ((y + yline) * 64))
					if s.gfx[offset] == 1 {
						// VF is set to 1 if any screen pixels are flipped from
						// set to unset when the sprite is drawn, and to 0 if
						// that doesn't happen.
						s.V[0xF] = 1
					}
					s.gfx[offset] ^= 1
				}
			}
		}
		draw = true

		// EX9E Skips the next instruction if the key stored in VX is pressed.
		// EXA1	Skips the next instruction if the key stored in VX isn't pressed.
		// FX07	Sets VX to the value of the delay timer.
		// FX0A	A key press is awaited, and then stored in VX.
	case 0xF000:
		switch opcode & 0x00FF {
		case 0x0015:
			// FX15	Sets the delay timer to VX.
			s.DelayTimer = byte(opcode & 0x00FF)
		default:
			goto NOTIMPLEMENTED
		}

		// FX18	Sets the sound timer to VX.
		// FX1E	Adds VX to I.[3]
		// FX29	Sets I to the location of the sprite for the character in VX. Characters 0-F (in hexadecimal) are represented by a 4x5 font.
		// FX33	Stores the Binary-coded decimal representation of VX, with the most significant of three digits at the address in I, the middle digit at I plus 1, and the least significant digit at I plus 2.
		// FX55	Stores V0 to VX in memory starting at address I.[4]
		// FX65	Fills V0 to VX with values from memory starting at address I.[4

	default:
		goto NOTIMPLEMENTED
	}
	// Common case.
	s.PC += 2
SKIPINC:
	if s.DelayTimer > 0 {
		s.DelayTimer -= 1
	}
	if s.SoundTimer > 0 {
		if s.SoundTimer == 1 {
			log.Println("BEEP!")
		}
		s.SoundTimer -= 1
	}

	if draw {
		s.video.draw(s.gfx)
	}
	// TODO(nictuku): Show CPU tracer on video.
	fmt.Println(s.String())
	return nil
NOTIMPLEMENTED:
	return fmt.Errorf("opcode not implemented: %x", opcode)
}
