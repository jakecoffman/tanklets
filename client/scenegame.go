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
	"math"
	"github.com/jakecoffman/tanklets/gutils"
)

type GameScene struct {
	window *glfw.Window
	ctx *nk.Context

	game *tanklets.Game
}

func NewGameScene(w *glfw.Window, ctx *nk.Context) Scene {
	game := tanklets.NewGame(800, 600)

	w.SetMouseButtonCallback(MouseButtonCallback)

	fmt.Println("Sending JOIN command")
	tanklets.ClientSend(tanklets.Join{})
	return &GameScene{
		window: w,
		ctx: ctx,
		game: game,
	}
}

var accumulator = 0.
const physicsTickrate = 1.0 / 180.0

func (g *GameScene) Update(dt float64) {
	tanklets.ProcessIncoming(g.game)

	accumulator += dt
	for accumulator >= physicsTickrate {
		myTank := g.game.Tanks[tanklets.Me]
		if myTank == nil {
			break
		}
		myTank.FixedUpdate(physicsTickrate)
		g.game.Space.Step(physicsTickrate)
		accumulator -= physicsTickrate
	}

	ProcessInput(g.game)
	g.game.Update(dt)
}

func (g *GameScene) Render() {
	// TODO only set projection when it changes
	Renderer.SetProjection(projection)

	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.ClearColor(.1, .1, .1, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	// useful for debugging space issues
	SpaceRenderer.SetProjection(projection)
	SpaceRenderer.DrawSpace(g.game.Space)

	for _, tank := range g.game.Tanks {
		DrawTank(tank)
	}

	for _, bullet := range g.game.Bullets {
		DrawBullet(g.game, bullet)
	}

	//for _, box := range g.game.Boxes {
	//	DrawBox(g.game, box)
	//}

	if g.game.State == tanklets.GameStateWaiting {
		Text.Print("Connecting", 50, 100, 1)
	}

	if g.game.State == tanklets.GameStateDead {
		Text.Print("You died", 50, 50, 1)
	}

	nk.NkPlatformNewFrame()

	bounds := nk.NkRect(0, 0, 200, 120)
	update := nk.NkBegin(g.ctx, "Debug", bounds, nk.WindowMinimizable)

	if update > 0 {
		nk.NkLayoutRowDynamic(g.ctx, 20, 1)
		{
			nk.NkLabel(g.ctx, fmt.Sprint("ping: ", tanklets.MyPing), nk.TextLeft)
		}
		nk.NkLayoutRowDynamic(g.ctx, 20, 1)
		{
			nk.NkLabel(g.ctx, fmt.Sprint("in: ", gutils.Bytes(tanklets.NetworkIn)), nk.TextLeft)
			nk.NkLabel(g.ctx, fmt.Sprint("out: ", gutils.Bytes(tanklets.NetworkOut)), nk.TextLeft)
		}
	}
	nk.NkEnd(g.ctx)
	nk.NkPlatformRender(nk.AntiAliasingOn, MaxVertexBuffer, MaxElementBuffer)
}

func (g *GameScene) Destroy() {
	tanklets.ClientSend(tanklets.Disconnect{})
	tanklets.ClientSend(tanklets.Disconnect{})
	tanklets.ClientSend(tanklets.Disconnect{})
	tanklets.ClientSend(tanklets.Disconnect{})
	tanklets.ClientSend(tanklets.Disconnect{})
	tanklets.NetClose()
}

func ProcessInput(game *tanklets.Game) {
	if game.State != tanklets.GameStatePlaying {
		return
	}

	if Player == nil {
		Player = game.Tanks[tanklets.Me]
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
	myTank := game.Tanks[tanklets.Me]
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
		tanklets.ClientSend(tanklets.Shoot{})
		Player.LastShot = time.Now()
	}

	RightDown = false
	LeftDown = false

	// TODO separate turret aim into a message sent less often since it's never 0 now
	if turn == 0.0 && throttle == 0.0 && turret.X == 0 && turret.Y == 0 {
		return
	}

	// send all of this input to the server
	myTank.NextMove = tanklets.Move{Turn: turn, Throttle: throttle, TurretAngle: math.Atan2(turret.Y, turret.X)}
	tanklets.ClientSend(myTank.NextMove)
}
