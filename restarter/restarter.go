package restarter

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
)

type Restart struct {
	c   chan string
	cmd *exec.Cmd
}

func New() Restart {
	r := Restart{
		c: make(chan string),
	}

	go r.Run()

	return r
}

func (r Restart) Run() {
	for {
		select {
		case filename := <-r.c:
			r.cmd = exec.Command(filename)

			//TODO delete with fmt.Printf("OUTPUT: %s", stdout.Bytes())
			var stdout bytes.Buffer
			r.cmd.Stdout = &stdout

			if err := r.cmd.Run(); err != nil {
				fmt.Printf("ERR: %s\n", err.Error())
			}

			//TODO delete
			fmt.Printf("OUTPUT: %s\n", stdout.Bytes())
		}
	}
}

func (r Restart) Restart(filename string) {
	if r.cmd != nil {
		r.cmd.Process.Kill()
	}

	removeFile(filename)
	build(filename)

	r.c <- filename
}

func build(filename string) {
	if err := exec.Command("go", "build", "-o", filename, ".").Run(); err != nil {
		fmt.Printf("ERR build: %s\n", err.Error())
	}
}

func removeFile(fileName string) {
	if ok, _ := existsFile(fileName); !ok {
		fmt.Printf("INF remove: file not exists\n")
		return
	}

	if err := exec.Command("rm", fileName).Run(); err != nil {
		fmt.Printf("ERR remove: %s\n", err.Error())
	}
}

func existsFile(filename string) (bool, error) {
	_, err := os.Stat(filename)

	if os.IsNotExist(err) {
		return false, nil
	}

	return err == nil, err
}
