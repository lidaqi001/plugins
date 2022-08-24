package client

import (
	"fmt"
	myLogger "github.com/lidaqi001/plugins/lib/logger"
	mq "github.com/lidaqi001/plugins/lib/rabbitmq"
	"github.com/lidaqi001/plugins/lib/rabbitmq/broker"
	"log"
)

var (
	// Pool 连接池
	Pool map[string]broker.Broker
	// TryCount 重试次数
	TryCount int
)

var errLog *log.Logger

func init() {
	TryCount = 5
	Pool = make(map[string]broker.Broker)
	// 初始化日志
	errLog = myLogger.ErrLogger("[rabbitmq.client]")
}

func NewClient(addr, exchange string, fanout bool) (cli broker.Broker, err error) {
	key := fmt.Sprintf("%v_%v_%v", addr, exchange, fanout)
	if item, ok := Pool[key]; ok {
		// pool
		cli = item

	} else {
		// 新建
		cli = newClient(addr, exchange, fanout)
	}

	for TryCount > 0 {
		err = cli.Connect()
		if err != nil {
			// 重新创建客户端
			cli = newClient(addr, exchange, fanout)
			errLog.Printf("fail to connect rabbitmq: %v, error: %v", addr, err)
			TryCount--
			continue
		}
		err = nil
		Pool[key] = cli
		break
	}
	return
}

func newClient(addr, exchange string, fanout bool) (cli broker.Broker) {
	// 设置rabbitmq服务器地址
	mq.DefaultRabbitURL = addr

	// 设置参数
	opts := []broker.Option{
		// 持久化exchange
		mq.DurableExchange(),
		// 订阅时创建一个持久化队列
		mq.PrefetchGlobal(),
		// 设置交换器名称，不存在的会新建
		mq.ExchangeName(exchange),
	}
	if fanout {
		// 设置交换器为fanout
		opts = append(opts, mq.FanoutExchange())
	}

	// 创建连接
	cli = mq.NewBroker(opts...)
	cli.Init()
	return
}
