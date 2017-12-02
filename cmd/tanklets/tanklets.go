package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"strconv"
	"time"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang-ui/nuklear/nk"
	"github.com/jakecoffman/tanklets"
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

	tanklets.NewGame(width, height)
	client.Init(width, height)

	tanklets.NetInit()

	fmt.Println("Sending JOIN command")
	tanklets.Send(tanklets.Join{}, tanklets.ServerAddr)
	defer func() {
		fmt.Println("Sending DISCONNECT")
		tanklets.Send(tanklets.Disconnect{}, tanklets.ServerAddr)
	}()

	deltaTime := 0.
	lastFrame := 0.

	frames := 0
	showFps := time.Tick(1 * time.Second)

	font := client.GuiInit(window)
	defer client.GuiDestroy()

	guiState := &client.State{
		BgColor: nk.NkRgba(28, 48, 62, 255),
	}

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

		tanklets.ProcessIncoming()
		client.ProcessInput(deltaTime)
		tanklets.Update(deltaTime)

		gl.ClearColor(.1, .1, .1, 1)
		gl.Clear(gl.COLOR_BUFFER_BIT)

		gl.Enable(gl.BLEND)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		client.Render()
		client.GuiRender(guiState)

		window.SwapBuffers()
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
