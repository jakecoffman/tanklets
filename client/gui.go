package client

import (
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang-ui/nuklear/nk"
)

const (
	MaxVertexBuffer  = 512 * 1024
	MaxElementBuffer = 128 * 1024
)

func GuiInit(win *glfw.Window) (*nk.Context, *nk.UserFont) {
	ctx := nk.NkPlatformInit(win, nk.PlatformInstallCallbacks)

	atlas := nk.NewFontAtlas()
	nk.NkFontStashBegin(&atlas)
	sansFont := nk.NkFontAtlasAddDefault(atlas, 16, nil)
	nk.NkFontStashEnd()
	if sansFont != nil {
		sansFontHandle := sansFont.Handle()
		nk.NkStyleSetFont(ctx, sansFontHandle)
		return ctx, sansFontHandle
	}
	return ctx, nil
}

func GuiDestroy() {
	nk.NkPlatformShutdown()
}

type Option uint8

const (
	Easy Option = 0
	Hard Option = 1
)

func flag(v bool) int32 {
	if v {
		return 1
	}
	return 0
}

type State struct {
	BgColor nk.Color
	Prop    int32
	Opt     Option
}
