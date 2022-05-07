package restarter

import (
	"bytes"
	"fmt"
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
				fmt.Printf("ERR: %s", err.Error())
			}

			//TODO delete
			fmt.Printf("OUTPUT: %s", stdout.Bytes())
		}
	}
}

func (r Restart) Restart(filename string) {
	//removeFile(filename)
	if r.cmd != nil {
		r.cmd.Process.Kill()
	}

	r.c <- filename
}

func removeFile(fileName string) {
	if fileName != "" {
		cmd := exec.Command("rm", fileName)
		cmd.Run()
		cmd.Wait()
	}
}
