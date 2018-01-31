package client

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/tanklets/client/glpers"
)

var ResourceManager = glpers.NewResourceManager()

var (
	Renderer      *glpers.SpriteRenderer
	Text          *glpers.TextRenderer
	Simple        *glpers.SimpleRenderer
	SpaceRenderer *glpers.CPRenderer
)

var (
	tankTexture   *glpers.Texture2D
	turretTexture *glpers.Texture2D
	bulletTexture *glpers.Texture2D
)

var projection mgl32.Mat4

func InitResources() {
	// shaders
	ResourceManager.LoadShader("shaders/main.vs.glsl", "shaders/main.fs.glsl", "sprite")
	ResourceManager.LoadShader("shaders/simple.vs.glsl", "shaders/simple.fs.glsl", "simple")
	ResourceManager.LoadShader("shaders/text.vs.glsl", "shaders/text.fs.glsl", "text")
	ResourceManager.LoadShader("shaders/cp.vs.glsl", "shaders/cp.fs.glsl", "cp")

	// renderers
	projection = mgl32.Ortho2D(0, float32(screenWidth), float32(screenHeight), 0)
	Text = glpers.NewTextRenderer(ResourceManager.Shader("text"), float32(screenWidth), float32(screenHeight), "textures/Roboto-Light.ttf")
	Text.SetColor(.8, .8, .3, 1)
	Simple = glpers.NewSimpleRenderer(ResourceManager.Shader("simple"), projection)
	Renderer = glpers.NewSpriteRenderer(ResourceManager.Shader("sprite"), projection)
	SpaceRenderer = glpers.NewCPRenderer(ResourceManager.Shader("cp"), projection)

	// textures
	tankTexture = ResourceManager.LoadTexture("textures/tank.png", "tank")
	turretTexture = ResourceManager.LoadTexture("textures/turret.png", "turret")
	bulletTexture = ResourceManager.LoadTexture("textures/bullet.png", "bullet")
}
