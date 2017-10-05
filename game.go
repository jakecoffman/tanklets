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
	ResourceManager.LoadShader("shaders/main.vs.glsl", "shaders/main.fs.glsl", "sprite")
	ResourceManager.LoadShader("shaders/particle.vs.glsl", "shaders/particle.fs.glsl", "particle")
	g.text = NewTextRenderer("shaders/text.vs.glsl", "shaders/text.fs.glsl", width, height)
	if err := g.text.Load("textures/Roboto-Light.ttf", 24); err != nil {
		panic(err)
	}
	g.text.SetColor(1, 1, 1, 1)

	projection := mgl32.Ortho(0, float32(g.Width), float32(g.Height), 0, -1, 1)
	ResourceManager.Shader("sprite").
		Use().
		SetInt("sprite", 0).
		SetMat4("projection", projection)
	g.renderer = NewSpriteRenderer(ResourceManager.Shader("sprite"))

	ResourceManager.LoadTexture("textures/tank.png", "tank")

	g.state = GAME_ACTIVE
}

func (g *Game) ProcessInput(dt float64) {
	if Tanklets.Keys[glfw.KeyD] {
		tank1.ControlBody.SetAngle(tank1.Body.Angle() - turnSpeed)
	}
	if Tanklets.Keys[glfw.KeyA] {
		tank1.ControlBody.SetAngle(tank1.Body.Angle() + turnSpeed)
	}
	if Tanklets.Keys[glfw.KeyW] {
		tank1.ControlBody.SetVelocityVector(tank1.Body.Rotation().Rotate(cp.Vector{maxSpeed, 0.0}))
	} else if Tanklets.Keys[glfw.KeyS] {
		tank1.ControlBody.SetVelocityVector(tank1.Body.Rotation().Rotate(cp.Vector{-maxSpeed, 0.0}))
	} else {
		tank1.ControlBody.SetVelocity(0, 0)
	}
}

func (g *Game) Update(dt float32, space *cp.Space) {
	space.Step(float64(dt))
}

func (g *Game) Render() {
	pos := tank1.Position()
	tank := ResourceManager.Texture("tank")

	g.renderer.DrawSprite(tank, mgl32.Vec2{float32(pos.X), float32(pos.Y)}, mgl32.Vec2{50, 50}, tank1.Angle(), DefaultColor)
	g.text.Print(fmt.Sprint("Tank: ", pos, tank1.Angle()), 10, 25, 1)
}

func (g *Game) Pause() {
	g.state = GAME_MENU
}

func (g *Game) Unpause() {
	g.state = GAME_ACTIVE
}
