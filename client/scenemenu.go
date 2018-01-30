package client

import (
	"github.com/golang-ui/nuklear/nk"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

const (
	PlayNone = iota
	PlayOnline
	PlayHost
	PlayLAN
)

type MainMenuScene struct {
	window *glfw.Window
	ctx *nk.Context
	state      int
	textBuffer []byte
}

func NewMainMenuScene(w *glfw.Window, ctx *nk.Context) Scene {
	// TODO load resources here
	return &MainMenuScene{
		window: w,
		ctx: ctx,
		state:      PlayNone,
		textBuffer: make([]byte, textBufferSize),
	}
}

const textBufferSize = 256*1024

func (m *MainMenuScene) Update(dt float64) {

}

func (m *MainMenuScene) Render() {
	nk.NkPlatformNewFrame()
	ctx := m.ctx

	// Layout
	bounds := nk.NkRect(50, 50, 230, 230)
	update := nk.NkBegin(ctx, "Tank Game", bounds, nk.WindowTitle | nk.WindowBorder)

	if update > 0 {
		switch m.state {
		case PlayLAN:
			nk.NkLayoutRowDynamic(ctx, 0, 1)
			{
				nk.NkEditStringZeroTerminated(ctx, nk.EditSimple, m.textBuffer, textBufferSize, nk.NkFilterDefault)
				nk.NkLayoutRowDynamic(ctx, 0, 2)
				{
					if nk.NkButtonLabel(ctx, "Connect") > 0 {
						m.state = PlayNone
					}
					if nk.NkButtonLabel(ctx, "Cancel") > 0 {
						m.state = PlayNone
					}
				}
			}
			nk.NkLayoutRowDynamic(ctx, 100, 1)
			{
				str := "I am a very model of a modern individual. I am a very model of a modern individual. I am a very model of a modern individual. I am a very model of a modern individual. I am a very model of a modern individual. I am a very model of a modern individual. I am a very model of a modern individual. I am a very model of a modern individual. I am a very model of a modern individual. I am a very model of a modern individual. "
				nk.NkEditStringZeroTerminated(ctx, nk.EditBox, []byte(str), int32(len(str)), nk.NkFilterDefault)
			}
		default:
			nk.NkLayoutRowDynamic(ctx, 0, 1)
			{
				if nk.NkButtonLabel(ctx, "Play") > 0 {
					m.state = PlayOnline
				}
			}
			nk.NkLayoutRowDynamic(ctx, 0, 1)
			{
				nk.NkLabel(ctx, "LAN", nk.TextLeft)
				if nk.NkButtonLabel(ctx, "Host") > 0 {
					m.state = PlayHost
				}
				if nk.NkButtonLabel(ctx, "Join") > 0 {
					m.state = PlayLAN
				}
			}
		}
	}

	nk.NkEnd(ctx)
	gl.ClearColor(.1, .1, .1, 1)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	nk.NkPlatformRender(nk.AntiAliasingOn, MaxVertexBuffer, MaxElementBuffer)

	if m.state == PlayOnline {
		CurrentScene = NewGameScene(m.window, ctx)
		m.Destroy()
	}
}

func (m *MainMenuScene) Destroy() {

}