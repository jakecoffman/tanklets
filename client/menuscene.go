package client

import (
	"github.com/golang-ui/nuklear/nk"
	"log"
)

const (
	PlayNone = iota
	PlayOnline
	PlayLAN
)

type MainMenuScene struct {
	play int
}

func NewMainMenuScene() *MainMenuScene {
	return &MainMenuScene{play: PlayNone}
}

func (m *MainMenuScene) Update(dt float64) {

}

func (m *MainMenuScene) Render() {
	nk.NkPlatformNewFrame()

	// Layout
	bounds := nk.NkRect(50, 50, 200, 230)
	update := nk.NkBegin(ctx, "Welcome", bounds,
		nk.WindowBorder|nk.WindowMovable|nk.WindowScalable|nk.WindowMinimizable|nk.WindowTitle)

	if update > 0 {
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		{
			nk.NkLabel(ctx, "Welcome to Tank Game, what do you want to do?", nk.TextLeft)
			if nk.NkButtonLabel(ctx, "Play Online") > 0 {
				log.Println("Play online")
				m.play = PlayOnline
			}
			if nk.NkButtonLabel(ctx, "Play LAN") > 0 {
				log.Println("Play lan")
				m.play = PlayLAN
			}
		}
	}

	nk.NkEnd(ctx)
	nk.NkPlatformRender(nk.AntiAliasingOn, maxVertexBuffer, maxElementBuffer)
}

func (m *MainMenuScene) Transition() Scene {
	if m.play == PlayOnline {
		return NewGameScene()
	}
	return nil
}

func (m *MainMenuScene) Destroy() {

}