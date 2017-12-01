package client

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/tanklets"
)

var (
//Player *tanklets.Tank // pointer to the tank that represents the local player
)

var (
	Renderer      *SpriteRenderer
	Text          *TextRenderer
	Simple        *SimpleRenderer
	SpaceRenderer *CPRenderer
)

var (
	tankTexture   *Texture2D
	turretTexture *Texture2D
	bulletTexture *Texture2D
)

func Init(width, height float32) {
	// shaders
	ResourceManager.LoadShader("shaders/main.vs.glsl", "shaders/main.fs.glsl", "sprite")
	ResourceManager.LoadShader("shaders/simple.vs.glsl", "shaders/simple.fs.glsl", "simple")
	ResourceManager.LoadShader("shaders/text.vs.glsl", "shaders/text.fs.glsl", "text")
	ResourceManager.LoadShader("shaders/cp.vs.glsl", "shaders/cp.fs.glsl", "cp")

	// renderers
	projection = mgl32.Ortho2D(0, width, height, 0)
	Text = NewTextRenderer(ResourceManager.Shader("text"), width, height, "textures/Roboto-Light.ttf")
	Text.SetColor(.8, .8, .3, 1)
	Simple = NewSimpleRenderer(ResourceManager.Shader("simple"), projection)
	Renderer = NewSpriteRenderer(ResourceManager.Shader("sprite"), projection)
	SpaceRenderer = NewCPRenderer(ResourceManager.Shader("cp"), projection)

	// textures
	tankTexture = ResourceManager.LoadTexture("textures/tank.png", "tank")
	turretTexture = ResourceManager.LoadTexture("textures/turret.png", "turret")
	bulletTexture = ResourceManager.LoadTexture("textures/bullet.png", "bullet")
}

var projection mgl32.Mat4

func Render() {
	myTank := tanklets.Tanks[tanklets.Me]
	pos := myTank.Position()
	x, y := float32(pos.X), float32(pos.Y)
	projection = mgl32.Ortho2D(x-800./2., x+800./2., y+600./2., y-600./2.)
	Renderer.SetProjection(projection)

	// useful for debugging space issues
	SpaceRenderer.SetProjection(projection)
	SpaceRenderer.DrawSpace(tanklets.Space)

	//for _, tank := range tanklets.Tanks {
	//	DrawTank(tank)
	//}

	//for _, bullet := range tanklets.Bullets {
	//	DrawBullet(bullet)
	//}

	if tanklets.State == tanklets.GAME_WAITING {
		Text.Print("Connecting", 50, 100, 1)
	}
	if tanklets.State == tanklets.GAME_DEAD {
		Text.Print("You died", 50, 50, 1)
	}
}
