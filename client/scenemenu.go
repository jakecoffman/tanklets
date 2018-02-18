package client

import (
	"github.com/golang-ui/nuklear/nk"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"time"
	"github.com/jakecoffman/tanklets"
)

const (
	PlayMenu   = iota
	PlayOnline
	PlayHost
	PlayJoin
)

type MainMenuScene struct {
	window   *glfw.Window
	ctx      *nk.Context
	state    int
	joinText []byte

	startedConnecting time.Time
}

func NewMainMenuScene(w *glfw.Window, ctx *nk.Context) Scene {
	// TODO load resources here
	return &MainMenuScene{
		window:   w,
		ctx:      ctx,
		state:    PlayMenu,
		joinText: make([]byte, joinTextSize),
	}
}

const joinTextSize = 256

func (m *MainMenuScene) Update(dt float64) {
	if tanklets.ClientIsConnected {
		return
	}

	// handle packets until we're connected
network:
	for {
		select {
		case incoming := <-tanklets.IncomingPackets:
			ProcessNetwork(incoming, nil)
			break network
		default:
			// no data to process this frame
			break network
		}
	}
}

func (m *MainMenuScene) Render() {
	nk.NkPlatformNewFrame()
	ctx := m.ctx

	// Layout
	bounds := nk.NkRect(50, 50, 230, 230)
	update := nk.NkBegin(ctx, "Tank Game", bounds, nk.WindowTitle | nk.WindowBorder)

	if update > 0 {
		switch m.state {
		case PlayJoin:
			nk.NkLayoutRowDynamic(ctx, 0, 1)
			{
				if tanklets.ClientIsConnected {
					nk.NkLabel(ctx, "Connected!", nk.TextLeft)
				} else {
					if time.Now().Sub(m.startedConnecting) > 2*time.Second {
						nk.NkLabel(ctx, "Timed out!", nk.TextLeft)
						tanklets.ClientIsConnecting = false
					}

					if tanklets.ClientIsConnecting {
						nk.NkLabel(ctx, "Connecting...", nk.TextLeft)
					}
				}
			}
			nk.NkLayoutRowDynamic(ctx, 0, 1)
			{
				if nk.NkButtonLabel(ctx, "Cancel") > 0 {
					m.state = PlayMenu
				}
			}
		default:
			nk.NkLayoutRowDynamic(ctx, 0, 1)
			{
				if nk.NkButtonLabel(ctx, "Play Online") > 0 {
					m.state = PlayJoin
					m.startedConnecting = time.Now()
					tanklets.NetInit("127.0.0.1:1234")
					go Recv()
				}
			}
			nk.NkLayoutRowDynamic(ctx, 0, 1)
			{
				nk.NkLabel(ctx, "Join Custom", nk.TextLeft)
				nk.NkEditStringZeroTerminated(ctx, nk.EditSimple, m.joinText, joinTextSize, nk.NkFilterDefault)
				if nk.NkButtonLabel(ctx, "Join") > 0 {
					m.state = PlayJoin
					m.startedConnecting = time.Now()
					tanklets.NetInit(string(m.joinText))
					go Recv()
				}
			}
		}
	}

	nk.NkEnd(ctx)
	gl.ClearColor(.1, .1, .1, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	nk.NkPlatformRender(nk.AntiAliasingOn, MaxVertexBuffer, MaxElementBuffer)

	if tanklets.ClientIsConnected {
		CurrentScene = NewGameScene(m.window, ctx)
		m.Destroy()
	}
}

func (m *MainMenuScene) Destroy() {

}
