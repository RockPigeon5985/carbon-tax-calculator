package main

import (
	aggregation "github.com/RockPigeon5985/carbon-tax-calculator/aggregator/client"
	"log"
)

const (
	kafkaTopic   = "obudata"
	httpEndpoint = "http://127.0.0.1:3000"
	grpcEndpoint = "http://127.0.0.1:3001"
)

func main() {
	var (
		err error
		svc CalculatorServicer
	)
	svc = NewCalculatorService()
	svc = NewLogMiddleware(svc)

	httpClient := aggregation.NewHTTPClient(httpEndpoint)
	/*grpcClient, err := aggregation.NewGRPCClient(grpcEndpoint)
	if err != nil {
		log.Fatal(err)
	}*/

	kafkaConsumer, err := NewKafkaConsumer(kafkaTopic, svc, httpClient)
	if err != nil {
		log.Fatal(err)
	}

	kafkaConsumer.Start()
}
