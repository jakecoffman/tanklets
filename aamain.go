package tanklets

import (
	"fmt"
	"runtime"
	"time"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

const (
	width  = 400
	height = 400

	hwidth  = width/2
	hheight = height/2
)

var (
	tank1 *Tank
)

var Tanklets = NewGame(width, height)

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
