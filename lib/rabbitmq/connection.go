package rabbitmq

//
// All credit to Mondo
//

import (
	"crypto/tls"
	"fmt"
	myLogger "github.com/lidaqi001/plugins/lib/logger"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

var (
	DefaultExchange = Exchange{
		Name: "micro",
		Kind: amqp.ExchangeTopic,
	}
	DefaultRabbitURL      = "amqp://guest:guest@127.0.0.1:5672"
	DefaultPrefetchCount  = 0
	DefaultPrefetchGlobal = false
	DefaultRequeueOnError = false

	// DefaultDelayExchange 延时队列交换机
	/**	@auth liqi lidaqi962464@qq.com	*/
	DefaultDelayExchange = Exchange{
		Name: "micro.delay",
		Kind: amqp.ExchangeTopic,
	}
	DefaultDelayQueue = DefaultDelayExchange.Name
	DelayPrefix       = DefaultDelayExchange.Name + "."

	// The amqp library does not seem to set these when using amqp.DialConfig
	// (even though it says so in the comments) so we set them manually to make
	// sure to not brake any existing functionality
	defaultHeartbeat = 10 * time.Second
	defaultLocale    = "en_US"

	defaultAmqpConfig = amqp.Config{
		Heartbeat: defaultHeartbeat,
		Locale:    defaultLocale,
	}

	dial       = amqp.Dial
	dialTLS    = amqp.DialTLS
	dialConfig = amqp.DialConfig
)

type rabbitMQConn struct {
	Connection      *amqp.Connection
	Channel         *rabbitMQChannel
	ExchangeChannel *rabbitMQChannel
	exchange        Exchange
	url             string
	prefetchCount   int
	prefetchGlobal  bool

	sync.Mutex
	connected bool
	close     chan bool

	waitConnection chan struct{}
}

// Exchange is the rabbitmq exchange
type Exchange struct {
	// Name of the exchange
	Name string
	// Whether its persistent
	Durable bool
	// Kind of the exchange
	Kind string
}

func newRabbitMQConn(ex Exchange, urls []string, prefetchCount int, prefetchGlobal bool) *rabbitMQConn {
	var url string

	if len(urls) > 0 && regexp.MustCompile("^amqp(s)?://.*").MatchString(urls[0]) {
		url = urls[0]
	} else {
		url = DefaultRabbitURL
	}

	ret := &rabbitMQConn{
		exchange:       ex,
		url:            url,
		prefetchCount:  prefetchCount,
		prefetchGlobal: prefetchGlobal,
		close:          make(chan bool),
		waitConnection: make(chan struct{}),
	}
	// its bad case of nil == waitConnection, so close it at start
	close(ret.waitConnection)
	return ret
}

func (r *rabbitMQConn) connect(secure bool, config *amqp.Config) error {
	// try connect
	if err := r.tryConnect(secure, config); err != nil {
		return err
	}

	// connected
	r.Lock()
	r.connected = true
	r.Unlock()

	// create reconnect loop
	go r.reconnect(secure, config)
	return nil
}

func (r *rabbitMQConn) reconnect(secure bool, config *amqp.Config) {
	// skip first connect
	var connect bool

	for {
		if connect {
			// try reconnect
			if err := r.tryConnect(secure, config); err != nil {
				time.Sleep(1 * time.Second)
				continue
			}

			// connected
			r.Lock()
			r.connected = true
			r.Unlock()
			//unblock resubscribe cycle - close channel
			//at this point channel is created and unclosed - close it without any additional checks
			close(r.waitConnection)
		}

		connect = true
		notifyClose := make(chan *amqp.Error)
		r.Connection.NotifyClose(notifyClose)
		chanNotifyClose := make(chan *amqp.Error)
		channel := r.ExchangeChannel.channel
		channel.NotifyClose(chanNotifyClose)
		// To avoid deadlocks it is necessary to consume the messages from all channels.
		for notifyClose != nil || chanNotifyClose != nil {
			// block until closed
			select {
			case err := <-chanNotifyClose:
				//if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
				myLogger.Error(err)
				errLog.Print(err)
				//}
				// block all resubscribe attempt - they are useless because there is no connection to rabbitmq
				// create channel 'waitConnection' (at this point channel is nil or closed, create it without unnecessary checks)
				r.Lock()
				r.connected = false
				r.waitConnection = make(chan struct{})
				r.Unlock()
				chanNotifyClose = nil
			case err := <-notifyClose:
				//if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
				myLogger.Error(err)
				errLog.Print(err)
				//}
				// block all resubscribe attempt - they are useless because there is no connection to rabbitmq
				// create channel 'waitConnection' (at this point channel is nil or closed, create it without unnecessary checks)
				r.Lock()
				r.connected = false
				r.waitConnection = make(chan struct{})
				r.Unlock()
				notifyClose = nil
			case <-r.close:
				return
			}
		}
	}
}

func (r *rabbitMQConn) Connect(secure bool, config *amqp.Config) error {
	r.Lock()

	// already connected
	if r.connected {
		r.Unlock()
		return nil
	}

	// check it was closed
	select {
	case <-r.close:
		r.close = make(chan bool)
	default:
		// no op
		// new conn
	}

	r.Unlock()

	return r.connect(secure, config)
}

func (r *rabbitMQConn) Close() error {
	r.Lock()
	defer r.Unlock()

	select {
	case <-r.close:
		return nil
	default:
		close(r.close)
		r.connected = false
	}

	return r.Connection.Close()
}

func (r *rabbitMQConn) tryConnect(secure bool, config *amqp.Config) error {
	var err error

	if config == nil {
		config = &defaultAmqpConfig
	}

	url := r.url

	if secure || config.TLSClientConfig != nil || strings.HasPrefix(r.url, "amqps://") {
		if config.TLSClientConfig == nil {
			config.TLSClientConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}

		url = strings.Replace(r.url, "amqp://", "amqps://", 1)
	}

	r.Connection, err = dialConfig(url, *config)

	if err != nil {
		return err
	}

	if r.Channel, err = newRabbitChannel(r.Connection, r.prefetchCount, r.prefetchGlobal); err != nil {
		return err
	}

	if r.exchange.Durable {
		r.Channel.DeclareDurableExchange(r.exchange)
	} else {
		r.Channel.DeclareExchange(r.exchange)
	}
	if r.ExchangeChannel, err = newRabbitChannel(r.Connection, r.prefetchCount, r.prefetchGlobal); err != nil {
		return err
	}

	// 声明延迟队列相关
	/**	@auth liqi lidaqi962464@qq.com	*/
	err = r.declareDelay()

	return err
}

func (r *rabbitMQConn) Consume(queue, key string, headers amqp.Table, qArgs amqp.Table, autoAck, durableQueue bool) (*rabbitMQChannel, <-chan amqp.Delivery, error) {

	// if kind equal 'fanout' and queue name is empty,
	// then queue name equal [exchange name] + [uuid]
	if r.exchange.Kind == amqp.ExchangeFanout {
		if queue == "" {
			queue = r.Channel.uuid
		}
		queue = fmt.Sprint("fanout:", r.exchange.Name, ":", queue)
	}

	// 延时队列
	/**	@auth liqi lidaqi962464@qq.com	*/
	if r.exchange == DefaultDelayExchange {
		// 延时队列前缀+队列名称
		queue = DelayPrefix + queue
		key = DelayPrefix + key
	}

	consumerChannel, err := newRabbitChannel(r.Connection, r.prefetchCount, r.prefetchGlobal)
	if err != nil {
		return nil, nil, err
	}

	if durableQueue {
		err = consumerChannel.DeclareDurableQueue(queue, qArgs)
	} else {
		err = consumerChannel.DeclareQueue(queue, qArgs)
	}

	if err != nil {
		return nil, nil, err
	}

	deliveries, err := consumerChannel.ConsumeQueue(queue, autoAck)
	if err != nil {
		return nil, nil, err
	}

	err = consumerChannel.BindQueue(queue, key, r.exchange.Name, headers)
	if err != nil {
		return nil, nil, err
	}

	return consumerChannel, deliveries, nil
}

func (r *rabbitMQConn) Publish(exchange, key string, msg amqp.Publishing) error {
	// 延时消息
	// 绑定普通交换机和延迟队列
	/**	@auth liqi lidaqi962464@qq.com	*/
	if msg.Expiration != "" {
		key = DelayPrefix + key
		if err := r.ExchangeChannel.BindQueue(DefaultDelayQueue, key, r.exchange.Name, amqp.Table{}); err != nil {
			return err
		}
	}
	return r.ExchangeChannel.Publish(exchange, key, msg)
}

// 声明延迟消息相关
/**	@auth liqi lidaqi962464@qq.com	*/
func (r *rabbitMQConn) declareDelay() error {
	var err error

	// 声明延时交换机
	delayerChannel, err := newRabbitChannel(r.Connection, r.prefetchCount, r.prefetchGlobal)
	if err != nil {
		return err
	}
	err = delayerChannel.DeclareDurableExchange(DefaultDelayExchange)
	if err != nil {
		return err
	}

	// 声明延时队列
	// 并将死信消息设置投递至延时交换机
	err = r.ExchangeChannel.DeclareDurableQueue(DefaultDelayQueue, amqp.Table{
		// 当消息过期时把消息发送到 DefaultDelayExchange 这个交换机
		"x-dead-letter-exchange": DefaultDelayExchange.Name,
	})

	return err
}
