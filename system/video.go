package system

import (
	"errors"

	"github.com/0xe2-0x9a-0x9b/Go-SDL/sdl"
)

const (
	// Multiplies screen size by 10.
	pixelSize = 10
)

type video struct {
}

func (v *video) init() error {
	if sdl.Init(sdl.INIT_EVERYTHING) != 0 {
		return errors.New(sdl.GetError())
	}
	screen := sdl.SetVideoMode(64*pixelSize, 32*pixelSize, 8, 0)
	if screen == nil {
		return errors.New(sdl.GetError())
	}
	sdl.EnableUNICODE(1)
	sdl.WM_SetCaption("nictuku's CHIP-8 emulator", "") // no icon.
	screen.FillRect(nil, 0x302019)
	return nil
}

func (v *video) quit() {
	sdl.Quit()
}

func (v *video) draw(pixels []byte) {
	surface := sdl.GetVideoSurface()

	// zero-out the screen.
	// sdl.GetVideoSurface().FillRect(nil, 0x302019)

	for yline := int16(0); yline < screenHeight; yline++ {
		for xline := int16(0); xline < screenWidth; xline++ {
			r := &sdl.Rect{xline * pixelSize, yline * pixelSize, pixelSize, pixelSize}
			if pixels[xline+yline*64] == 0 {
				surface.FillRect(r, sdl.MapRGBA(surface.Format, 255, 255, 255, 128))
			} else {
				surface.FillRect(r, sdl.MapRGBA(surface.Format, 0, 0, 0, 128))
			}
		}
	}
	surface.Flip()
}

func (v *video) SaveBMP(filename string) error {
	surface := sdl.GetVideoSurface()
	if surface == nil {
		return errors.New("video.SaveBMP: surface is nil")
	}
	if ret := surface.SaveBMP(filename); ret != 0 {
		return errors.New("video.SaveBMP: returned a non-zero value")
	}
	return nil
}
