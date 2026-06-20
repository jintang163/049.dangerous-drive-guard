package mq

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/apache/rocketmq-clients/golang/v5"
	"github.com/apache/rocketmq-clients/golang/v5/credentials"
	"github.com/dangerous-drive-guard/backend/pkg/config"
	"github.com/dangerous-drive-guard/backend/pkg/logger"
)

var (
	once       sync.Once
	producer   golang.Producer
	TopicMap   map[string]string
)

func Init(cfg *config.RocketMQConfig) error {
	var initErr error
	once.Do(func() {
		TopicMap = cfg.Topics
		provider := golang.NewProvider()
		opts := []golang.Option{
			golang.WithNameServer([]string{cfg.NameServer}),
			golang.WithSessionCredentialsProvider(&credentials.StaticCredentialsProvider{
				Credential: &credentials.SessionCredentials{
					AccessKey: "",
					AccessSecret: "",
				},
			}),
		}
		p, err := provider.NewProducer(opts...)
		if err != nil {
			initErr = fmt.Errorf("create producer: %w", err)
			return
		}
		if err = p.Start(); err != nil {
			initErr = fmt.Errorf("start producer: %w", err)
			return
		}
		producer = p
		logger.Sugar.Infof("RocketMQ producer started: %s", cfg.NameServer)
	})
	return initErr
}

type Message struct {
	Topic string
	Tag   string
	Key   string
	Body  []byte
}

func Send(ctx context.Context, msg Message) error {
	if producer == nil {
		return fmt.Errorf("producer not initialized")
	}
	topic := msg.Topic
	if t, ok := TopicMap[topic]; ok {
		topic = t
	}
	mqMsg := &golang.Message{
		Topic: topic,
		Body:  msg.Body,
	}
	if msg.Tag != "" {
		mqMsg.SetTag(msg.Tag)
	}
	if msg.Key != "" {
		mqMsg.SetKeys([]string{msg.Key})
	}
	mqMsg.SetDelayTimestamp(time.Now())
	sendCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	receipt, err := producer.Send(sendCtx, mqMsg)
	if err != nil {
		return err
	}
	logger.Sugar.Debugf("RocketMQ send ok: topic=%s, msgId=%s", topic, receipt.MessageID)
	return nil
}

type ConsumerHandler func(ctx context.Context, msg *golang.MessageView) error

func StartConsumer(cfg *config.RocketMQConfig, topicKey string, handler ConsumerHandler, concurrency int) error {
	provider := golang.NewProvider()
	topic := topicKey
	if t, ok := TopicMap[topicKey]; ok {
		topic = t
	}
	consumer, err := provider.NewSimpleConsumer(
		golang.WithNameServer([]string{cfg.NameServer}),
		golang.WithConsumerGroup(cfg.ConsumerGroup+"_"+topicKey),
		golang.WithSubscriptionExpressions(map[string]*golang.FilterExpression{
			topic: golang.SUB_ALL,
		}),
		golang.WithAwaitDuration(30*time.Second),
	)
	if err != nil {
		return fmt.Errorf("create consumer: %w", err)
	}
	if err = consumer.Start(); err != nil {
		return fmt.Errorf("start consumer: %w", err)
	}
	logger.Sugar.Infof("RocketMQ consumer started: topic=%s, group=%s", topic, cfg.ConsumerGroup+"_"+topicKey)

	go func() {
		for {
			msgs, err := consumer.Receive(context.Background(), 32, 30*time.Second)
			if err != nil {
				if err != golang.ErrNoNewMessage {
					logger.Sugar.Errorf("RocketMQ receive error: %v", err)
					time.Sleep(time.Second)
				}
				continue
			}
			for _, mv := range msgs {
				mv := mv
				go func() {
					if err := handler(context.Background(), mv); err != nil {
						logger.Sugar.Errorf("RocketMQ handle error: msgId=%s, err=%v", mv.MessageID, err)
						return
					}
					if err := consumer.Ack(context.Background(), mv); err != nil {
						logger.Sugar.Errorf("RocketMQ ack error: msgId=%s, err=%v", mv.MessageID, err)
					}
				}()
			}
		}
	}()
	return nil
}

func Shutdown() {
	if producer != nil {
		_ = producer.GracefulStop()
	}
}
