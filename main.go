package main

import (
	"github.com/foxtech6/realtime-build-go/restarter"
	"github.com/foxtech6/realtime-build-go/spier"
	"log"
)

func main() {
	watcher, err := spier.NewWatcher()
	r := restarter.New()
	go r.Run()

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
			r.Restart("./file123")
			//cmd1, _ := exec.Command("./file123").Output()
			//fmt.Printf("OUTPUT: %s", cmd1)
			//cmd1.Process.Kill()

		case err := <-watcher.Error:
			log.Println("error:", err)
		}
	}
}
