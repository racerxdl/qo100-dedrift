package ui

import (
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang-ui/nuklear/nk"
)

type DrawableComponent interface {
	Draw(win *glfw.Window, ctx *nk.Context)
	Lock()
	Unlock()
}
