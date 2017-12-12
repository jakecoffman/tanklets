package client

type Scene interface {
	Update(dt float64)
	Render()
	Transition() Scene
	Destroy()
}