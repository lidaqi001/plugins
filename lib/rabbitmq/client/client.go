package client

import (
	"fmt"
	mq "github.com/lidaqi001/plugins/lib/rabbitmq"
	"github.com/lidaqi001/plugins/lib/rabbitmq/broker"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	// Pool 连接池
	Pool map[string]broker.Broker
	// TryCount 重试次数
	TryCount int
)

//var errLog *log.Logger

func init() {
	TryCount = 5
	Pool = make(map[string]broker.Broker)
	// 初始化日志
	//errLog = myLogger.ErrLogger("[rabbitmq.client]")
}

func NewClient(addr, exchange string, opt ...broker.Option) (cli broker.Broker, err error) {
	key := fmt.Sprintf("%v_%v", addr, exchange)
	if item, ok := Pool[key]; ok {
		// pool
		cli = item

	} else {
		// 新建
		cli, err = newClient(addr, exchange, opt...)
	}

	for TryCount > 0 {
		err = cli.Connect()
		if err != nil {
			log.Printf("[rabbitmq.client] fail to connect rabbitmq: %v, error: %v \n", addr, err)
			// 重新创建客户端
			cli, err = newClient(addr, exchange, opt...)
			if err != nil {
				log.Printf("[rabbitmq.client] renew client err:%v - %v \n", addr, err)
			}
			TryCount--
			continue
		}
		err = nil
		Pool[key] = cli
		break
	}
	return
}

func newClient(addr, exchange string, opt ...broker.Option) (cli broker.Broker, err error) {
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
	for _, o := range opt {
		opts = append(opts, o)
	}

	// 设置交换器为fanout
	//opts = append(opts, mq.FanoutExchange())

	// 创建连接
	cli = mq.NewBroker(opts...)
	err = cli.Init()
	return
}

// WaitForSignals 监听信号
func WaitForSignals(shutFunc func()) error {
	println("Send signal TERM or INT to terminate the process")
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGTERM, syscall.SIGINT)
	<-sigs

	//执行回调
	shutFunc()

	return nil
}
