package system

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"image/png"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"code.google.com/p/go.image/bmp"
)

const numCyclesPerGame = 100

var games []string = []string{
	"ibm",
	// Not supported yet.
	// "zero",
	// "PONG",
}

func TestGames(t *testing.T) {
	for _, game := range games {
		sys := New()
		defer sys.Quit()
		if err := sys.Init(); err != nil {
			t.Fatal(err)
			return
		}

		rom, err := ioutil.ReadFile("../games/" + game)
		if err != nil {
			t.Fatal(err)
		}
		sys.LoadGame(rom)
		sys.runCycles(numCyclesPerGame)

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
	}
}

// Produce a screenshot in PNG format and read its contents.
func newScreenshotHash(sys *Sys, game string) (h string, err error) {
	d, err := ioutil.TempDir("", game)
	if err != nil {
		return
	}
	fmt.Println("Screenshots directory:", d)
	//defer os.RemoveAll(d)

	screenshot := filepath.Join(d, game)
	if err = sys.video.SaveBMP(screenshot); err != nil {
		return
	}
	f, err := os.Open(screenshot)
	if err != nil {
		return
	}
	defer f.Close()
	m, err := bmp.Decode(f)
	if err != nil {
		return
	}

	buf := new(bytes.Buffer)
	err = png.Encode(buf, m)
	if err != nil {
		return
	}
	// The buffer is enough to produce the hash, but it's nice to have the PNG
	// files around when I need to (re)populate the screenshots directory.
	err = ioutil.WriteFile(screenshot+".png", buf.Bytes(), 0600)
	if err != nil {
		// Not a fatal error.
		log.Println(err)
	}

	got := md5.New()
	if _, err = io.Copy(got, buf); err != nil {
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
	wanted := md5.New()
	f, err := os.Open(filepath.Join("..", "screenshots", game+".png"))
	if err != nil {
		return
	}
	defer f.Close()
	if _, err = io.Copy(wanted, f); err != nil {
		return
	}
	h = fmt.Sprintf("%x", wanted.Sum(nil))
	return
}
