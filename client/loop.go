package client

import (
	"log"
	"os"
	"runtime"
	"strconv"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

var (
	screenWidth  = 800
	screenHeight = 600

	fps = 0
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
	window.SetKeyCallback(KeyCallback)
	window.SetFramebufferSizeCallback(FramebufferSizeCallback)
	window.SetSizeCallback(windowSizeCallback)
	window.SetCursorPosCallback(CursorCallback)
	glfw.SwapInterval(1)

	if err := gl.Init(); err != nil {
		panic(err)
	}

	InitResources()

	dt := 0.
	lastFrame := 0.
	lastFps := 0.
	frames := 0

	ctx, font := GuiInit(window)
	defer GuiDestroy()

	CurrentScene = NewMainMenuScene(window, ctx)
	var currentFrame float64

	for !window.ShouldClose() {
		glfw.PollEvents()

		currentFrame = glfw.GetTime()
		dt = currentFrame - lastFrame
		lastFrame = currentFrame
		frames++
		if currentFrame - lastFps >= 1 {
			fps = frames
			frames = 0
			lastFps += 1
		}

		CurrentScene.Update(dt)
		CurrentScene.Render()

		window.SwapBuffers()
	}

	CurrentScene.Destroy()
	ResourceManager.Clear()
	runtime.KeepAlive(font)
}

func FramebufferSizeCallback(w *glfw.Window, width int, height int) {
	gl.Viewport(0, 0, int32(width), int32(height))
}

func windowSizeCallback(w *glfw.Window, width int, height int) {
	screenWidth = width
	screenHeight = height
}