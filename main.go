package main

import (
	"errors"
	"golang.org/x/sys/unix"
	"log"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

const InAllEvents uint32 = syscall.IN_ALL_EVENTS

type Op uint32

const (
	Create Op = 1 << iota
	Write
	Remove
	Rename
	Chmod
)

func main() {
	watcher, err := NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	err = watcher.Watch("/home/mbavdys/Projects/spyfiles/test")
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

func (w *Watcher) Watch(path string) error {
	return w.AddWatch(path, InAllEvents)
}

func (w *Watcher) AddWatch(path string, flags uint32) error {
	if w.isClosed {
		return errors.New("inotify instance already closed")
	}

	watchEntry, found := w.watches[path]
	if found {
		watchEntry.flags |= flags
		flags |= syscall.IN_MASK_ADD
	}
	wd, errno := syscall.InotifyAddWatch(w.fd, path, flags)
	if wd == -1 {
		return &os.PathError{Op: "inotify_add_watch", Path: path, Err: errno}
	}

	if !found {
		w.watches[path] = &watch{wd: uint32(wd), flags: flags}
		w.paths[wd] = path
	}
	return nil
}

type Event struct {
	Mask   uint32
	Cookie uint32
	Name   string
	Name1  [0]uint8
}

type watch struct {
	wd    uint32
	flags uint32
}

type Watcher struct {
	fd       int
	watches  map[string]*watch
	paths    map[int]string
	Error    chan error
	Event    chan *Event
	done     chan bool
	isClosed bool
}

func NewWatcher() (*Watcher, error) {
	fd, errno := syscall.InotifyInit()
	if fd == -1 {
		return nil, os.NewSyscallError("inotify_init", errno)
	}
	w := &Watcher{
		fd:      fd,
		watches: make(map[string]*watch),
		paths:   make(map[int]string),
		Event:   make(chan *Event),
		Error:   make(chan error),
		done:    make(chan bool, 1),
	}

	go w.readEvents()

	return w, nil
}

func (w *Watcher) readEvents() {
	var (
		buf   [syscall.SizeofInotifyEvent * 4096]byte
		n     int
		errno error
	)

	for {
		n, errno = syscall.Read(w.fd, buf[0:])
		var done bool
		select {
		case done = <-w.done:
		default:
		}

		if n == 0 || done {
			errno := syscall.Close(w.fd)
			if errno != nil {
				w.Error <- os.NewSyscallError("close", errno)
			}
			close(w.Event)
			close(w.Error)
			return
		}
		if n < 0 {
			w.Error <- os.NewSyscallError("read", errno)
			continue
		}
		if n < syscall.SizeofInotifyEvent {
			w.Error <- errors.New("inotify: short read in readEvents()")
			continue
		}

		var offset uint32 = 0
		for offset <= uint32(n-syscall.SizeofInotifyEvent) {
			raw := (*syscall.InotifyEvent)(unsafe.Pointer(&buf[offset]))
			//fmt.Println(fmt.Sprintf("%#v", raw))
			event := new(Event)
			event.Mask = raw.Mask
			event.Cookie = raw.Cookie
			event.Name1 = raw.Name
			nameLen := raw.Len

			event.Name = w.paths[int(raw.Wd)]
			if nameLen > 0 {
				bytes := (*[syscall.PathMax]byte)(unsafe.Pointer(&buf[offset+syscall.SizeofInotifyEvent]))
				event.Name += "/" + strings.TrimRight(string(bytes[0:nameLen]), "\000")
			}

			if event.Mask&unix.IN_MODIFY == unix.IN_MODIFY {
				println(event.Mask)

				w.Event <- event
			}
			//if event.Mask == 1<<1 {
			//	//fmt.Println(fmt.Sprintf("%#v", event.Name1))
			//
			//}

			offset += syscall.SizeofInotifyEvent + nameLen
		}
	}
}
