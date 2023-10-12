package client

import (
	"github.com/lidaqi001/plugins/lib/rabbitmq"
	"github.com/lidaqi001/plugins/lib/rabbitmq/broker"
)

type Sub struct {
	Exchange string
	mq       broker.Broker
}

func GetSub(addr, exchange string, opt ...broker.Option) (*Sub, error) {
	m, err := NewClient(addr, exchange, opt...)
	return &Sub{
		Exchange: exchange,
		mq:       m,
	}, err
}

func (s *Sub) Subscribe(queue string, call func(e broker.Event) error, opts ...broker.SubscribeOption) (err error) {
	option := []broker.SubscribeOption{
		// 持久化队列
		rabbitmq.DurableQueue(),
		// 设置队列名称
		broker.Queue(queue),
	}
	for _, opt := range opts {
		option = append(option, opt)
	}
	_, err = s.mq.Subscribe(
		queue,
		call,
		option...,
	)

	return
}

func (s *Sub) SubscribeFanout(queue string, call func(e broker.Event) error) (err error) {
	_, err = s.mq.SubscribeFanout(
		call,
		// 持久化队列
		rabbitmq.DurableQueue(),
		broker.Queue(queue),
	)

	return
}

// 关闭连接
func (s *Sub) Close() {
	s.mq.Disconnect()
}

// 连接状态
func (s *Sub) Connected() (err error) {
	err = s.mq.Connect()
	return
}
