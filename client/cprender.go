package client

import (
	"math"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/jakecoffman/cp"
	"github.com/go-gl/mathgl/mgl32"
)

type CPRenderer struct {
	shader   *Shader
	vao, vbo uint32
}

func NewCPRenderer(shader *Shader, projection mgl32.Mat4) *CPRenderer {
	shader.Use().SetMat4("projection", projection)

	var vao, vbo uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	CheckGLErrors()

	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)

	CheckGLErrors()

	SetAttribute(shader.ID, "vertex", 2, gl.FLOAT, 48, 0)
	SetAttribute(shader.ID, "aa_coord", 2, gl.FLOAT, 48, 8)
	SetAttribute(shader.ID, "fill_color", 4, gl.FLOAT, 48, 16)
	SetAttribute(shader.ID, "outline_color", 4, gl.FLOAT, 48, 32)

	gl.BindVertexArray(0)

	CheckGLErrors()
	return &CPRenderer{
		shader: shader,
		vao:    vao,
		vbo:    vbo,
	}
}

func (r *CPRenderer) SetProjection(projection mgl32.Mat4) {
	r.shader.Use().SetMat4("projection", projection)
}

const (
	DrawPointLineScale = 1
	DrawOutlineWidth   = 1
)

type FColor struct {
	R, G, B, A float32
}

func (r *CPRenderer) DrawSpace(space *cp.Space) {
	ClearRenderer()
	space.EachShape(func(obj *cp.Shape) {
		DrawShape(obj)
	})
	r.FlushRenderer()

	//for _, constraint := range space.constraints {
	//	DrawConstraint(constraint, options)
	//}

	//for _, arb := range space.arbiters {
	//	n := arb.n
	//
	//	for j := 0; j < arb.count; j++ {
	//		p1 := arb.body_a.p.Add(arb.contacts[j].r1)
	//		p2 := arb.body_b.p.Add(arb.contacts[j].r2)
	//
	//		a := p1.Add(n.Mult(-2))
	//		b := p2.Add(n.Mult(2))
	//		drawSeg(a, b, options.CollisionPointColor(), data)
	//	}
	//}
}

func DrawShape(shape *cp.Shape) {
	body := shape.Body()

	outline := FColor{1, 1, 1, 1}
	fill := FColor{1, 1, 1, 1}

	switch shape.Class.(type) {
	case *cp.Circle:
		circle := shape.Class.(*cp.Circle)
		DrawCircle(circle.TransformC(), body.Angle(), circle.Radius(), outline, fill)
	case *cp.Segment:
		seg := shape.Class.(*cp.Segment)
		DrawFatSegment(seg.TransformA(), seg.TransformB(), seg.Radius(), outline, fill)
	case *cp.PolyShape:
		poly := shape.Class.(*cp.PolyShape)

		count := poly.Count()
		verts := make([]cp.Vector, count)

		for i := 0; i < count; i++ {
			verts[i] = poly.TransformVert(i)
		}
		DrawPolygon(count, verts, poly.Radius(), outline, fill)
	default:
		panic("Unknown shape type")
	}
}

// 8 bytes
type v2f struct {
	x, y float32
}

func V2f(v cp.Vector) v2f {
	return v2f{float32(v.X), float32(v.Y)}
}

// 8*2 + 16*2 bytes = 48 bytes
type Vertex struct {
	vertex, aa_coord          v2f
	fill_color, outline_color FColor
}

type Triangle struct {
	a, b, c Vertex
}

var triangleStack []Triangle = []Triangle{}

func DrawCircle(pos cp.Vector, angle, radius float64, outline, fill FColor) {
	r := radius + 1/DrawPointLineScale
	a := Vertex{
		v2f{float32(pos.X - r), float32(pos.Y - r)},
		v2f{-1, -1},
		fill,
		outline,
	}
	b := Vertex{
		v2f{float32(pos.X - r), float32(pos.Y + r)},
		v2f{-1, 1},
		fill,
		outline,
	}
	c := Vertex{
		v2f{float32(pos.X + r), float32(pos.Y + r)},
		v2f{1, 1},
		fill,
		outline,
	}
	d := Vertex{
		v2f{float32(pos.X + r), float32(pos.Y - r)},
		v2f{1, -1},
		fill,
		outline,
	}

	t0 := Triangle{a, b, c}
	t1 := Triangle{a, c, d}

	triangleStack = append(triangleStack, t0)
	triangleStack = append(triangleStack, t1)

	DrawFatSegment(pos, pos.Add(cp.ForAngle(angle).Mult(radius-DrawPointLineScale*0.5)), 0, outline, fill)
}

func DrawSegment(a, b cp.Vector, fill FColor) {
	DrawFatSegment(a, b, 0, fill, fill)
}

func DrawFatSegment(a, b cp.Vector, radius float64, outline, fill FColor) {
	n := b.Sub(a).ReversePerp().Normalize()
	t := n.ReversePerp()

	var half float64 = 1.0 / DrawPointLineScale
	r := radius + half

	if r <= half {
		r = half
		fill = outline
	}

	nw := n.Mult(r)
	tw := t.Mult(r)
	v0 := V2f(b.Sub(nw.Add(tw)))
	v1 := V2f(b.Add(nw.Sub(tw)))
	v2 := V2f(b.Sub(nw))
	v3 := V2f(b.Add(nw))
	v4 := V2f(a.Sub(nw))
	v5 := V2f(a.Add(nw))
	v6 := V2f(a.Sub(nw.Sub(tw)))
	v7 := V2f(a.Add(nw.Add(tw)))

	t0 := Triangle{
		Vertex{v0, v2f{1, -1}, fill, outline},
		Vertex{v1, v2f{1, 1}, fill, outline},
		Vertex{v2, v2f{0, -1}, fill, outline},
	}
	t1 := Triangle{
		Vertex{v3, v2f{0, 1}, fill, outline},
		Vertex{v1, v2f{1, 1}, fill, outline},
		Vertex{v2, v2f{0, -1}, fill, outline},
	}
	t2 := Triangle{
		Vertex{v3, v2f{0, 1}, fill, outline},
		Vertex{v4, v2f{0, -1}, fill, outline},
		Vertex{v2, v2f{0, -1}, fill, outline},
	}
	t3 := Triangle{
		Vertex{v3, v2f{0, 1}, fill, outline},
		Vertex{v4, v2f{0, -1}, fill, outline},
		Vertex{v5, v2f{0, 1}, fill, outline},
	}
	t4 := Triangle{
		Vertex{v6, v2f{-1, -1}, fill, outline},
		Vertex{v4, v2f{0, -1}, fill, outline},
		Vertex{v5, v2f{0, 1}, fill, outline},
	}
	t5 := Triangle{
		Vertex{v6, v2f{-1, -1}, fill, outline},
		Vertex{v7, v2f{-1, 1}, fill, outline},
		Vertex{v5, v2f{0, 1}, fill, outline},
	}

	triangleStack = append(triangleStack, t0)
	triangleStack = append(triangleStack, t1)
	triangleStack = append(triangleStack, t2)
	triangleStack = append(triangleStack, t3)
	triangleStack = append(triangleStack, t4)
	triangleStack = append(triangleStack, t5)
}

func DrawPolygon(count int, verts []cp.Vector, radius float64, outline, fill FColor) {
	type ExtrudeVerts struct {
		offset, n cp.Vector
	}
	extrude := make([]ExtrudeVerts, count)

	for i := 0; i < count; i++ {
		v0 := verts[(i-1+count)%count]
		v1 := verts[i]
		v2 := verts[(i+1)%count]

		n1 := v1.Sub(v0).ReversePerp().Normalize()
		n2 := v2.Sub(v1).ReversePerp().Normalize()

		offset := n1.Add(n2).Mult(1.0 / (n1.Dot(n2) + 1.0))
		extrude[i] = ExtrudeVerts{offset, n2}
	}

	inset := -math.Max(0, 1.0/DrawPointLineScale-radius)
	for i := 0; i < count-2; i++ {
		v0 := V2f(verts[0].Add(extrude[0].offset.Mult(inset)))
		v1 := V2f(verts[i+1].Add(extrude[i+1].offset.Mult(inset)))
		v2 := V2f(verts[i+2].Add(extrude[i+2].offset.Mult(inset)))

		triangleStack = append(triangleStack, Triangle{
			Vertex{v0, v2f{}, fill, fill},
			Vertex{v1, v2f{}, fill, fill},
			Vertex{v2, v2f{}, fill, fill},
		})
	}

	outset := 1.0/DrawPointLineScale + radius - inset
	j := count - 1
	for i := 0; i < count; {
		vA := verts[i]
		vB := verts[j]

		nA := extrude[i].n
		nB := extrude[j].n

		offsetA := extrude[i].offset
		offsetB := extrude[j].offset

		innerA := vA.Add(offsetA.Mult(inset))
		innerB := vB.Add(offsetB.Mult(inset))

		inner0 := V2f(innerA)
		inner1 := V2f(innerB)
		outer0 := V2f(innerA.Add(nB.Mult(outset)))
		outer1 := V2f(innerB.Add(nB.Mult(outset)))
		outer2 := V2f(innerA.Add(offsetA.Mult(outset)))
		outer3 := V2f(innerA.Add(nA.Mult(outset)))

		n0 := V2f(nA)
		n1 := V2f(nB)
		offset0 := V2f(offsetA)

		triangleStack = append(triangleStack, Triangle{
			Vertex{inner0, v2f{}, fill, outline},
			Vertex{inner1, v2f{}, fill, outline},
			Vertex{outer1, n1, fill, outline},
		})
		triangleStack = append(triangleStack, Triangle{
			Vertex{inner0, v2f{}, fill, outline},
			Vertex{outer0, n1, fill, outline},
			Vertex{outer1, n1, fill, outline},
		})
		triangleStack = append(triangleStack, Triangle{
			Vertex{inner0, v2f{}, fill, outline},
			Vertex{outer0, n1, fill, outline},
			Vertex{outer2, offset0, fill, outline},
		})
		triangleStack = append(triangleStack, Triangle{
			Vertex{inner0, v2f{}, fill, outline},
			Vertex{outer2, offset0, fill, outline},
			Vertex{outer3, n0, fill, outline},
		})

		j = i
		i++
	}
}

func DrawDot(size float64, pos cp.Vector, fill FColor) {
	r := size * 0.5 / DrawPointLineScale
	a := Vertex{v2f{float32(pos.X - r), float32(pos.Y - r)}, v2f{-1, -1}, fill, fill}
	b := Vertex{v2f{float32(pos.X - r), float32(pos.Y + r)}, v2f{-1, 1}, fill, fill}
	c := Vertex{v2f{float32(pos.X + r), float32(pos.Y + r)}, v2f{1, 1}, fill, fill}
	d := Vertex{v2f{float32(pos.X + r), float32(pos.Y - r)}, v2f{1, -1}, fill, fill}

	triangleStack = append(triangleStack, Triangle{a, b, c})
	triangleStack = append(triangleStack, Triangle{a, c, d})
}

func DrawBB(bb cp.BB, outline FColor) {
	verts := []cp.Vector{
		{bb.R, bb.B},
		{bb.R, bb.T},
		{bb.L, bb.T},
		{bb.L, bb.B},
	}
	DrawPolygon(4, verts, 0, outline, FColor{})
}

func (r *CPRenderer) FlushRenderer() {
	gl.BindBuffer(gl.ARRAY_BUFFER, r.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(triangleStack)*(48*3), gl.Ptr(triangleStack), gl.STREAM_DRAW)

	gl.UseProgram(r.shader.ID)
	gl.Uniform1f(gl.GetUniformLocation(r.shader.ID, gl.Str("u_outline_coef\x00")), DrawPointLineScale)

	gl.BindVertexArray(r.vao)
	gl.DrawArrays(gl.TRIANGLES, 0, int32(len(triangleStack)*3))
	CheckGLErrors()
}

func ClearRenderer() {
	triangleStack = triangleStack[:0]
}
