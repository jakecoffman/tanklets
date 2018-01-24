package client

import (
	"os"
	"strconv"
	"log"
	"fmt"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/gl/v3.2-core/gl"
	"runtime"
)

var (
	screenWidth  = 800
	screenHeight = 600
)

func Loop() {
	// glfw: initialize and configure
	glfw.Init()
	defer glfw.Terminate()
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 2)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	// glfw window creation
	window, err := glfw.CreateWindow(screenWidth, screenHeight, "Tanklets", nil, nil)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()
	if len(os.Args) > 1 {
		y, err := strconv.Atoi(os.Args[1])
		if err != nil {
			log.Println(err)
		} else {
			window.SetPos(0, y)
		}
	} else {
		window.SetPos(0, 0)
	}
	window.MakeContextCurrent()
	window.SetKeyCallback(keyCallback)
	window.SetFramebufferSizeCallback(framebufferSizeCallback)
	window.SetSizeCallback(windowSizeCallback)
	window.SetCursorPosCallback(CursorCallback)
	glfw.SwapInterval(1)

	if err := gl.Init(); err != nil {
		panic(err)
	}

	InitResources()

	dt := 0.
	lastFrame := 0.
	startFrame := glfw.GetTime()
	frames := 0

	ctx, font := GuiInit(window)
	defer GuiDestroy()

	scene := Scene(NewMainMenuScene())

	for !window.ShouldClose() {
		glfw.PollEvents()

		currentFrame := glfw.GetTime()
		dt = currentFrame - lastFrame
		lastFrame = currentFrame

		scene.Update(dt)
		scene.Render(ctx)

		window.SwapBuffers()

		if newScene := scene.Transition(window); newScene != nil {
			scene.Destroy()
			scene = newScene
		}

		frames++
		if frames > 100 {
			window.SetTitle(fmt.Sprintf("Tanklets | %d FPS", int(float64(frames)/(currentFrame-startFrame))))
			frames = 0
			startFrame = currentFrame
		}
	}

	scene.Destroy()
	ResourceManager.Clear()
	runtime.KeepAlive(font)
}

func keyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
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

func framebufferSizeCallback(w *glfw.Window, width int, height int) {
	gl.Viewport(0, 0, int32(width), int32(height))
}

func windowSizeCallback(w *glfw.Window, width int, height int) {
	screenWidth = width
	screenHeight = height
}