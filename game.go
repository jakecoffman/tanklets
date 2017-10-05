package tanklets

import (
	"fmt"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
)

type Game struct {
	state         int
	Keys          [1024]bool
	Width, Height int

	renderer *SpriteRenderer
	text     *TextRenderer

	space *cp.Space
}

// Game state
const (
	GAME_ACTIVE = iota
	GAME_MENU
	GAME_WIN
)

func NewGame(width, height int) *Game {
	return &Game{
		Width:  width,
		Height: height,
		Keys:   [1024]bool{},
	}
}

func (g *Game) Init() {
	// shaders
	ResourceManager.LoadShader("shaders/main.vs.glsl", "shaders/main.fs.glsl", "sprite")
	g.text = NewTextRenderer("shaders/text.vs.glsl", "shaders/text.fs.glsl", width, height)
	if err := g.text.Load("textures/Roboto-Light.ttf", 24); err != nil {
		panic(err)
	}
	g.text.SetColor(.8, .8, .3, 1)

	projection := mgl32.Ortho2D(0, float32(width), float32(height), 0)
	ResourceManager.Shader("sprite").
		Use().
		SetInt("sprite", 0).
		SetMat4("projection", projection)
	g.renderer = NewSpriteRenderer(ResourceManager.Shader("sprite"))

	// textures
	ResourceManager.LoadTexture("textures/tank.png", "tank")

	// physics
	g.space = cp.NewSpace()

	sides := []cp.Vector{
		{0, 0},
		{0, height},
		{width, height},
		{width, 0},
		{0, 0},
	}

	for i := 1; i < len(sides); i++ {
		var seg *cp.Shape
		seg = g.space.AddShape(cp.NewSegment(g.space.StaticBody, sides[i-1], sides[i], 0))
		seg.SetElasticity(1)
		seg.SetFriction(0)
		//seg.SetFilter(core.NotGrabbableFilter)
	}
	tankTexture := ResourceManager.Texture("tank")
	tank1 = NewTank(g.space, tankTexture.Width, tankTexture.Height)
	tank1.Body.SetPosition(cp.Vector{100, 100})

	g.state = GAME_ACTIVE
}

func (g *Game) ProcessInput(dt float64) {
	if Tanklets.Keys[glfw.KeyD] {
		tank1.ControlBody.SetAngle(tank1.Body.Angle() + turnSpeed)
	}
	if Tanklets.Keys[glfw.KeyA] {
		tank1.ControlBody.SetAngle(tank1.Body.Angle() - turnSpeed)
	}
	if Tanklets.Keys[glfw.KeyW] {
		tank1.ControlBody.SetVelocityVector(tank1.Body.Rotation().Rotate(cp.Vector{Y: -maxSpeed}))
	} else if Tanklets.Keys[glfw.KeyS] {
		tank1.ControlBody.SetVelocityVector(tank1.Body.Rotation().Rotate(cp.Vector{Y: maxSpeed}))
	} else {
		tank1.ControlBody.SetVelocity(0, 0)
	}
}

func (g *Game) Update(dt float64) {
	g.space.Step(dt)
}

func (g *Game) Render() {
	pos := tank1.Position()
	tank := ResourceManager.Texture("tank")

	g.renderer.DrawSprite(tank, mgl32.Vec2{float32(pos.X), float32(pos.Y)}, mgl32.Vec2{float32(tank.Width), float32(tank.Height)}, tank1.Angle(), DefaultColor)
	g.text.Print(fmt.Sprint("Tanklets! ", tank.Width, tank.Height), 10, 25, 1)
}

func (g *Game) Pause() {
	g.state = GAME_MENU
}

func (g *Game) Unpause() {
	g.state = GAME_ACTIVE
}
