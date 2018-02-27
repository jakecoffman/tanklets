package client

import (
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang-ui/nuklear/nk"
	"github.com/jakecoffman/tanklets/server"
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

	client *Client // may be nil!
	server *server.Server // may be nil!

	startedConnecting time.Time
}

const (
	defaultJoin = "127.0.0.1:8999"
	joinTextSize = 256
)

func NewMainMenuScene(w *glfw.Window, ctx *nk.Context) Scene {
	// TODO load resources here
	return &MainMenuScene{
		window:   w,
		ctx:      ctx,
		state:    PlayMenu,
		joinText: append([]byte(defaultJoin), make([]byte, joinTextSize-len(defaultJoin))...),
	}
}

func (m *MainMenuScene) Update(dt float64) {
	if m.client == nil {
		return
	}

	if m.client.IsConnected {
		return
	}

	// handle packets until we're connected
	for {
		select {
		case incoming := <-m.client.IncomingPackets:
			ProcessNetwork(incoming, nil, m.client)
			return
		default:
			// no data to process this frame
			return
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
				if m.client.IsConnected {
					nk.NkLabel(ctx, "Connected!", nk.TextLeft)
				} else {
					if time.Now().Sub(m.startedConnecting) > 5*time.Second {
						nk.NkLabel(ctx, "Timed out!", nk.TextLeft)
						m.client.IsConnecting = false
					}

					if m.client.IsConnecting {
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
				if nk.NkButtonLabel(ctx, "Host & Join") > 0 {
					m.server = server.NewServer("0.0.0.0:8999")
					// TODO a way to stop this loop
					go m.server.Recv()
					go server.Loop(m.server)
					m.state = PlayJoin
					m.startedConnecting = time.Now()
					// TODO error handling
					var err error
					m.client, err = NewClient("127.0.0.1:8999")
					if err != nil {
						log.Println(err)
					} else {
						go m.client.Recv()
					}
				}
			}
			nk.NkLayoutRowDynamic(ctx, 0, 1)
			{
				nk.NkLabel(ctx, "Online", nk.TextLeft)
				if nk.NkButtonLabel(ctx, "Play Now") > 0 {
					m.state = PlayJoin
					m.startedConnecting = time.Now()
					var err error
					m.client, err = NewClient("tanks.jakecoffman.com:1234")
					if err != nil {
						log.Println(err)
					} else {
						go m.client.Recv()
					}
				}
				nk.NkLabel(ctx, "Local", nk.TextLeft)
				nk.NkEditStringZeroTerminated(ctx, nk.EditSimple, m.joinText, joinTextSize, nk.NkFilterDefault)
			}
			nk.NkLayoutRowDynamic(ctx, 0, 2)
			{
				if nk.NkButtonLabel(ctx, "Join") > 0 {
					m.state = PlayJoin
					m.startedConnecting = time.Now()
					var err error
					m.client, err = NewClient(string(m.joinText))
					if err != nil {
						log.Println(err)
					} else {
						go m.client.Recv()
					}
				}
				if nk.NkButtonLabel(ctx, "Clear") > 0 {

				}
			}
		}
	}

	nk.NkEnd(ctx)
	gl.ClearColor(.1, .1, .1, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	nk.NkPlatformRender(nk.AntiAliasingOn, MaxVertexBuffer, MaxElementBuffer)

	if m.client != nil && m.client.IsConnected {
		CurrentScene = NewGameScene(m.window, ctx, m.client)
		m.Destroy()
	}
}

func (m *MainMenuScene) Destroy() {

}
