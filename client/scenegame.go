package client

import (
	"github.com/jakecoffman/tanklets"
	"fmt"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/golang-ui/nuklear/nk"
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
	Renderer.SetProjection(projection)

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