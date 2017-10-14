package main

import (
	"log"
	"time"

	"github.com/jakecoffman/tanklets"
)

const (
	step = 16666666
	stepDuration = step * time.Nanosecond
	serverUpdateRate = 200 * time.Millisecond
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	tanklets.NewGame(800, 600)
	tanklets.IsServer = true
	tanklets.NetInit()
	defer func() { log.Println(tanklets.NetClose()) }()

	tick := time.Tick(stepDuration)
	var ticks int

	update := time.Tick(serverUpdateRate)

	for {
		select {
		case <-tick:
			ticks++
			tanklets.Space.Step(1.0 / 60.0)

			select {
			case <-update:
				for _, tank := range tanklets.Tanks {
					data, err := tank.Location().MarshalBinary()
					if err != nil {
						log.Println(err)
						continue
					}
					tanklets.Send(data, tank.Addr)
				}
			default:
			}

		case incoming := <-tanklets.Incomings:
			if err := incoming.Handler.Handle(incoming.Addr); err != nil {
				log.Fatal(err)
			}
		}
	}
}
