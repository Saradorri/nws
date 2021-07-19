package main

import (
	"nws/pkg/rbmq"
)

func main() {
	var c rbmq.RabbitClient
	b := []byte("no connected!!!!!!!!!!")
	c.Publish("q2", b)
}
