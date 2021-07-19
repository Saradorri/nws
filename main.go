package main

import (
	"nws/internal/v1/controller/rbm"
	"nws/internal/v1/controller/ws"
)

func main() {
	go ws.Index()
	rbm.Get()
}
