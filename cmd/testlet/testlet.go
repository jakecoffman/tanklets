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
	start(serverCmd)

	game1Cmd := exec.Command("go", "run", "cmd/tanklets/tanklets.go")
	start(game1Cmd)

	game2Cmd := exec.Command("go", "run", "cmd/tanklets/tanklets.go", "650")
	start(game2Cmd)

	game1Cmd.Wait()
	game2Cmd.Wait()
	serverCmd.Wait()
}

func start(cmd *exec.Cmd) {
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}
	if err = cmd.Start(); err != nil {
		log.Fatal(err)
		return
	}
	go copyOut(stdout)
	go copyErr(stderr)
}

func copyOut(a io.Reader) {
	io.Copy(os.Stdout, a)
}

func copyErr(a io.Reader) {
	io.Copy(os.Stderr, a)
}
