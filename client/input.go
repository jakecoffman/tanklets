package client

import (
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/tanklets"
)

var (
	Player     *tanklets.Tank
	Keys       = [1024]bool{}
	mouse      cp.Vector
	LeftDown   bool
	RightDown  bool
	LeftClick  bool
	RightClick bool
)

var identityMatrix = mgl32.Mat4{
	1, 0, 0, 0,
	0, 1, 0, 0,
	0, 0, 1, 0,
	0, 0, 0, 1,
}

func CursorCallback(w *glfw.Window, xpos float64, ypos float64) {
	mouse = cp.Vector{xpos, ypos}
}

func MouseButtonCallback(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mod glfw.ModifierKey) {
	if button == glfw.MouseButton1 {
		LeftDown = action == glfw.Press
		LeftClick = LeftDown
		//if action == glfw.Press {
		//			// give the mouse click a little radius to make it easier to click small shapes.
		//			//radius := 5.0
		//
		//			//info := Space.PointQueryNearest(mouse, radius, GrabFilter)
		//			//
		//			//if info.Shape != nil && info.Shape.Body().Mass() < INFINITY {
		//			//	var nearest Vector
		//			//	if info.Distance > 0 {
		//			//		nearest = info.Point
		//			//	} else {
		//			//		nearest = mouse
		//			//	}
		//			//
		//			//	body := info.Shape.Body()
		//			//	mouseJoint = NewPivotJoint2(mouseBody, body, Vector{}, body.WorldToLocal(nearest))
		//			//	mouseJoint.SetMaxForce(50000)
		//			//	mouseJoint.SetErrorBias(math.Pow(1.0-0.15, 60.0))
		//			//	Space.AddConstraint(mouseJoint)
		//			}
		//		//} else if mouseJoint != nil {
		//		//	Space.RemoveConstraint(mouseJoint)
		//		//	mouseJoint = nil
		//		//}
	} else if button == glfw.MouseButton2 {
		RightDown = action == glfw.Press
		RightClick = RightDown
	}
}
