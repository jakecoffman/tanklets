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
	width  = 640
	height = 480

	hwidth  = width / 2
	hheight = height / 2
)

var (
	tank1 *Tank
)

var Tanklets = NewGame(width, height)

func Main() {
	space := cp.NewSpace()

	sides := []cp.Vector{
		{-hwidth, -hheight}, {-hwidth, hheight},
		{hwidth, -hheight}, {hwidth, hheight},
		{-hwidth, -hheight}, {hwidth, -hheight},
		{-hwidth, hheight}, {hwidth, hheight},
	}

	for i := 0; i < len(sides); i += 2 {
		var seg *cp.Shape
		seg = space.AddShape(cp.NewSegment(space.StaticBody, sides[i], sides[i+1], 0))
		seg.SetElasticity(1)
		seg.SetFriction(0)
		//seg.SetFilter(core.NotGrabbableFilter)
	}

	tank1 = NewTank(space)
	tank1.Body.SetPosition(cp.Vector{100, 100})

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
	window, err := glfw.CreateWindow(width, height, "Breakout", nil, nil)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()
	window.MakeContextCurrent()
	window.SetKeyCallback(keyCallback)

	if err := gl.Init(); err != nil {
		panic(err)
	}

	gl.Enable(gl.CULL_FACE)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	Tanklets.Init()

	deltaTime := 0.5
	lastFrame := 0.0

	frames := 0
	showFps := time.Tick(1 * time.Second)

	for !window.ShouldClose() {
		currentFrame := glfw.GetTime()
		frames++
		select {
		case <-showFps:
			window.SetTitle(fmt.Sprintf("Breakout | %d FPS", frames))
			frames = 0
		default:
		}
		deltaTime = currentFrame - lastFrame
		lastFrame = currentFrame
		glfw.PollEvents()

		Tanklets.ProcessInput(deltaTime)
		Tanklets.Update(float32(deltaTime), space)

		gl.ClearColor(.2, .2, .8, 1)
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
