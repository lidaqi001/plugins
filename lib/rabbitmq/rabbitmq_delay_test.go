package rabbitmq

import (
	"fmt"
	"github.com/lidaqi001/plugins/lib/rabbitmq/broker"
	"github.com/golang-module/carbon"
	"testing"
)

/**
延时队列
*/

// 生产者
func TestDelayProduce(t *testing.T) {
	b := newBroker()
	err := b.Publish(
		"test_queue_2",
		&broker.Message{
			Header: nil,
			Body:   []byte("test message"),
		},
		// 延时10秒
		Expiration("10000"),
	)
	fmt.Println("publish message:", err)
}

// 消费者1
func TestDelayConsume1(t *testing.T) {
	consumeDelay(QUEUE, "TestConsume1:")
	select {}
}

// 消费者2
func TestDelayConsume2(t *testing.T) {
	consumeDelay("test_queue_2", "TestConsume2:")
	select {}
}

func consumeDelay(queue, flag string) {
	b := newBroker()
	_, err := b.Subscribe(
		queue,
		func(e broker.Event) error {
			fmt.Println(flag, carbon.Now().ToDateTimeString(), ", msg:", e.Message(), ",")
			return nil
		},
		// 持久化队列
		DurableQueue(),
		DelayQueue(),
		broker.Queue(queue),
	)
	fmt.Println(err)
}
