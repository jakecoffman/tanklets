package client

import (
	"log"
	"time"

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


func ProcessInput() {
	if tanklets.State != tanklets.GAME_PLAYING {
		return
	}

	if Player == nil {
		Player = tanklets.Tanks[tanklets.Me]
		if Player == nil {
			return
		}
	}

	var turn, throttle int8
	if Keys[glfw.KeyD] {
		turn = 1
	} else if Keys[glfw.KeyA] {
		turn = -1
	}
	if Keys[glfw.KeyW] {
		throttle = -1
	} else if Keys[glfw.KeyS] {
		throttle = 1
	}

	// update projection and mouse world position
	myTank := tanklets.Tanks[tanklets.Me]
	pos := myTank.Position()
	x, y := float32(pos.X), float32(pos.Y)
	sw, sh := float32(screenWidth), float32(screenHeight)
	projection = mgl32.Ortho2D(x-sw/2., x+sw/2., y+sh/2., y-sh/2.)
	obj, err := mgl32.UnProject(
		mgl32.Vec3{float32(mouse.X), sh - float32(mouse.Y), 0},
		identityMatrix,
		projection,
		0, 0,
		screenWidth, screenHeight,
	)
	var turret cp.Vector
	if err != nil {
		log.Println(err)
	} else {
		mouseWorld := cp.Vector{float64(obj.X()), float64(obj.Y())}
		turret = mouseWorld.Sub(Player.Turret.Body.Position())
	}

	if LeftClick {
		tanklets.Send(tanklets.Shoot{}, tanklets.ServerAddr)
		Player.LastShot = time.Now()
	}

	RightDown = false
	LeftDown = false

	// TODO separate turret aim into a message sent less often since it's never 0 now
	if turn == 0.0 && throttle == 0.0 && turret.X == 0 && turret.Y == 0 {
		return
	}

	// send all of this input to the server
	myTank.NextMove = tanklets.Move{Turn: turn, Throttle: throttle, TurretX: turret.X, TurretY: turret.Y}
	tanklets.Send(myTank.NextMove, tanklets.ServerAddr)
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
