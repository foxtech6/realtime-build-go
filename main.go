package main

import (
	"fmt"
	"github.com/foxtech6/realtime-build-go/spier"
	"log"
	"os/exec"
)

func main() {
	watcher, err := spier.NewWatcher()
	runC := make(chan string)
	go run(runC)

	if err != nil {
		log.Fatal(err)
	}

	if err = watcher.Watch("/home/mbavdys/Projects/spyfiles/test"); err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case ev := <-watcher.Event:
			//cmd := exec.Command("go", "build", "-o", "file")
			//println(cmd.CombinedOutput())
			log.Println(ev.Mask)
			runC <- "./file123"
			//cmd1, _ := exec.Command("./file123").Output()
			//fmt.Printf("OUTPUT: %s", cmd1)
			//cmd1.Process.Kill()

		case err := <-watcher.Error:
			log.Println("error:", err)
		}
	}
}

func run(run <-chan string) {
	for {
		select {
		case <-run:
			cmd1, _ := exec.Command("./file123").Output()
			fmt.Printf("OUTPUT: %s", cmd1)
		}
	}
}
