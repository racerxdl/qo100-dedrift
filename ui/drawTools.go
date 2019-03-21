package ui

import (
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/golang-ui/nuklear/nk"
	"github.com/racerxdl/segdsp/tools"
	"image"
	"image/color"
	"unsafe"
)

type UIState struct {
	bgColor nk.Color
}

func combine(c1, c2 color.Color) color.Color {
	r, g, b, a := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()

	return color.RGBA{
		R: uint8((r + r2) >> 9), // div by 2 followed by ">> 8"  is ">> 9"
		G: uint8((g + g2) >> 9),
		B: uint8((b + b2) >> 9),
		A: uint8((a + a2) >> 9),
	}
}

func DrawLine(x0, y0, x1, y1 float32, color color.Color, img *image.RGBA) {
	// DDA
	_, _, _, a := color.RGBA()
	needsCombine := a != 255 && a != 0
	var dx = x1 - x0
	var dy = y1 - y0
	var steps float32
	if tools.Abs(dx) > tools.Abs(dy) {
		steps = tools.Abs(dx)
	} else {
		steps = tools.Abs(dy)
	}

	var xinc = dx / steps
	var yinc = dy / steps

	var x = x0
	var y = y0
	for i := 0; i < int(steps); i++ {
		if needsCombine {
			var p = img.At(int(x), int(y))
			img.Set(int(x), int(y), combine(p, color))
		} else {
			img.Set(int(x), int(y), color)
		}
		x = x + xinc
		y = y + yinc
	}
}

func rgbaTex(tex int32, rgba *image.RGBA) (nk.Image, int32) {
	var t uint32
	if tex == -1 {
		gl.GenTextures(1, &t)
	} else {
		t = uint32(tex)
	}
	gl.BindTexture(gl.TEXTURE_2D, t)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_NEAREST)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR_MIPMAP_NEAREST)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA8, int32(rgba.Bounds().Dx()), int32(rgba.Bounds().Dy()),
		0, gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&rgba.Pix[0]))
	gl.GenerateMipmap(gl.TEXTURE_2D)
	return nk.NkImageId(int32(t)), int32(t)
}
