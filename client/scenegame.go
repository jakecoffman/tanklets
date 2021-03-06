package client

import (
	"fmt"
	"log"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/golang-ui/nuklear/nk"
	"github.com/jakecoffman/cp"
	"github.com/jakecoffman/tanklets"
	"github.com/jakecoffman/tanklets/gutils"
	"github.com/jakecoffman/tanklets/pkt"
)

// keep track of who you are
var Me pkt.PlayerID

type GameScene struct {
	window *glfw.Window
	ctx    *nk.Context

	network *Client

	game         *tanklets.Game
	isReady      bool
	hideDebug    bool
	displayNames bool

	nameText []byte
}

func NewGameScene(w *glfw.Window, ctx *nk.Context, network *Client) Scene {
	game := tanklets.NewGame(800, 600)

	w.SetMouseButtonCallback(MouseButtonCallback)

	fmt.Println("Sending JOIN command")
	network.Send(pkt.Join{})

	name := fmt.Sprintf("Player %v", Me)
	nameText := append([]byte(name), make([]byte, 256-len(name))...)

	return &GameScene{
		window:   w,
		ctx:      ctx,
		game:     game,
		nameText: nameText,
		network: network,
	}
}

var accumulator = 0.

const physicsTickrate = 1.0 / 180.0

func (g *GameScene) Update(dt float64) {
	if !g.network.IsConnected {
		CurrentScene = NewMainMenuScene(g.window, g.ctx)
		g.Destroy()
		return
	}

network:
	for {
		select {
		case incoming := <-g.network.IncomingPackets:
			ProcessNetwork(incoming, g.game, g.network)
		default:
			// no data to process this frame
			break network
		}
	}
	g.ProcessInput()
	g.game.Update(dt)

	accumulator += dt
	for accumulator >= physicsTickrate {
		for _, tank := range g.game.Tanks {
			FixedUpdate(tank)
		}
		g.game.Space.Step(physicsTickrate)
		accumulator -= physicsTickrate
	}
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

	Text.SetProjection(projection)
	Text.SetColor(1, 0, 1, 0.5)
	for _, t := range g.game.Tanks {
		Text.Print(t.Name, float32(t.Position().X-float64(len(t.Name)*5)), float32(t.Position().Y-20), 0.2)
	}

	switch g.game.State {
	case tanklets.StateStartCountdown:
		Text.SetProjection(mgl32.Ortho2D(0, float32(screenWidth), float32(screenHeight), 0))
		Text.SetColor(0, 1, 0, 1)
		diff := time.Now().Sub(g.game.StartTime)
		if diff < 1 * time.Second {
			Text.Print("3", float32(screenWidth)/2-10, float32(screenHeight)/2, 1)
		} else if diff < 2 * time.Second {
			Text.Print("2", float32(screenWidth)/2-10, float32(screenHeight)/2, 1)
		} else if diff < 3 * time.Second {
			Text.Print("1", float32(screenWidth)/2-10, float32(screenHeight)/2, 1)
		}
	case tanklets.StateWinCountdown:
		Text.SetProjection(mgl32.Ortho2D(0, float32(screenWidth), float32(screenHeight), 0))
		Text.SetColor(0, 1, 0, 1)
		Text.Print(g.game.WinningPlayer.Name + " won", float32(screenWidth)/2-200, float32(screenHeight)/2, .5)
	case tanklets.StateFailCountdown:
		Text.SetProjection(mgl32.Ortho2D(0, float32(screenWidth), float32(screenHeight), 0))
		Text.SetColor(0, 1, 0, 1)
		Text.Print("Everyone's dead.", float32(screenWidth)/2-200, float32(screenHeight)/2, .5)
	}

	g.Gui()
}

const (
	debugW = 180
	debugH = 110
)

func (g *GameScene) Gui() {
	nk.NkPlatformNewFrame()

	if g.hideDebug {
		bounds := nk.NkRect(float32(screenWidth)-debugW, 0, debugW, debugH)
		update := nk.NkBegin(g.ctx, "Debug", bounds, nk.WindowMovable)
		if update > 0 {
			nk.NkLayoutRowDynamic(g.ctx, 20, 1)
			{
				nk.NkLabel(g.ctx, fmt.Sprint("ping: ", pkt.MyPing), nk.TextLeft)
				nk.NkLabel(g.ctx, fmt.Sprint("fps: ", fps), nk.TextLeft)
				nk.NkLabel(g.ctx, fmt.Sprint("in: ", gutils.Bytes(g.network.NetworkIn)), nk.TextLeft)
				nk.NkLabel(g.ctx, fmt.Sprint("out: ", gutils.Bytes(g.network.NetworkOut)), nk.TextLeft)
			}
		}
		nk.NkEnd(g.ctx)
	}

	if g.game.State == tanklets.StateWaiting {
		if g.isReady {
			bounds := nk.NkRect(float32(screenWidth)/2-200, float32(screenHeight)/2-50, 400, 100)
			update := nk.NkBegin(g.ctx, "Waiting", bounds, nk.WindowTitle|nk.WindowBorder|nk.WindowMovable)
			if update > 0 {
				nk.NkLayoutRowDynamic(g.ctx, 0, 1)
				nk.NkLabel(g.ctx, "Waiting for all users to ready up...", nk.TextCentered)
			}
			nk.NkEnd(g.ctx)
		} else {
			bounds := nk.NkRect(100, 50, 400, 300)
			update := nk.NkBegin(g.ctx, "Ready?", bounds, nk.WindowTitle|nk.WindowBorder|nk.WindowMovable)

			if update > 0 {
				nk.NkLayoutRowDynamic(g.ctx, 0, 1)
				{
					nk.NkLabel(g.ctx, "Click ready to begin.", nk.TextCentered)
					if nk.NkButtonLabel(g.ctx, "Ready") > 0 {
						g.isReady = true
						LeftClick = false
						g.network.Send(pkt.Ready{})
						g.network.Send(pkt.Ready{})
						g.network.Send(pkt.Ready{})
					}
					nk.NkLabel(g.ctx, "Enter your name", nk.TextLeft)
					nk.NkEditStringZeroTerminated(g.ctx, nk.EditSimple, g.nameText, 11, nk.NkFilterDefault)
					if nk.NkButtonLabel(g.ctx, "Rename") > 0 {
						fmt.Println("Sending rename", strings.TrimRight(string(g.nameText), "\x00"))
						g.network.Send(pkt.Join{Name: strings.TrimRight(string(g.nameText), "\x00")})
					}
				}
			}
			nk.NkEnd(g.ctx)
		}
	}

	if g.game.State != tanklets.StatePlaying {
		bounds := nk.NkRect(0, 0, 200, 60+float32(30*len(g.game.Tanks)))
		update := nk.NkBegin(g.ctx, "Score", bounds, nk.WindowTitle|nk.WindowMinimizable)
		if update > 0 {
			nk.NkLayoutRowDynamic(g.ctx, 0, 1)
			tanks := make([]*tanklets.Tank, 0, len(g.game.Tanks))
			for _, t := range g.game.Tanks {
				tanks = append(tanks, t)
			}
			sort.Slice(tanks, func(i, j int) bool {
				return tanks[i].ID < tanks[j].ID
			})
			for _, t := range tanks {
				if t.Name != "" {
					nk.NkLabel(g.ctx, fmt.Sprint(t.Name, " - ", t.Score), nk.TextLeft)
				} else {
					nk.NkLabel(g.ctx, fmt.Sprint("Player ", t.ID, " - ", t.Score), nk.TextLeft)
				}
			}
		}
		nk.NkEnd(g.ctx)
	}

	nk.NkPlatformRender(nk.AntiAliasingOn, MaxVertexBuffer, MaxElementBuffer)
}

func (g *GameScene) Destroy() {
	g.network.Send(pkt.Disconnect{})
	g.network.Send(pkt.Disconnect{})
	g.network.Send(pkt.Disconnect{})
	g.network.Send(pkt.Disconnect{})
	g.network.Send(pkt.Disconnect{})
	g.network.Close()
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

	if game.State < tanklets.StatePlaying || myTank.Destroyed {
		return
	}

	turretAngle := math.Atan2(turret.Y, turret.X)
	angle := Player.Turret.Angle() - Player.Turret.Rotation().Unrotate(turret).ToAngle()
	if LeftClick {
		g.network.Send(pkt.Shoot{Angle: angle})
		Player.LastShot = time.Now()
	}

	RightDown = false
	LeftDown = false

	nextMove := pkt.Move{Turn: turn, Throttle: throttle}
	if nextMove != myTank.NextMove {
		myTank.NextMove = nextMove
		g.network.Send(myTank.NextMove)
	}
	aim := pkt.Aim{TurretAngle: float32(turretAngle)}
	if math.Abs(float64(aim.TurretAngle - myTank.Aim)) > 0.001 {
		myTank.Aim = aim.TurretAngle
		g.network.Send(aim)
	}
}
