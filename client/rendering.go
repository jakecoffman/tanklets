package client

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/tanklets"
)

var (
//Player *tanklets.Tank // pointer to the tank that represents the local player
)

var (
	Renderer *SpriteRenderer
	Text     *TextRenderer
	Simple   *SimpleRenderer
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

	// renderers
	projection := mgl32.Ortho2D(0, width, height, 0)
	Text = NewTextRenderer(ResourceManager.Shader("text"), width, height, "textures/Roboto-Light.ttf")
	Text.SetColor(.8, .8, .3, 1)
	Simple = NewSimpleRenderer(ResourceManager.Shader("simple"), projection)
	Renderer = NewSpriteRenderer(ResourceManager.Shader("sprite"), projection)

	// textures
	tankTexture = ResourceManager.LoadTexture("textures/tank.png", "tank")
	turretTexture = ResourceManager.LoadTexture("textures/turret.png", "turret")
	bulletTexture = ResourceManager.LoadTexture("textures/bullet.png", "bullet")
}

func Render() {
	for _, tank := range tanklets.Tanks {
		DrawTank(tank)
	}

	for _, bullet := range tanklets.Bullets {
		DrawBullet(bullet)
	}
	//g.text.Print(fmt.Sprint(g.Player.Position()), 20, 30, 1)

	//g.simple.Draw(float32(pos.X), float32(pos.Y), float32(tank1.width), float32(tank1.height), float32(tank1.Angle()), 0, 0, 0, .5)

	if tanklets.State == tanklets.GAME_WAITING {
		Text.Print("Connecting", 50, 100, 1)
	}
	if tanklets.State == tanklets.GAME_DEAD {
		Text.Print("You died", 50, 50, 1)
	}
}
