package client

import (
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type SpriteRenderer struct {
	shader  *Shader
	quadVAO uint32
}

func NewSpriteRenderer(shader *Shader, projection mgl32.Mat4) *SpriteRenderer {
	renderer := &SpriteRenderer{shader: shader}
	shader.Use().SetInt("sprite", 0).SetMat4("projection", projection)
	renderer.initRenderData()
	return renderer
}

var (
	DefaultSpriteSize = mgl32.Vec2{10, 10}
	DefaultRotate     = 0.0
	DefaultColor      = mgl32.Vec3{1, 1, 1}
)

func (s *SpriteRenderer) SetProjection(projection mgl32.Mat4) {
	s.shader.Use().SetMat4("projection", projection)
}

func (s *SpriteRenderer) DrawSprite(texture *Texture2D, position, size mgl32.Vec2, rotate float64, color mgl32.Vec3) {
	s.shader.Use()
	model := mgl32.Translate3D(position.X()-0.5*size.X(), position.Y()-0.5*size.Y(), 0)

	model = model.Mul4(mgl32.Translate3D(0.5*size.X(), 0.5*size.Y(), 0))
	model = model.Mul4(mgl32.HomogRotate3D(float32(rotate), mgl32.Vec3{0, 0, 1}))
	model = model.Mul4(mgl32.Translate3D(-0.5*size.X(), -0.5*size.Y(), 0))

	model = model.Mul4(mgl32.Scale3D(size.X(), size.Y(), 1))

	s.shader.SetMat4("model", model).SetVec3f("spriteColor", color)

	gl.ActiveTexture(gl.TEXTURE0)
	texture.Bind()

	gl.BindVertexArray(s.quadVAO)
	gl.DrawArrays(gl.TRIANGLES, 0, 6)
	gl.BindVertexArray(0)
}

func (s *SpriteRenderer) initRenderData() {
	var VBO uint32
	vertices := []float32{
		0, 1, 0, 1,
		1, 0, 1, 0,
		0, 0, 0, 0,

		0, 1, 0, 1,
		1, 1, 1, 1,
		1, 0, 1, 0,
	}

	gl.GenVertexArrays(1, &s.quadVAO)
	gl.GenBuffers(1, &VBO)

	gl.BindBuffer(gl.ARRAY_BUFFER, VBO)
	gl.BufferData(gl.ARRAY_BUFFER, len(vertices)*4, gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.BindVertexArray(s.quadVAO)
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointer(0, 4, gl.FLOAT, false, 4*4, gl.PtrOffset(0))
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)
	gl.BindVertexArray(0)
}

func (s *SpriteRenderer) Destroy() {
	gl.DeleteVertexArrays(1, &s.quadVAO)
}
