package main

import (
	"github.com/foxtech6/realtime-build-go/restarter"
	"github.com/foxtech6/realtime-build-go/spier"
	"log"
)

func main() {
	s, err := spier.New()
	r := restarter.New()

	if err != nil {
		log.Fatal(err)
	}

	if err = s.Watch("/home/mbavdys/Projects/spyfiles/test"); err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case ev := <-s.Event:
			//cmd := exec.Command("go", "build", "-o", "file")
			//println(cmd.CombinedOutput())
			log.Println(ev.Mask)
			r.Restart("test1")
			//cmd1, _ := exec.Command("./file123").Output()
			//fmt.Printf("OUTPUT: %s", cmd1)
			//cmd1.Process.Kill()

		case err := <-s.Error:
			log.Println("error:", err)
		}
	}
}
