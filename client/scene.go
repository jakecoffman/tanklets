package client

import "github.com/golang-ui/nuklear/nk"

type Scene interface {
	Update(dt float64)
	Render(ctx *nk.Context)
	Transition() Scene
	Destroy()
}
