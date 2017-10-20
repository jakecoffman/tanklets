package main

import (
	"io"
	"log"
	"os"
	"os/exec"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	serverCmd := exec.Command("go", "run", "cmd/tankserv/tankserv.go")
	stderr, err := serverCmd.StderrPipe()
	if err != nil {
		log.Println(err)
	}
	if err = serverCmd.Start(); err != nil {
		log.Println(err)
		return
	}
	go copyC(stderr)

	//time.Sleep(1*time.Second)

	game1Cmd := exec.Command("go", "run", "cmd/tanklets/tanklets.go")
	stderr2, err := game1Cmd.StderrPipe()
	if err != nil {
		log.Println(err)
	}
	if err = game1Cmd.Start(); err != nil {
		log.Println(err)
		return
	}
	go copyC(stderr2)

	game2Cmd := exec.Command("go", "run", "cmd/tanklets/tanklets.go", "650")
	stderr3, err := game2Cmd.StderrPipe()
	if err != nil {
		log.Println(err)
	}
	if err = game2Cmd.Start(); err != nil {
		log.Println(err)
		return
	}
	go copyC(stderr3)

	game1Cmd.Wait()
	game2Cmd.Wait()
	serverCmd.Wait()
}

func copyC(a io.Reader) {
	io.Copy(os.Stderr, a)
}
