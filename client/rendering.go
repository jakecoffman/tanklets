package client

import (
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/tanklets/client/glpers"
)

//go:generate go-bindata -ignore \.DS_Store -pkg data -o data/data.go assets/...

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
	ResourceManager.LoadShader("main.vs.glsl", "main.fs.glsl", "sprite")
	ResourceManager.LoadShader("simple.vs.glsl", "simple.fs.glsl", "simple")
	ResourceManager.LoadShader("text.vs.glsl", "text.fs.glsl", "text")
	ResourceManager.LoadShader("cp.vs.glsl", "cp.fs.glsl", "cp")

	// renderers
	w, h := float32(screenWidth), float32(screenHeight)
	projection = mgl32.Ortho2D(0, w, h, 0)
	Text = glpers.NewTextRenderer(ResourceManager.Shader("text"), w, h, "Roboto-Light.ttf", 96)
	Text.SetColor(.8, .8, .3, 1)
	Simple = glpers.NewSimpleRenderer(ResourceManager.Shader("simple"), projection)
	Renderer = glpers.NewSpriteRenderer(ResourceManager.Shader("sprite"), projection)
	SpaceRenderer = glpers.NewCPRenderer(ResourceManager.Shader("cp"), projection)

	// textures
	tankTexture = ResourceManager.LoadTexture("tank.png", "tank")
	turretTexture = ResourceManager.LoadTexture("turret.png", "turret")
	bulletTexture = ResourceManager.LoadTexture("bullet.png", "bullet")
}
