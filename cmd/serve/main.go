package main

import (
	"net/http"

	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"os"
)

type handle int

var hand handle

type echo int

var ech echo
var upgrader = websocket.Upgrader{}

func (echo) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Upgrade error", err)
		return
	}
	defer c.Close()

	for {
		mt, msg, err := c.ReadMessage()

		if err != nil {
			fmt.Println("Connection closed", err)
			return
		}
		err = c.WriteMessage(mt, msg)
		if err != nil {
			fmt.Println("Writeerr", err)
			return
		}
		fmt.Println(string(msg))
	}

}

func (h handle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/wasm")
	fmt.Println("wasm")
	f, _ := os.Open("main.wasm")
	p := make([]byte, 4)
	for {
		n, err := f.Read(p)
		if err == io.EOF {
			break
		}
		w.Write(p[:n])
	}

}

func main() {
	fmt.Println("asdf")
	http.Handle("/main.wasm", hand)
	http.Handle("/sock", ech)
	http.Handle("/", http.FileServer(http.Dir(".")))
	fmt.Print(http.ListenAndServe(":8080", nil))
}
