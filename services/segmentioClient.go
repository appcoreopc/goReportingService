package services

import (
	"context"
	"fmt"

	kafka "github.com/segmentio/kafka-go"
)

const (
	partitionlocal = 0
)

type SegmentIOService struct {
	kafkaAddrs []string
	topicName  *string
	writer     *kafka.Writer
}

func (ss *SegmentIOService) Init(hosts []string, topicName string) {

	ss.kafkaAddrs = hosts
	ss.topicName = &topicName

	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  ss.kafkaAddrs,
		Topic:    *ss.topicName,
		Balancer: &kafka.LeastBytes{},
	})
	ss.writer = w
}

func (ss *SegmentIOService) Subscribe() {

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   ss.kafkaAddrs,
		Topic:     *ss.topicName,
		Partition: 0,
		MinBytes:  10e3, // 10KB
		MaxBytes:  10e6, // 10MB
	})

	reader.SetOffset(0)

	for {

		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			break
		}
		fmt.Printf("message at offset %d: %s = %s\n", m.Offset, string(m.Key), string(m.Value))
	}

	defer reader.Close()
}

func (ss *SegmentIOService) Send(message string) {

	ss.writer.WriteMessages(context.Background(), kafka.Message{Value: []byte(message)})

}
