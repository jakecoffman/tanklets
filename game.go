package tanklets

import (
	"time"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
	"fmt"
)

type Game struct {
	state         int
	Keys          [1024]bool
	Width, Height int

	renderer *SpriteRenderer
	text     *TextRenderer
	simple   *SimpleRenderer

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
	ResourceManager.LoadShader("shaders/simple.vs.glsl", "shaders/simple.fs.glsl", "simple")
	ResourceManager.LoadShader("shaders/text.vs.glsl", "shaders/text.fs.glsl", "text")

	// renderers
	projection := mgl32.Ortho2D(0, width, height, 0)
	g.text = NewTextRenderer(ResourceManager.Shader("text"), width, height, "textures/Roboto-Light.ttf")
	g.text.SetColor(.8, .8, .3, 1)
	g.simple = NewSimpleRenderer(ResourceManager.Shader("simple"), projection)
	g.renderer = NewSpriteRenderer(ResourceManager.Shader("sprite"), projection)

	// textures
	tankTexture, err := ResourceManager.LoadTexture("textures/tank.png", "tank")
	if err != nil {
		panic(err)
	}
	turretTexture, err := ResourceManager.LoadTexture("textures/turret.png", "turret")
	if err != nil {
		panic(err)
	}
	bulletTexture, err := ResourceManager.LoadTexture("textures/bullet.png", "bullet")
	if err != nil {
		panic(err)
	}

	// physics
	g.space = cp.NewSpace()

	sides := []cp.Vector{
		{0, 0}, {0, height},
		{width, 0}, {width, height},
		{0, 0}, {width, 0},
		{0, height}, {width, height},
	}

	for i := 0; i < len(sides); i += 2 {
		var seg *cp.Shape
		seg = g.space.AddShape(cp.NewSegment(g.space.StaticBody, sides[i], sides[i+1], 0))
		seg.SetElasticity(1)
		seg.SetFriction(1)
		//seg.SetFilter(examples.NotGrabbableFilter)
	}
	tank1 = NewTank(g.space, tankTexture, turretTexture, bulletTexture, 20, 30)
	tank1.color = mgl32.Vec3{.4, .2, .8}
	tank1.Body.SetPosition(cp.Vector{100, 100})

	g.state = GAME_ACTIVE
}

func (g *Game) ProcessInput(dt float64) {
	if Tanklets.Keys[glfw.KeyD] {
		tank1.ControlBody.SetAngle(tank1.Body.Angle() + turnSpeed)
		// by applying to the body too, it will allow getting unstuck from corners
		tank1.Body.SetAngle(tank1.Body.Angle() + turnSpeed)
	}
	if Tanklets.Keys[glfw.KeyA] {
		tank1.ControlBody.SetAngle(tank1.Body.Angle() - turnSpeed)
		// by applying to the body too, it will allow getting unstuck from corners
		tank1.Body.SetAngle(tank1.Body.Angle() - turnSpeed)
	}
	if Tanklets.Keys[glfw.KeyW] {
		tank1.ControlBody.SetVelocityVector(tank1.Body.Rotation().Rotate(cp.Vector{Y: -maxSpeed}))
	} else if Tanklets.Keys[glfw.KeyS] {
		tank1.ControlBody.SetVelocityVector(tank1.Body.Rotation().Rotate(cp.Vector{Y: maxSpeed}))
	} else {
		tank1.ControlBody.SetVelocity(0, 0)
	}

	if LeftClick && time.Now().Sub(tank1.LastShot) > shotCooldown {
		tank1.Shoot(g.space)
		tank1.LastShot = time.Now()
	}
}

func (g *Game) Update(dt float64) {
	tank1.Update()

	for i := 0; i < len(Bullets); {
		now := time.Now()
		if now.Sub(Bullets[i].firedAt) > bulletTTL {
			Bullets = append(Bullets[:i], Bullets[i+1:]...)
		} else {
			break
		}
	}

	g.space.Step(dt)
	RightDown = false
	LeftDown = false
}

func (g *Game) Render() {
	tank1.Draw(g.renderer)
	for _, bullet := range Bullets {
		bullet.Draw(g.renderer)
	}
	g.text.Print(fmt.Sprint(tank1.Position()), 20, 30, 1)

	//g.simple.Draw(float32(pos.X), float32(pos.Y), float32(tank1.width), float32(tank1.height), float32(tank1.Angle()), 0, 0, 0, .5)
}

func (g *Game) Pause() {
	g.state = GAME_MENU
}

func (g *Game) Unpause() {
	g.state = GAME_ACTIVE
}
