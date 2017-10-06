package tanklets

import (
	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type SimpleRenderer struct {
	shader        *Shader
	vao, vbo, ebo uint32
}

func NewSimpleRenderer(shader *Shader, projection mgl32.Mat4) *SimpleRenderer {
	renderer := &SimpleRenderer{shader: shader}
	shader.Use().SetMat4("projection", projection)

	vertices := []float32{
		0, 1, 0, 1,
		1, 0, 1, 0,
		0, 0, 0, 0,

		0, 1, 0, 1,
		1, 1, 1, 1,
		1, 0, 1, 0,
	}
	var indices = []uint32{
		0, 1, 3, // first triangle
		1, 2, 3, // second triangle
	}

	// VBO stores vertices in memory on the GPU
	// VAO defines the data layout (*VertexAttrib* calls)
	// EBO (element buffer) allows storing indices of what to draw, to make the VBO smaller (*Element* calls)
	gl.GenVertexArrays(1, &renderer.vao)
	gl.GenBuffers(1, &renderer.vbo)
	gl.GenBuffers(1, &renderer.ebo)
	// bind the Vertex Array Object first, then bind and set vertex buffer(s), and then configure vertex attributes(s).
	gl.BindVertexArray(renderer.vao)

	gl.BindBuffer(gl.ARRAY_BUFFER, renderer.vbo)
	// Can't take unsafe.Sizeof an array/slice in Go, it returns the size of the header.
	gl.BufferData(gl.ARRAY_BUFFER, 4*len(vertices), gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, renderer.ebo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, 4*len(indices), gl.Ptr(indices), gl.STATIC_DRAW)

	// Tells how to interpret the VBO data (stored in the VAO)
	gl.VertexAttribPointer(0, 3, gl.FLOAT, false, 3*4, gl.PtrOffset(0))
	gl.EnableVertexAttribArray(0)

	// note that this is allowed, the call to glVertexAttribPointer registered VBO as
	// the vertex attribute's bound vertex buffer object so afterwards we can safely unbind
	gl.BindBuffer(gl.ARRAY_BUFFER, 0)

	// remember: do NOT unbind the EBO while a VAO is active as the bound element buffer
	// object IS stored in the VAO; keep the EBO bound.
	//gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, 0);

	// You can unbind the VAO afterwards so other VAO calls won't accidentally modify this
	// VAO, but this rarely happens. Modifying other VAOs requires a call to glBindVertexArray
	// anyways so we generally don't unbind VAOs (nor VBOs) when it's not directly necessary.
	gl.BindVertexArray(0)

	return renderer
}

func (s *SimpleRenderer) Draw(x, y float32, sizex, sizey float32, rotate float32, r, g, b, a float32) {
	transform := mgl32.Translate3D(x, y, 0)

	transform = transform.Mul4(mgl32.Translate3D(0.5*sizex, 0.5*sizey, 0))
	transform = transform.Mul4(mgl32.HomogRotate3D(rotate, mgl32.Vec3{0, 0, 1}))
	transform = transform.Mul4(mgl32.Translate3D(-0.5*sizex, -0.5*sizey, 0))

	transform = transform.Mul4(mgl32.Scale3D(sizex, sizey, 1))

	s.shader.Use().SetMat4("transform", transform).SetVec4f("color", mgl32.Vec4{r, g, b, a})
	gl.BindVertexArray(s.vao)
	gl.DrawElements(gl.TRIANGLES, 6, gl.UNSIGNED_INT, gl.PtrOffset(0))
}
