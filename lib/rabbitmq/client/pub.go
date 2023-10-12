package client

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/lidaqi001/plugins/lib/rabbitmq/broker"
)

type Pub struct {
	Exchange string
	mq       broker.Broker
}

func GetPub(addr, exchange string, opt ...broker.Option) (*Pub, error) {
	m, err := NewClient(addr, exchange, opt...)
	return &Pub{
		Exchange: exchange,
		mq:       m,
	}, err
}

func (p *Pub) Publish(topic string, body interface{}) (err error) {
	return p.publish(body, func(pub *Pub, message *broker.Message) error {
		return pub.mq.Publish(topic, message)
	})
}

func (p *Pub) PublishFanout(body interface{}) (err error) {
	return p.publish(body, func(pub *Pub, message *broker.Message) error {
		return pub.mq.PublishFanout(message)
	})
}

func (p *Pub) publish(body interface{}, call func(pub *Pub, message *broker.Message) error) (err error) {
	var js []byte
	if js, err = jsoniter.Marshal(body); err != nil {
		return
	}
	message := &broker.Message{
		Header: nil,
		Body:   js,
	}
	return call(p, message)
}

func (p *Pub) Close() {
	p.mq.Disconnect()
}
