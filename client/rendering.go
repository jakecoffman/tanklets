package client

import (
	"github.com/go-gl/mathgl/mgl32"
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

var projection mgl32.Mat4

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
