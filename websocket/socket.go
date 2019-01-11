package websocket

import (
	"errors"
	"sync"
	"syscall/js"
)

type Socket struct {
	jsSocket    js.Value
	receiveChan chan string
	closeChan   chan struct{}
	active      bool
	sync.Mutex
}

func GetSocket(name string) *Socket {
	sock := js.Global().Get(name)
	ch := make(chan string, 100)
	sock.Set("onmessage", js.NewCallback(func(args []js.Value) {
		ch <- args[0].Get("data").String()
	}))

	s := &Socket{
		jsSocket:    sock,
		receiveChan: ch,
		closeChan:   make(chan struct{}, 100),
		active:      true,
	}

	sock.Set("onclose", js.NewCallback(func(args []js.Value) {
		s.closeChan <- struct{}{}
		s.Lock()
		s.active = false
		s.Unlock()
	}))
	return s
}

func (s *Socket) Receive() ([]byte, error) {
	s.Lock()
	if !s.active {
		s.Unlock()
		return []byte{}, errors.New("Socket Closed")
	}
	s.Unlock()
	select {
	case <-s.closeChan:
		return []byte{}, errors.New("Socket Closed")
	case msg := <-s.receiveChan:
		return []byte(msg), nil
	}
}

func (s *Socket) Send(msg []byte) error {
	s.Lock()
	if !s.active {
		s.Unlock()
		return errors.New("Socket Closed")
	}
	s.Unlock()
	s.jsSocket.Call("send", string(msg))
	return nil
}



