package wasmsocket

import (
	"errors"
	"sync"
	"syscall/js"
	"fmt"
	"encoding/base64"
)

type Socket struct {
	jsSocket    js.Value
	receiveChan chan []byte
	closeChan   chan struct{}
	active      bool
	sync.Mutex
}

func GetSocket(name string) *Socket {
	sock := js.Global().Get(name)
	fmt.Println(sock)
	ch := make(chan []byte, 100)
	sock.Set("onmessage", js.NewCallback(func(args []js.Value) {
		input := args[0].Get("data").String()
		decoded, err := base64.StdEncoding.DecodeString(input)
		if err != nil {
			fmt.Println("error decoding", err)
			return
		}
		ch <- decoded
	}))

	s := &Socket{
		jsSocket:    sock,
		receiveChan: ch,
		closeChan:   make(chan struct{}, 100),
		active:      true,
	}
	fmt.Println(s.jsSocket)

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
		return msg, nil
	}
}

func (s *Socket) Send(msg []byte) error {
	s.Lock()
	if !s.active {
		s.Unlock()
		return errors.New("Socket Closed")
	}
	s.Unlock()
	str := base64.StdEncoding.EncodeToString(msg)
	s.jsSocket.Call("send", str)
	return nil
}



