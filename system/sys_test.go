package system

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"

	"os"
	"path/filepath"
	"runtime"
	"testing"
)

const numCyclesPerGame = 120 // 2 seconds per test.

var games []string = []string{
	"ibm",
	"zero",

	// TODO(nictuku): These games should be tested with keyboard events too.
	"GUESS",
	"SYZYGY",

	// Inconsistent spaceship blinking?
	// "UFO",

	// Although pong works for 120 cycles, it crashes later, when the balls
	// leaves the screen: Missing instruction 80d5.
	// "PONG",

	// Missing SCHIP instructions.
	// "BLINKY",
}

func Test8XY0(t *testing.T) {
	s := New()
	s.V[0xd] = 1
	s.mem[s.PC] = 0x88
	s.mem[s.PC+1] = 0xd0
	if err := s.stepCycle(); err != nil {
		t.Fatal(err)
		return
	}
	if s.V[0x8] != byte(1) {
		t.Errorf("s.V[0x8], got %x, wanted %x", s.V[0x8], 1)
	}
}

func TestFX1E(t *testing.T) {
	s := New()
	s.V[3] = 2 // Add 2.

	// Overflow test.
	s.I = 65535
	s.mem[s.PC] = 0xF3
	s.mem[s.PC+1] = 0x1E
	if err := s.stepCycle(); err != nil {
		t.Fatal(err)
		return
	}
	if s.I != uint16(1) {
		t.Errorf("s.I, got %x, wanted %x", s.I, 1)
	}
	if s.V[0xf] != 0x1 {
		t.Errorf("s.V[0xf], got %x, wanted %x", s.V[0xf], 0x1)
	}
	// Again, but no overflow.
	s.mem[s.PC] = 0xF3
	s.mem[s.PC+1] = 0x1E
	if err := s.stepCycle(); err != nil {
		t.Fatal(err)
		return
	}
	if s.I != uint16(3) {
		t.Errorf("s.I, got %x, wanted %x", s.I, 3)
	}
	if s.V[0xf] != 0 {
		t.Errorf("s.V[0xf], got %x, wanted %x", s.V[0xf], 0)
	}
}

func TestFX33(t *testing.T) {
	s := New()
	s.V[3] = 34
	s.I = 1337
	s.mem[s.PC] = 0xF3
	s.mem[s.PC+1] = 0x33
	if err := s.stepCycle(); err != nil {
		t.Fatal(err)
		return
	}
	if s.mem[s.I] != byte(0) {
		t.Errorf("s.I, got %x, wanted %x", s.mem[s.I], 0)
	}
	if s.mem[s.I+1] != byte(3) {
		t.Errorf("s.I, got %x, wanted %x", s.mem[s.I+1], 3)
	}
	if s.mem[s.I+2] != byte(4) {
		t.Errorf("s.I, got %x, wanted %x", s.mem[s.I+2], 4)
	}
}

func TestGames(t *testing.T) {
	// Necessary for being able to test opengl initialization.
	runtime.LockOSThread()

	for _, game := range games {
		// Run a little function so defer works.
		func() {
			sys := New()

			if err := sys.Init(); err != nil {
				t.Fatal(err)
				return
			}
			defer sys.Close()

			rom, err := ioutil.ReadFile("../games/" + game)
			if err != nil {
				t.Fatal(err)
			}
			sys.LoadGame(rom)
			if err := sys.runCycles(numCyclesPerGame); err != nil {
				t.Fatal(err)
			}

			gh, err := newScreenshotHash(sys, game)
			if err != nil {
				t.Fatal(err)
			}
			wh, err := archivedScreenshotHash(game)
			if err != nil {
				t.Fatal(err)
			}

			if wh != gh {
				t.Fatalf("Game %v, test failed. Wanted: %v, got %v", game, wh, gh)
			}
		}()
	}
}

// Produce a screenshot in PNG format and read its contents.
func newScreenshotHash(sys *Sys, game string) (h string, err error) {
	d, err := ioutil.TempDir("", game)
	if err != nil {
		return
	}

	// It's nice to have the PNG files around when I need to (re)populate the screenshots directory,
	// so I don't delete them.

	screenshot := filepath.Join(d, game) + ".png"
	fmt.Println("Screenshot:", screenshot)

	if err = sys.video.SavePNG(screenshot); err != nil {
		return
	}
	f, err := os.Open(screenshot)
	if err != nil {
		return
	}
	defer f.Close()

	got := md5.New()
	if _, err = io.Copy(got, f); err != nil {
		return
	}

	h = fmt.Sprintf("%x", got.Sum(nil))
	return

}

// archivedScreenshotHash reads the png file in ../screenshots/game.png and
// returns its md5 hash. The test could simply use the hash itself, but I have
// to store the screenshots themselves for show-off and debugging anyway, so I
// might as well just read the files.
func archivedScreenshotHash(game string) (h string, err error) {
	f, err := os.Open(filepath.Join("..", "screenshots", game+".png"))
	if err != nil {
		return
	}
	defer f.Close()

	wanted := md5.New()
	if _, err = io.Copy(wanted, f); err != nil {
		return
	}
	h = fmt.Sprintf("%x", wanted.Sum(nil))
	return
}
