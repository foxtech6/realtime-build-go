package spier

import (
	"errors"
	"golang.org/x/sys/unix"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

type Event struct {
	Mask   uint32
	Cookie uint32
	Name   string
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
	fd, err := syscall.InotifyInit()

	if fd == -1 {
		return nil, os.NewSyscallError("inotify_init", err)
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

func (w *Watcher) Watch(path string) error {
	if w.isClosed {
		return errors.New("inotify instance already closed")
	}

	watchEntry, found := w.watches[path]

	flags := uint32(syscall.IN_ALL_EVENTS)

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

func (w *Watcher) readEvents() {
	var (
		buf [syscall.SizeofInotifyEvent * 4096]byte
		n   int
		err error
	)

	for {
		n, err = syscall.Read(w.fd, buf[0:])

		var done bool
		select {
		case done = <-w.done:
		default:
		}

		if n == 0 || done {
			err := syscall.Close(w.fd)

			if err != nil {
				w.Error <- os.NewSyscallError("close", err)
			}

			close(w.Event)
			close(w.Error)

			return
		}

		if n < 0 {
			w.Error <- os.NewSyscallError("read", err)
			continue
		}

		if n < syscall.SizeofInotifyEvent {
			w.Error <- errors.New("inotify: short read in readEvents()")
			continue
		}

		var offset uint32 = 0
		for offset <= uint32(n-syscall.SizeofInotifyEvent) {
			raw := (*syscall.InotifyEvent)(unsafe.Pointer(&buf[offset]))
			event := Event{
				Mask:   raw.Mask,
				Cookie: raw.Cookie,
				Name:   w.paths[int(raw.Wd)],
			}

			nameLen := raw.Len

			if nameLen > 0 {
				bytes := (*[syscall.PathMax]byte)(unsafe.Pointer(&buf[offset+syscall.SizeofInotifyEvent]))
				event.Name += "/" + strings.TrimRight(string(bytes[0:nameLen]), "\000")
			}

			//:TODO change it
			if event.Mask&unix.IN_MODIFY == unix.IN_MODIFY {
				w.Event <- &event
			}

			offset += syscall.SizeofInotifyEvent + nameLen
		}
	}
}
