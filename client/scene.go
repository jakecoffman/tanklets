package client

import (
	"github.com/golang-ui/nuklear/nk"
	"github.com/go-gl/glfw/v3.2/glfw"
)

type Scene interface {
	Update(dt float64)
	Render(ctx *nk.Context)
	Transition(*glfw.Window) Scene
	Destroy()
}
