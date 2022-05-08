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

type Spy struct {
	fd       int
	watches  map[string]*watch
	paths    map[int]string
	Error    chan error
	Event    chan *Event
	done     chan bool
	isClosed bool
}

func New() (*Spy, error) {
	fd, err := syscall.InotifyInit()

	if fd == -1 {
		return nil, os.NewSyscallError("inotify_init", err)
	}

	s := &Spy{
		fd:      fd,
		watches: make(map[string]*watch),
		paths:   make(map[int]string),
		Event:   make(chan *Event),
		Error:   make(chan error),
		done:    make(chan bool, 1),
	}

	go s.readEvents()

	return s, nil
}

func (s *Spy) Watch(path string) error {
	if s.isClosed {
		return errors.New("inotify instance already closed")
	}

	watchEntry, found := s.watches[path]

	flags := uint32(syscall.IN_ALL_EVENTS)

	if found {
		watchEntry.flags |= flags
		flags |= syscall.IN_MASK_ADD
	}

	wd, err := syscall.InotifyAddWatch(s.fd, path, flags)

	if wd == -1 {
		return &os.PathError{Op: "inotify_add_watch", Path: path, Err: err}
	}

	if !found {
		s.watches[path] = &watch{wd: uint32(wd), flags: flags}
		s.paths[wd] = path
	}

	return nil
}

func (s *Spy) readEvents() {
	var (
		buf [syscall.SizeofInotifyEvent * 4096]byte
		n   int
		err error
	)

	for {
		n, err = syscall.Read(s.fd, buf[0:])

		var done bool

		select {
		case done = <-s.done:
		default:
		}

		if n == 0 || done {
			err := syscall.Close(s.fd)

			if err != nil {
				s.Error <- os.NewSyscallError("close", err)
			}

			close(s.Event)
			close(s.Error)

			return
		}

		if n < 0 {
			s.Error <- os.NewSyscallError("read", err)
			continue
		}

		if n < syscall.SizeofInotifyEvent {
			s.Error <- errors.New("inotify: short read in readEvents()")
			continue
		}

		var offset uint32 = 0
		for offset <= uint32(n-syscall.SizeofInotifyEvent) {
			raw := (*syscall.InotifyEvent)(unsafe.Pointer(&buf[offset]))
			event := Event{
				Mask:   raw.Mask,
				Cookie: raw.Cookie,
				Name:   s.paths[int(raw.Wd)],
			}

			nameLen := raw.Len

			if nameLen > 0 {
				bytes := (*[syscall.PathMax]byte)(unsafe.Pointer(&buf[offset+syscall.SizeofInotifyEvent]))
				event.Name += "/" + strings.TrimRight(string(bytes[0:nameLen]), "\000")
			}

			//:TODO change it
			if event.Mask&unix.IN_MODIFY == unix.IN_MODIFY {
				println(22)
				s.Event <- &event
			}

			offset += syscall.SizeofInotifyEvent + nameLen
		}
	}
}
