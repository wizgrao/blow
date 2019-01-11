package main

import (
	"net/http"

	"fmt"
	"github.com/gorilla/websocket"
	"io"
	"os"
	"github.com/wizgrao/blow/maps"
	"github.com/wizgrao/blow/cmd/fizzbuzz"
)

type wasmHandler int

var wasm wasmHandler

type newConnectionHandler int

var nch newConnectionHandler
var upgrader = websocket.Upgrader{}

var pool *maps.WorkerPool
var fizzmapper fizzbuzz.FizzMapper
var generator fizzbuzz.FizzGenerator


type gorillaWrapper struct {
	c *websocket.Conn
}

func (g *gorillaWrapper) Receive() ([]byte, error){
	_, p, err := g.c.ReadMessage()
	return p, err
}

func (g *gorillaWrapper) Send(b []byte) error {
	return g.c.WriteMessage(websocket.TextMessage, b)
}

func (newConnectionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err !=nil {
		fmt.Println("Error upgrading http", err)
		return
	}
	connection := &gorillaWrapper{c}
	fmt.Println("New Connection")
	pool.AddWorker(connection)
	select {}
}

func (h wasmHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/wasm")
	fmt.Println("wasm")
	f, _ := os.Open("slave/main.wasm")
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
	pool = maps.NewWorkerPool()
	pool.Register(fizzmapper)
	go func() {
		maps.GeneratorSource(generator, pool).MapDispatch(fizzmapper).MapLocalParallel(&maps.PrintMapper{}, 10).Sink()
	}()
	fmt.Println("asdf")
	http.Handle("/main.wasm", wasm)
	http.Handle("/sock", nch)
	http.Handle("/", http.FileServer(http.Dir("slave/")))
	fmt.Print(http.ListenAndServe(":8090", nil))
}

