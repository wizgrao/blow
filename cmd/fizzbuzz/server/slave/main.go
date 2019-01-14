package main

import (
	"github.com/wizgrao/blow/wasmsocket"
	"github.com/wizgrao/blow/cmd/fizzbuzz"
	"github.com/wizgrao/blow/maps"
)

func main() {
	sock := wasmsocket.GetSocket("socket")
	var mapper fizzbuzz.FizzMapper
	h := maps.NewHost(sock)
	h.Register(mapper)
	h.Start()
	select {}
}
