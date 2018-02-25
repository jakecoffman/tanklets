package client

import (
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang-ui/nuklear/nk"

	"log"
	"time"
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

	network *Client // may be nil!

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
	if m.network == nil {
		return
	}

	if m.network.IsConnected {
		return
	}

	// handle packets until we're connected
network:
	for {
		select {
		case incoming := <-m.network.IncomingPackets:
			ProcessNetwork(incoming, nil, m.network)
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
	bounds := nk.NkRect(20, 20, 400, 400)
	update := nk.NkBegin(ctx, "Tank Game", bounds, nk.WindowTitle | nk.WindowMovable | nk.WindowScalable)

	if update > 0 {
		switch m.state {
		case PlayJoin:
			nk.NkLayoutRowDynamic(ctx, 0, 1)
			{
				if m.network.IsConnected {
					nk.NkLabel(ctx, "Connected!", nk.TextLeft)
				} else {
					if time.Now().Sub(m.startedConnecting) > 2*time.Second {
						nk.NkLabel(ctx, "Timed out!", nk.TextLeft)
						m.network.IsConnecting = false
					}

					if m.network.IsConnecting {
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
				if nk.NkButtonLabel(ctx, "Play Now") > 0 {
					m.state = PlayJoin
					m.startedConnecting = time.Now()
					// TODO error handling
					var err error
					m.network, err = NewClient("127.0.0.1:1234")
					if err != nil {
						log.Println(err)
					} else {
						go m.network.Recv()
					}
				}
			}
			nk.NkLayoutRowDynamic(ctx, 0, 1)
			{
				nk.NkLabel(ctx, "Join Custom", nk.TextLeft)
				nk.NkEditStringZeroTerminated(ctx, nk.EditSimple, m.joinText, joinTextSize, nk.NkFilterDefault)
				if nk.NkButtonLabel(ctx, "Join") > 0 {
					m.state = PlayJoin
					m.startedConnecting = time.Now()
					var err error
					m.network, err = NewClient(string(m.joinText))
					if err != nil {
						log.Println(err)
					} else {
						go m.network.Recv()
					}
				}
			}
		}
	}

	nk.NkEnd(ctx)
	gl.ClearColor(.1, .1, .1, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	nk.NkPlatformRender(nk.AntiAliasingOn, MaxVertexBuffer, MaxElementBuffer)

	if m.network != nil && m.network.IsConnected {
		CurrentScene = NewGameScene(m.window, ctx, m.network)
		m.Destroy()
	}
}

func (m *MainMenuScene) Destroy() {

}
