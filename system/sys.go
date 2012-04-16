package system

import (
	"errors"
	"log"
	"time"

	"github.com/0xe2-0x9a-0x9b/Go-SDL/sdl"
)

func New() *Sys {
	mem := make(memory, 4096)

	// Build fontset.
	copy(mem, fontset)

	return &Sys{
		v:   make([]byte, 16),
		pc:  0x200,
		i:   0x200,
		sp:  0,
		mem: mem,
		gfx: make([]byte, screenWidth*screenHeight),
	}
}

type Sys struct {
	v          []byte
	pc         uint16
	sp         byte
	mem        []byte
	i          uint16
	gfx        []byte
	delayTimer byte
	soundTimer byte
}

func (s *Sys) Init() error {
	if sdl.Init(sdl.INIT_EVERYTHING) != 0 {
		return errors.New(sdl.GetError())
	}

	screen := sdl.SetVideoMode(640, 320, 8, 0)
	if screen == nil {
		return errors.New(sdl.GetError())
	}

	sdl.EnableUNICODE(1)

	sdl.WM_SetCaption("nictuku's CHIP-8 emulator", "") // no icon.

	screen.FillRect(nil, 0x302019)

	return nil
}

func (s *Sys) LoadGame(rom []byte) {
	if len(rom) == 0 {
		log.Fatal("Tried to load zero-length ROM.")
	}
	for i := 0; i < len(rom); i += 2 {
		s.mem[0x200+i] = rom[i]
		s.mem[0x200+i+1] = rom[i+1]

	}
	log.Printf("rom loaded %04x", s.mem[0x202])
}

func (s *Sys) Run() error {
	for _ = range [100]bool{} {
		s.stepCycle()
	}
	log.Println("run done")
	return nil
}

func (s *Sys) stepCycle() {

	draw := false

	// pc points to next opcode.
	opcode := uint16(s.mem[s.pc])<<8 | uint16(s.mem[s.pc+1])
	log.Printf("opcode 0x%04x", opcode)
	log.Printf("pc 0x%04x", s.pc)

	masked := opcode & 0xF000
	log.Printf("%x", masked)
	switch masked {

	case 0x0000:
		switch masked & 0x000F {
		case 0x0000: // 00E0	Clears the screen.
			// XXXX
			log.Println("clearing screen")
			sdl.GetVideoSurface().FillRect(nil, 0x302019)
			s.pc += 2
			// 00EE	Returns from a subroutine.
		default:
			log.Printf("opcode not implemented: %x", opcode)
			return
		}
		//	default:
		//		// 0NNN	Calls RCA 1802 program at address NNN.
		//		s.pc = opcode

	// 1NNN	Jumps to address NNN.
	// 2NNN	Calls subroutine at NNN.
	// 3XNN	Skips the next instruction if VX equals NN.
	// 4XNN	Skips the next instruction if VX doesn't equal NN.
	// 5XY0	Skips the next instruction if VX equals VY.
	case 0x6000:
		// 6XNN	Sets VX to NN.
		vx := (opcode & 0x0F00) >> 8
		s.v[vx] = byte(opcode & 0x00FF)
		s.pc += 2
	case 0x7000:
		// 7XNN	Adds NN to VX.
		vx := (opcode & 0x0F00) >> 8
		s.v[vx] += byte(opcode & 0x00FF)
		// XXX increment VF.
		s.pc += 2

	// 8XY0	Sets VX to the value of VY.
	// 8XY1	Sets VX to VX or VY.
	// 8XY2	Sets VX to VX and VY.
	// 8XY3	Sets VX to VX xor VY.
	// 8XY4	Adds VY to VX. VF is set to 1 when there's a carry, and to 0 when there isn't.
	// 8XY5	VY is subtracted from VX. VF is set to 0 when there's a borrow, and 1 when there isn't.
	// 8XY6	Shifts VX right by one. VF is set to the value of the least significant bit of VX before the shift.[2]
	// 8XY7	Sets VX to VY minus VX. VF is set to 0 when there's a borrow, and 1 when there isn't.
	// 8XYE	Shifts VX left by one. VF is set to the value of the most significant bit of VX before the shift.[2]
	// 9XY0	Skips the next instruction if VX doesn't equal VY.

	case 0xA000: // ANNN	Sets I to the address NNN.
		s.i = opcode & 0x0FFF
		s.pc += 2 // BNNN	Jumps to the address NNN plus V0.
		// CXNN	Sets VX to a random number and NN.
		log.Println("ANNN")
	case 0xD000:
		// DXYN	Draws a sprite at coordinate (VX, VY) that has a width of 8
		// pixels and a height of N pixels. Each row of 8 pixels is read as
		// bit-coded (with the most significant bit of each byte displayed on
		// the left) starting from memory location I; I value doesn't change
		// after the execution of this instruction.
		var pixel byte
		x := uint16(s.v[(opcode&0x0F00)>>8])
		y := uint16(s.v[(opcode&0x00F0)>>4])
		height := uint16(opcode & 0xF)
		s.v[0xF] = 0
		log.Println("height", height)

		for yline := uint16(0); yline < height; yline++ {
			pixel = s.mem[s.i+yline]
			log.Printf("pixel: %x", pixel)
			for xline := uint16(0); xline < 8; xline++ {
				if (pixel & (0x80 >> xline)) != 0 {
					offset := (x + xline + ((y + yline) * 64))
					if s.gfx[offset] == 1 {
						// VF is set to 1 if any screen pixels are flipped from
						// set to unset when the sprite is drawn, and to 0 if
						// that doesn't happen.
						s.v[0xF] = 1
					}
					s.gfx[offset] ^= 1
					log.Printf("setting something in offset %x", offset)
				}
			}
		}
		draw = true
		s.pc += 2

		// EX9E Skips the next instruction if the key stored in VX is pressed.
		// EXA1	Skips the next instruction if the key stored in VX isn't pressed.
		// FX07	Sets VX to the value of the delay timer.
		// FX0A	A key press is awaited, and then stored in VX.
		// FX15	Sets the delay timer to VX.
		// FX18	Sets the sound timer to VX.
		// FX1E	Adds VX to I.[3]
		// FX29	Sets I to the location of the sprite for the character in VX. Characters 0-F (in hexadecimal) are represented by a 4x5 font.
		// FX33	Stores the Binary-coded decimal representation of VX, with the most significant of three digits at the address in I, the middle digit at I plus 1, and the least significant digit at I plus 2.
		// FX55	Stores V0 to VX in memory starting at address I.[4]
		// FX65	Fills V0 to VX with values from memory starting at address I.[4

	default:
		log.Printf("opcode not implemented: %x", opcode)
		return
	}
	if s.delayTimer > 0 {
		s.delayTimer -= 1
	}
	if s.soundTimer > 0 {
		if s.soundTimer == 1 {
			log.Println("BEEP!")
		}
		s.soundTimer -= 1
	}

	if draw {
		surface := sdl.GetVideoSurface()
		for yline := int16(0); yline < screenHeight; yline++ {
			for xline := int16(0); xline < screenWidth; xline++ {
				if s.gfx[xline+yline*64] == 0 {
					// unsafe.Pointer, do arithm to find what to update.
					surface.FillRect(&sdl.Rect{xline * 10, yline * 10, 10, 10},
						sdl.MapRGBA(surface.Format, 255, 255, 255, 128),
					)
				} else {
					surface.FillRect(&sdl.Rect{xline * 10, yline * 10, 10, 10},
						sdl.MapRGBA(surface.Format, 0, 0, 0, 128),
					)
				}
			}
		}
		surface.Flip()

	}

	time.Sleep(time.Second / 5) // 60hz.
	// time.Sleep(time.Second / cpuFrequency) // 60hz.
}
