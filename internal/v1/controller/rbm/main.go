package rbm

import (
	"github.com/streadway/amqp"
	"nws/config"
	"nws/internal/v1/controller/ws"
	"nws/pkg/rbmq"
)

func Get() {
	var rc rbmq.RabbitClient

	for i, value := range config.CNF.Queues {
		if i == len(config.CNF.Queues)-1 {
			rc.Consume(value, Process)
		} else {
			go rc.Consume(value, Process)
		}
	}

}

func Process(m amqp.Delivery, q string) error {

	ws.HandleMessages(q, m.Body)
	return nil
}
