package restarter

import (
	"fmt"
	"os/exec"
)

type Restart struct {
	Chan chan string
}

func New() Restart {
	return Restart{
		Chan: make(chan string),
	}
}

func (r Restart) Run() {
	for {
		select {
		case filename := <-r.Chan:
			cmd, _ := exec.Command(filename).Output()
			fmt.Printf("OUTPUT: %s", cmd)
		}
	}
}

func (r Restart) Restart(filename string) {
	//removeFile(filename)
	r.Chan <- filename
}

func removeFile(fileName string) {
	if fileName != "" {
		cmd := exec.Command("rm", fileName)
		cmd.Run()
		cmd.Wait()
	}
}
