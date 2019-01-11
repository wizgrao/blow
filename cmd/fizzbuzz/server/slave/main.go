package main

import (
	"github.com/wizgrao/blow/websocket"
	"github.com/wizgrao/blow/cmd/fizzbuzz"
	"github.com/wizgrao/blow/maps"
	"fmt"
)

func main() {
	sock := websocket.GetSocket("socket")
	var mapper fizzbuzz.FizzMapper
	h := maps.NewHost(sock)
	h.Register(mapper)
	h.Start()
	select {}
	fmt.Println("huh")
}
