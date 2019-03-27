package services

import (
	"fmt"
	"log"
	"time"

	"github.com/optiopay/kafka"
	"github.com/optiopay/kafka/proto"
)

const (
	partition = 0
)

type MessagingService struct {
	broker     kafka.Client
	kafkaAddrs []string
	topicName  *string
}

func (ms *MessagingService) Init(hosts []string, brokerConfigName string, topicName string) {

	//ms.kafkaAddrs = []string{"localhost:9092", "localhost:9093"}
	//conf := kafka.NewBrokerConf("test-client")
	ms.kafkaAddrs = hosts
	conf := kafka.NewBrokerConf(brokerConfigName)
	ms.topicName = &topicName

	conf.AllowTopicCreation = true

	// connect to kafka cluster
	broker, err := kafka.Dial(ms.kafkaAddrs, conf)
	if err != nil {
		log.Fatalf("cannot connect to kafka cluster: %s", err)
	}

	ms.broker = broker

	defer broker.Close()
}

func (ms *MessagingService) Subscribe() {

	conf := kafka.NewConsumerConf(*ms.topicName, partition)
	conf.StartOffset = kafka.StartOffsetNewest
	consumer, err := ms.broker.Consumer(conf)
	if err != nil {
		log.Fatalf("cannot create kafka consumer for %s:%d: %s", ms.topicName, partition, err)
	}

	for {
		msg, err := consumer.Consume()
		if err != nil {
			if err != kafka.ErrNoData {
				log.Printf("cannot consume %q topic message: %s", ms.topicName, err)
			}
			break
		}
		log.Printf("message %d: %s", msg.Offset, msg.Value)
	}
	log.Print("consumer quit")
}

func (ms *MessagingService) Send(message string) {

	fmt.Println("will be sending out messagess...." + time.Now().String())
	time.Sleep(2000)
	producer := ms.broker.Producer(kafka.NewProducerConf())

	msg := &proto.Message{Value: []byte("hello world")}
	if _, err := producer.Produce(*ms.topicName, partition, msg); err != nil {
		log.Fatalf("cannot produce message to %s:%d: %s", ms.topicName, partition, err)
	}
}
