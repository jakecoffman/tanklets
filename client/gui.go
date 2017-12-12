package client

import (
	"fmt"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang-ui/nuklear/nk"
	"github.com/jakecoffman/tanklets"
)

var ctx *nk.Context

func GuiInit(win *glfw.Window) *nk.UserFont {
	ctx = nk.NkPlatformInit(win, nk.PlatformInstallCallbacks)

	atlas := nk.NewFontAtlas()
	nk.NkFontStashBegin(&atlas)
	//sansFont := nk.NkFontAtlasAddFromFile(atlas, "assets/FreeSans.ttf", 16, nil)
	//sansFont := nk.NkFontAtlasAddFromBytes(atlas, MustAsset("assets/FreeSans.ttf"), 16, nil)
	sansFont := nk.NkFontAtlasAddDefault(atlas, 16, nil)
	nk.NkFontStashEnd()
	if sansFont != nil {
		sansFontHandle := sansFont.Handle()
		nk.NkStyleSetFont(ctx, sansFontHandle)
		return sansFontHandle
	}
	return nil
}

func GuiDestroy() {
	nk.NkPlatformShutdown()
}

func GuiRender() {
	nk.NkPlatformNewFrame()

	// Layout
	bounds := nk.NkRect(50, 50, 200, 230)
	update := nk.NkBegin(ctx, "Debug", bounds,
		nk.WindowBorder|nk.WindowMovable|nk.WindowScalable|nk.WindowMinimizable|nk.WindowTitle)

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
		//nk.NkLayoutRowStatic(ctx, 30, 80, 1)
		//{
		//	if nk.NkButtonLabel(ctx, "button") > 0 {
		//		log.Println("[INFO] button pressed!")
		//	}
		//}
		//nk.NkLayoutRowDynamic(ctx, 30, 2)
		//{
		//	if nk.NkOptionLabel(ctx, "easy", flag(state.Opt == Easy)) > 0 {
		//		state.Opt = Easy
		//	}
		//	if nk.NkOptionLabel(ctx, "hard", flag(state.Opt == Hard)) > 0 {
		//		state.Opt = Hard
		//	}
		//}
		//nk.NkLayoutRowDynamic(ctx, 25, 1)
		//{
		//	nk.NkPropertyInt(ctx, "Compression:", 0, &state.Prop, 100, 10, 1)
		//}
		//nk.NkLayoutRowDynamic(ctx, 20, 1)
		//{
		//	nk.NkLabel(ctx, "background:", nk.TextLeft)
		//}
		//nk.NkLayoutRowDynamic(ctx, 25, 1)
		//{
		//	size := nk.NkVec2(nk.NkWidgetWidth(ctx), 400)
		//	if nk.NkComboBeginColor(ctx, state.BgColor, size) > 0 {
		//		nk.NkLayoutRowDynamic(ctx, 120, 1)
		//		state.BgColor = nk.NkColorPicker(ctx, state.BgColor, nk.ColorFormatRGBA)
		//		nk.NkLayoutRowDynamic(ctx, 25, 1)
		//		r, g, b, a := state.BgColor.RGBAi()
		//		r = nk.NkPropertyi(ctx, "#R:", 0, r, 255, 1, 1)
		//		g = nk.NkPropertyi(ctx, "#G:", 0, g, 255, 1, 1)
		//		b = nk.NkPropertyi(ctx, "#B:", 0, b, 255, 1, 1)
		//		a = nk.NkPropertyi(ctx, "#A:", 0, a, 255, 1, 1)
		//		state.BgColor.SetRGBAi(r, g, b, a)
		//		nk.NkComboEnd(ctx)
		//	}
		//}
	}
	nk.NkEnd(ctx)

	// Render
	//bg := make([]float32, 4)
	//nk.NkColorFv(bg, state.BgColor)
	//width, height := win.GetSize()
	//gl.Viewport(0, 0, int32(width), int32(height))
	//gl.Clear(gl.COLOR_BUFFER_BIT)
	//gl.ClearColor(bg[0], bg[1], bg[2], bg[3])
	nk.NkPlatformRender(nk.AntiAliasingOn, maxVertexBuffer, maxElementBuffer)
	//win.SwapBuffers()
}

const (
	maxVertexBuffer  = 512 * 1024
	maxElementBuffer = 128 * 1024
)

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
