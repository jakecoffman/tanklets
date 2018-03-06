package client

import (
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/tanklets"
)

var (
	Player     *tanklets.Tank
	Keys       = [1024]bool{}
	mouse      cp.Vector
	LeftDown   bool
	RightDown  bool
	LeftClick  bool
	RightClick bool
)

var identityMatrix = mgl32.Mat4{
	1, 0, 0, 0,
	0, 1, 0, 0,
	0, 0, 1, 0,
	0, 0, 0, 1,
}

func CursorCallback(w *glfw.Window, xpos float64, ypos float64) {
	mouse = cp.Vector{xpos, ypos}
}

func MouseButtonCallback(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
	if button == glfw.MouseButton1 {
		LeftDown = action == glfw.Press
		LeftClick = LeftDown
	} else if button == glfw.MouseButton2 {
		RightDown = action == glfw.Press
		RightClick = RightDown
	}
}

func KeyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeyEscape && action == glfw.Press {
		window.SetShouldClose(true)
	}
	if key >= 0 && key < 1024 {
		if action == glfw.Press {
			Keys[key] = true
		} else if action == glfw.Release {
			Keys[key] = false
		}
	}
}
