package ui

import (
	"image/color"
)

var waterFallLut = make([]color.Color, 0)

func init() {
	// From GQRX: https://github.com/csete/gqrx -> qtgui/plotter.cpp
	for i := 0; i < 256; i++ {
		if i < 20 {
			// level 0: black background
			waterFallLut = append(waterFallLut, color.NRGBA{A: 255})
		} else if (i >= 20) && (i < 70) {
			// level 1: black -> blue
			waterFallLut = append(waterFallLut, color.NRGBA{B: uint8(140 * (i - 20) / 50), A: 255})
		} else if (i >= 70) && (i < 100) {
			// level 2: blue -> light-blue / greenish
			waterFallLut = append(waterFallLut, color.NRGBA{R: uint8(60 * (i - 70) / 30), G: uint8(125 * (i - 70) / 30), B: uint8(115*(i-70)/30 + 140), A: 255})
		} else if (i >= 100) && (i < 150) {
			// level 3: light blue -> yellow
			waterFallLut = append(waterFallLut, color.NRGBA{R: uint8(195*(i-100)/50 + 60), G: uint8(130*(i-100)/50 + 125), B: uint8(255 - (255 * (i - 100) / 50)), A: 255})
		} else if (i >= 150) && (i < 250) {
			// level 4: yellow -> red
			waterFallLut = append(waterFallLut, color.NRGBA{R: 255, G: uint8(255 - 255*(i-150)/100), A: 255})
		} else {
			// level 5: red -> white
			waterFallLut = append(waterFallLut, color.NRGBA{R: 255, G: uint8(255 * (i - 250) / 5), B: uint8(255 * (i - 250) / 5), A: 255})
		}
	}
}
