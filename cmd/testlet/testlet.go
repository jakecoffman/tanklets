package main

import (
	"io"
	"log"
	"os"
	"os/exec"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	cmd1 := exec.Command("go", "install")
	cmd1.Dir = "cmd/tankserv"
	cmd1.Env = os.Environ()
	start(cmd1)
	if err := cmd1.Wait(); err != nil {
		return
	}

	cmd2 := exec.Command("go", "install")
	cmd2.Dir = "cmd/tanklets"
	cmd2.Env = os.Environ()
	start(cmd2)
	if err := cmd2.Wait(); err != nil {
		return
	}

	//serverCmd := exec.Command("tankserv")
	//serverCmd.Env = os.Environ()
	//start(serverCmd)

	game1Cmd := exec.Command("tanklets")
	game1Cmd.Env = os.Environ()
	start(game1Cmd)
	game2Cmd := exec.Command("tanklets", "650")
	game2Cmd.Env = os.Environ()
	start(game2Cmd)

	game1Cmd.Wait()
	game2Cmd.Wait()
	//serverCmd.Wait()
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
