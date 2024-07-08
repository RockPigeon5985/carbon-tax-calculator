package main

import (
	"context"
	"encoding/json"
	aggregation "github.com/RockPigeon5985/carbon-tax-calculator/aggregator/client"
	"github.com/RockPigeon5985/carbon-tax-calculator/types"
	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/sirupsen/logrus"
	"time"
)

type KafkaConsumer struct {
	consumer    *kafka.Consumer
	isRunning   bool
	calcService CalculatorServicer
	aggClient   aggregation.Client
}

func NewKafkaConsumer(topic string, svc CalculatorServicer, aggClient aggregation.Client) (*KafkaConsumer, error) {
	c, err := kafka.NewConsumer(&kafka.ConfigMap{
		"bootstrap.servers": "localhost",
		"group.id":          "myGroup",
		"auto.offset.reset": "earliest",
	})

	if err != nil {
		return nil, err
	}

	c.SubscribeTopics([]string{topic}, nil)

	return &KafkaConsumer{
		consumer:    c,
		calcService: svc,
		aggClient:   aggClient,
	}, nil
}

func (c *KafkaConsumer) readMessageLoop() {
	for c.isRunning {
		msg, err := c.consumer.ReadMessage(-1)
		if err != nil {
			logrus.Errorf("kafka consume error %s", err)
			continue
		}

		var data types.OBUData
		if err = json.Unmarshal(msg.Value, &data); err != nil {
			logrus.Errorf("JSON serialization error %s", err)
			logrus.WithFields(logrus.Fields{
				"err":       err,
				"requestID": data.RequestID,
			}).Info()
			continue
		}

		distance, err := c.calcService.CalculateDistance(data)
		if err != nil {
			logrus.Errorf("calculation error %s", err)
			continue
		}

		req := &types.AggregateRequest{
			Value: distance,
			ObuID: int32(data.OBUID),
			Unix:  time.Now().Unix(),
			//RequestID: data.RequestID,
		}
		if err = c.aggClient.Aggregate(context.Background(), req); err != nil {
			logrus.Error("aggregate error:", err)
			continue
		}
	}
}

func (c *KafkaConsumer) Start() {
	logrus.Info("kafka transport started")
	c.isRunning = true
	c.readMessageLoop()
}

func (c *KafkaConsumer) Close() {
	logrus.Info("kafka transport closed")
	c.isRunning = false
}
