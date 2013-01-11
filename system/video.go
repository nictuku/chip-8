package system

import (
	"fmt"
	"github.com/disintegration/imaging"
	"github.com/go-gl/gl"
	"github.com/go-gl/glfw"
	"github.com/go-gl/glh"
	"image"
	"image/png"
	"os"
)

const (
	// Multiplies screen size by 10.
	pixelSize = 10
)

type video struct {
}

func (v *video) init() error {
	// Disclaimer: I'm an absolute noob in openGL and I just slammed the
	// keyboard with my text editor open, until it started to produce usable
	// results. I don't fully understand everything I'm doing here.
	if err := glfw.Init(); err != nil {
		return err
	}

	err := glfw.OpenWindow(screenWidth*pixelSize, screenHeight*pixelSize, 0, 0, 0, 0, 0, 0, glfw.Windowed)
	if err != nil {
		return err
	}

	// Enable vertical sync on cards that support it.
	glfw.SetSwapInterval(1)

	glfw.SetWindowTitle("nictuku's CHIP-8 emulator")

	gl.Init()
	if err = glh.CheckGLError(); err != nil {
		return err
	}

	gl.ClearColor(0, 0, 0, 0)
	gl.MatrixMode(gl.PROJECTION)

	// Change coordinates to range from [0, 64] and [0,32].
	gl.Ortho(0, screenWidth, screenHeight, 0, 0, 1)

	// Unnecessary sanity check. :-P
	if glfw.WindowParam(glfw.Opened) == 0 {
		return fmt.Errorf("No window opened")
	}

	return nil
}

func (v *video) quit() {
	gl.End()
	glfw.Terminate()
}

func (v *video) close() {
	glfw.CloseWindow()
}

func (v *video) draw(pixels []byte) {

	// No need to clear the screen since I explicitly redraw all pixels, at
	// least currently.

	gl.MatrixMode(gl.POLYGON)

	for yline := 0; yline < screenHeight; yline++ {

		for xline := 0; xline < screenWidth; xline++ {

			x, y := float32(xline), float32(yline)
			if pixels[xline+yline*64] == 0 {
				gl.Color3f(0, 0, 0)
			} else {
				gl.Color3f(1, 1, 1)
			}
			gl.Rectf(x, y, x+1, y+1)
		}
	}

	glfw.SwapBuffers()
}

func (v *video) SavePNG(filename string) error {
	// If I don't do this, it doesn't show the most recent screen state in the screenshot.
	glfw.SwapBuffers()

	return CaptureToPng(filename)
}

func CaptureToPng(filename string) error {
	// Based on a version from https://github.com/go-gl/glh.

	// Copyright (c) 2012 The go-gl Authors. All rights reserved.
	w, h := glh.GetViewportWH()

	im := image.NewNRGBA(image.Rect(0, 0, w, h))
	gl.ReadBuffer(gl.BACK_LEFT)
	gl.ReadPixels(0, 0, w, h, gl.RGBA, gl.UNSIGNED_BYTE, im.Pix)

	// Flip the image vertically.
	//
	// From IRC:
	// <ClaudiusMaximus> nictuku: glReadPixels uses (0,0) at bottom left always 
	// - some (most?) image formats use (0,0) as top left
	im = imaging.FlipV(im)

	fd, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer fd.Close()

	png.Encode(fd, im)
	return nil
}
