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
	"github.com/jakecoffman/tanklets/pkt"
)

// keep track of who you are
var Me pkt.PlayerID

type GameScene struct {
	window *glfw.Window
	ctx *nk.Context

	game      *tanklets.Game
	isReady   bool
	hideDebug bool
	displayNames bool

	nameText []byte
}

func NewGameScene(w *glfw.Window, ctx *nk.Context) Scene {
	game := tanklets.NewGame(800, 600)

	w.SetMouseButtonCallback(MouseButtonCallback)

	fmt.Println("Sending JOIN command")
	tanklets.ClientSend(pkt.Join{})
	return &GameScene{
		window: w,
		ctx: ctx,
		game: game,
		nameText: make([]byte, 256),
	}
}

var accumulator = 0.
const physicsTickrate = 1.0 / 180.0

func (g *GameScene) Update(dt float64) {
	accumulator += dt
	for accumulator >= physicsTickrate {
		myTank := g.game.Tanks[Me]
		if myTank == nil {
			break
		}
		myTank.FixedUpdate(physicsTickrate)
		g.game.Space.Step(physicsTickrate)
		accumulator -= physicsTickrate
	}

network:
	for {
		select {
		case incoming := <-tanklets.IncomingPackets:
			ProcessNetwork(incoming, g.game)
		default:
			// no data to process this frame
			break network
		}
	}
	g.ProcessInput()
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
	SpaceRenderer.ClearRenderer()

	for _, tank := range g.game.Tanks {
		DrawTank(tank)
	}

	for _, bullet := range g.game.Bullets {
		DrawBullet(g.game, bullet)
	}

	for _, box := range g.game.Boxes {
		box.Body.EachShape(func(shape *cp.Shape) {
			SpaceRenderer.DrawShape(shape)
		})
	}

	for _, wall := range g.game.Walls {
		SpaceRenderer.DrawShape(wall)
	}

	SpaceRenderer.FlushRenderer()

	myTank := g.game.Tanks[Me]

	if myTank != nil && myTank.Destroyed {
		Text.SetProjection(mgl32.Ortho2D(0, float32(screenWidth), float32(screenHeight), 0))
		Text.Print("You died", float32(screenWidth)/2, float32(screenHeight)/2, 2)
	}

	if g.displayNames {
		Text.SetProjection(projection)
		for _, t := range g.game.Tanks {
			Text.Print(t.Name, float32(t.Position().X-10), float32(t.Position().Y-20), 0.5)
		}
	}

	g.Gui()
}

const (
	debugW = 180
	debugH = 140
)

func (g *GameScene) Gui() {
	nk.NkPlatformNewFrame()

	if !g.hideDebug {
		bounds := nk.NkRect(float32(screenWidth)-debugW, 0, debugW, debugH)
		update := nk.NkBegin(g.ctx, "Debug", bounds, nk.WindowMovable)
		if update > 0 {
			nk.NkLayoutRowDynamic(g.ctx, 20, 1)
			{
				nk.NkLabel(g.ctx, fmt.Sprint("ping: ", pkt.MyPing), nk.TextLeft)
				nk.NkLabel(g.ctx, fmt.Sprint("fps: ", fps), nk.TextLeft)
				nk.NkLabel(g.ctx, fmt.Sprint("in: ", gutils.Bytes(tanklets.NetworkIn)), nk.TextLeft)
				nk.NkLabel(g.ctx, fmt.Sprint("out: ", gutils.Bytes(tanklets.NetworkOut)), nk.TextLeft)
			}
		}
		nk.NkEnd(g.ctx)
	}

	if g.game.State == tanklets.GameStateWaiting {
		bounds := nk.NkRect(100, 50, 400, 300)
		update := nk.NkBegin(g.ctx, "Ready?", bounds, nk.WindowTitle|nk.WindowBorder|nk.WindowMovable)

		if update > 0 {
			nk.NkLayoutRowDynamic(g.ctx, 0, 1)
			{
				nk.NkLabel(g.ctx, "Waiting for all users to ready up...", nk.TextCentered)
				if !g.isReady {
					if nk.NkButtonLabel(g.ctx, "Ready") > 0 {
						g.isReady = true
						LeftClick = false
						tanklets.ClientSend(pkt.Ready{})
						tanklets.ClientSend(pkt.Ready{})
						tanklets.ClientSend(pkt.Ready{})
					}
					nk.NkLabel(g.ctx, "Enter your name", nk.TextLeft)
					nk.NkEditStringZeroTerminated(g.ctx, nk.EditSimple, g.nameText, 11, nk.NkFilterDefault)
					if nk.NkButtonLabel(g.ctx, "Rename") > 0 {
						fmt.Println("Sending rename", string(g.nameText[:11]))
						tanklets.ClientSend(pkt.Join{Name: string(g.nameText[:11])})
					}
				}
			}
		}
		nk.NkEnd(g.ctx)
	}
	nk.NkPlatformRender(nk.AntiAliasingOn, MaxVertexBuffer, MaxElementBuffer)
}

func (g *GameScene) Destroy() {
	tanklets.ClientSend(pkt.Disconnect{})
	tanklets.ClientSend(pkt.Disconnect{})
	tanklets.ClientSend(pkt.Disconnect{})
	tanklets.ClientSend(pkt.Disconnect{})
	tanklets.ClientSend(pkt.Disconnect{})
	tanklets.NetClose()
}

func (g *GameScene) ProcessInput() {
	game := g.game

	if Player == nil {
		Player = game.Tanks[Me]
		if Player == nil {
			return
		}
	}

	if Keys[glfw.KeyF10] {
		g.hideDebug = !g.hideDebug
		Keys[glfw.KeyF10] = false
	}

	g.displayNames = Keys[glfw.KeyTab]

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
	myTank := game.Tanks[Me]
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

	if game.State != tanklets.GameStatePlaying || myTank.Destroyed {
		return
	}

	if LeftClick {
		tanklets.ClientSend(pkt.Shoot{})
		Player.LastShot = time.Now()
	}

	RightDown = false
	LeftDown = false

	// send all of this input to the server
	myTank.NextMove = pkt.Move{Turn: turn, Throttle: throttle, TurretAngle: math.Atan2(turret.Y, turret.X)}
	tanklets.ClientSend(myTank.NextMove)
}
