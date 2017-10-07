package tanklets

import (
	"fmt"
	"runtime"
	"time"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/jakecoffman/cp"
)

const (
	width  = 400
	height = 400
)

var (
	tank1      Tank
	Mouse      cp.Vector
	LeftClick = false
	LeftDown = false
	RightClick = false
	RightDown  = false

	Tanklets = NewGame(width, height)
)

func Main() {
	runtime.LockOSThread()

	// glfw: initialize and configure
	glfw.Init()
	defer glfw.Terminate()
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)

	if runtime.GOOS == "darwin" {
		glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	}

	// glfw window creation
	window, err := glfw.CreateWindow(width, height, "Tanklets", nil, nil)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()
	window.MakeContextCurrent()
	window.SetKeyCallback(keyCallback)
	window.SetFramebufferSizeCallback(framebufferSizeCallback)
	window.SetCursorPosCallback(cursorCallback)
	window.SetMouseButtonCallback(mouseButtonCallback)

	if err := gl.Init(); err != nil {
		panic(err)
	}

	gl.Enable(gl.CULL_FACE)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	Tanklets.Init()

	deltaTime := 0.
	lastFrame := 0.

	frames := 0
	showFps := time.Tick(1 * time.Second)

	for !window.ShouldClose() {
		currentFrame := glfw.GetTime()
		frames++
		select {
		case <-showFps:
			window.SetTitle(fmt.Sprintf("Tanklets | %d FPS", frames))
			frames = 0
		default:
		}
		deltaTime = currentFrame - lastFrame
		lastFrame = currentFrame
		glfw.PollEvents()

		Tanklets.ProcessInput(deltaTime)
		Tanklets.Update(deltaTime)

		gl.ClearColor(.1, .1, .1, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		Tanklets.Render()

		window.SwapBuffers()
	}

	ResourceManager.Clear()
}

func keyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeyEscape && action == glfw.Press {
		window.SetShouldClose(true)
	}
	if key >= 0 && key < 1024 {
		if action == glfw.Press {
			Tanklets.Keys[key] = true
		} else if action == glfw.Release {
			Tanklets.Keys[key] = false
		}
	}
}

func framebufferSizeCallback(w *glfw.Window, width int, height int) {
	gl.Viewport(0, 0, int32(width), int32(height))
}

func cursorCallback(w *glfw.Window, xpos float64, ypos float64) {
	Mouse = cp.Vector{xpos, ypos}
}

func mouseButtonCallback(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
	if button == glfw.MouseButton1 {
		LeftDown = action == glfw.Press
		LeftClick = LeftDown
		//if action == glfw.Press {
//			// give the mouse click a little radius to make it easier to click small shapes.
//			//radius := 5.0
//
//			//info := space.PointQueryNearest(Mouse, radius, GrabFilter)
//			//
//			//if info.Shape != nil && info.Shape.Body().Mass() < INFINITY {
//			//	var nearest Vector
//			//	if info.Distance > 0 {
//			//		nearest = info.Point
//			//	} else {
//			//		nearest = Mouse
//			//	}
//			//
//			//	body := info.Shape.Body()
//			//	mouseJoint = NewPivotJoint2(mouseBody, body, Vector{}, body.WorldToLocal(nearest))
//			//	mouseJoint.SetMaxForce(50000)
//			//	mouseJoint.SetErrorBias(math.Pow(1.0-0.15, 60.0))
//			//	space.AddConstraint(mouseJoint)
//			}
//		//} else if mouseJoint != nil {
//		//	space.RemoveConstraint(mouseJoint)
//		//	mouseJoint = nil
//		//}
	} else if button == glfw.MouseButton2 {
		RightDown = action == glfw.Press
		RightClick = RightDown
	}
}
