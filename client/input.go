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

func ProcessInput(dt float64) {
	if tanklets.State != tanklets.GAME_PLAYING {
		return
	}

	if Player == nil {
		Player = tanklets.Tanks[tanklets.Me]
		if Player == nil {
			return
		}
	}

	var turn float64
	if Keys[glfw.KeyD] {
		turn = tanklets.TurnSpeed
		Player.ControlBody.SetAngle(Player.Body.Angle() + tanklets.TurnSpeed)
		// by applying to the body too, it will allow getting unstuck from corners
		Player.Body.SetAngle(Player.Body.Angle() + tanklets.TurnSpeed)
	}
	if Keys[glfw.KeyA] {
		turn = -tanklets.TurnSpeed
		Player.ControlBody.SetAngle(Player.Body.Angle() - tanklets.TurnSpeed)
		// by applying to the body too, it will allow getting unstuck from corners
		Player.Body.SetAngle(Player.Body.Angle() - tanklets.TurnSpeed)
	}
	var throttle float64
	if Keys[glfw.KeyW] {
		throttle = -1
		Player.ControlBody.SetVelocityVector(Player.Body.Rotation().Rotate(cp.Vector{Y: -tanklets.MaxSpeed}))
	} else if Keys[glfw.KeyS] {
		throttle = 1
		Player.ControlBody.SetVelocityVector(Player.Body.Rotation().Rotate(cp.Vector{Y: tanklets.MaxSpeed}))
	} else {
		Player.ControlBody.SetVelocity(0, 0)
	}

	if LeftClick {
		Player.Shoot(tanklets.Space)
		Player.LastShot = time.Now()
	}

	// update projection and mouse world position
	myTank := tanklets.Tanks[tanklets.Me]
	pos := myTank.Position()
	x, y := float32(pos.X), float32(pos.Y)
	projection = mgl32.Ortho2D(x-800./2., x+800./2., y+600./2., y-600./2.)
	obj, err := mgl32.UnProject(
		mgl32.Vec3{float32(mouse.X), 600 - float32(mouse.Y), 0},
		identityMatrix,
		projection,
		0, 0,
		800, 600,
	)
	var turretTurn float64
	if err != nil {
		log.Println(err)
	} else {
		mouseWorld := cp.Vector{float64(obj.X()), float64(obj.Y())}
		mouseDelta := mouseWorld.Sub(Player.Turret.Body.Position())
		turretTurn = Player.Turret.Rotation().Unrotate(mouseDelta).ToAngle()
		Player.Turret.SetAngle(Player.Turret.Angle() - turretTurn)
		Player.Turret.SetPosition(Player.Position())
	}

	RightDown = false
	LeftDown = false

	if turn == 0.0 && throttle == 0.0 && turretTurn == 0.0 {
		return
	}

	// send all of this input to the server
	move := tanklets.Move{Turn: turn, Throttle: throttle, Turret: turretTurn}
	tanklets.Send(move, tanklets.ServerAddr)
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
