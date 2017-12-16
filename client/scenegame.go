package client

import (
	"github.com/jakecoffman/tanklets"
	"fmt"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/golang-ui/nuklear/nk"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/jakecoffman/cp"
	"log"
	"time"
	"github.com/go-gl/glfw/v3.2/glfw"
)

type GameScene struct {

}

func NewGameScene() *GameScene {
	tanklets.NewGame(800, 600)
	tanklets.NetInit()

	fmt.Println("Sending JOIN command")
	tanklets.Send(tanklets.Join{}, tanklets.ServerAddr)
	return &GameScene{}
}

var accumulator = 0.
const physicsTickrate = 1.0 / 180.0

func (g *GameScene) Update(dt float64) {
	tanklets.ProcessIncoming()

	accumulator += dt
	for accumulator >= physicsTickrate {
		myTank := tanklets.Tanks[tanklets.Me]
		if myTank == nil {
			break
		}
		myTank.FixedUpdate(physicsTickrate)
		tanklets.Space.Step(physicsTickrate)
		accumulator -= physicsTickrate
	}

	ProcessInput()
	tanklets.Update(dt)

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
}

func (g *GameScene) Render(ctx *nk.Context) {
	// TODO only set projection when it changes
	Renderer.SetProjection(projection)

	gl.ClearColor(.1, .1, .1, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	// useful for debugging space issues
	SpaceRenderer.SetProjection(projection)
	SpaceRenderer.DrawSpace(tanklets.Space)

	for _, tank := range tanklets.Tanks {
		DrawTank(tank)
	}

	for _, bullet := range tanklets.Bullets {
		DrawBullet(bullet)
	}

	if tanklets.State == tanklets.GAME_WAITING {
		Text.Print("Connecting", 50, 100, 1)
	}

	if tanklets.State == tanklets.GAME_DEAD {
		Text.Print("You died", 50, 50, 1)
	}

	nk.NkPlatformNewFrame()

	bounds := nk.NkRect(0, 0, 200, 120)
	update := nk.NkBegin(ctx, "Debug", bounds, nk.WindowMinimizable)

	if update > 0 {
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		{
			nk.NkLabel(ctx, fmt.Sprint("ping: ", tanklets.MyPing), nk.TextLeft)
		}
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		{
			nk.NkLabel(ctx, fmt.Sprint("in: ", tanklets.Bytes(tanklets.NetworkIn)), nk.TextLeft)
			nk.NkLabel(ctx, fmt.Sprint("out: ", tanklets.Bytes(tanklets.NetworkOut)), nk.TextLeft)
		}
	}
	nk.NkEnd(ctx)
	nk.NkPlatformRender(nk.AntiAliasingOn, MaxVertexBuffer, MaxElementBuffer)
}

func (g *GameScene) Transition() Scene {
	return nil
}

func (g *GameScene) Destroy() {
	fmt.Println("Sending DISCONNECT")
	tanklets.Send(tanklets.Disconnect{}, tanklets.ServerAddr)
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
	// TODO only recalculate when things have changed
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
