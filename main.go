package main

import (
	"github.com/foxtech6/realtime-build-go/spier"
	"log"
)

func main() {
	watcher, err := spier.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	err = watcher.Watch("/home/mbavdys/Projects/spyfiles/test")
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case ev := <-watcher.Event:
			log.Println(ev.Mask)
		case err := <-watcher.Error:
			log.Println("error:", err)
		}
	}
}
