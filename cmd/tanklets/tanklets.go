package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"strconv"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/jakecoffman/tanklets/client"
)

const (
	width  = 800
	height = 600
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	runtime.LockOSThread()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			pprof.Lookup("goroutine").WriteTo(os.Stdout, 1)
			os.Exit(1)
		}
	}()

	// glfw: initialize and configure
	glfw.Init()
	defer glfw.Terminate()
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 2)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	// glfw window creation
	window, err := glfw.CreateWindow(width, height, "Tanklets", nil, nil)
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
	window.SetCursorPosCallback(client.CursorCallback)
	window.SetMouseButtonCallback(client.MouseButtonCallback)
	glfw.SwapInterval(1)

	if err := gl.Init(); err != nil {
		panic(err)
	}

	client.Init(width, height)

	dt := 0.
	lastFrame := 0.
	startFrame := glfw.GetTime()
	frames := 0

	font := client.GuiInit(window)
	defer client.GuiDestroy()

	scene := client.Scene(client.NewMainMenuScene())

	for !window.ShouldClose() {
		glfw.PollEvents()

		currentFrame := glfw.GetTime()
		dt = currentFrame - lastFrame
		lastFrame = currentFrame

		scene.Update(dt)

		gl.ClearColor(.1, .1, .1, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		scene.Render()

		if newScene := scene.Transition(); newScene != nil {
			scene.Destroy()
			scene = newScene
		}

		window.SwapBuffers()

		frames++
		if frames > 100 {
			window.SetTitle(fmt.Sprintf("Tanklets | %d FPS", int(float64(frames)/(currentFrame-startFrame))))
			frames = 0
			startFrame = currentFrame
		}
	}

	client.ResourceManager.Clear()
	runtime.KeepAlive(font)
}

func keyCallback(window *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeyEscape && action == glfw.Press {
		window.SetShouldClose(true)
	}
	if key >= 0 && key < 1024 {
		if action == glfw.Press {
			client.Keys[key] = true
		} else if action == glfw.Release {
			client.Keys[key] = false
		}
	}
}

func framebufferSizeCallback(w *glfw.Window, width int, height int) {
	gl.Viewport(0, 0, int32(width), int32(height))
}
