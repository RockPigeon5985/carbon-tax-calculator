package main

import (
	"context"
	"fmt"
	"github.com/RockPigeon5985/carbon-tax-calculator/aggregator/client"
	"github.com/RockPigeon5985/carbon-tax-calculator/types"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}
	var (
		store          = makeStore()
		svc            = NewInvoiceAggregator(store)
		httpListenAddr = os.Getenv("AGG_HTTP_ENDPOINT")
		grpcListenAddr = os.Getenv("AGG_GRPC_ENDPOINT")
	)

	svc = NewMetricsMiddleware(svc)
	svc = NewLogMiddleware(svc)

	go func() {
		log.Fatal(makeGRPCTransport(grpcListenAddr, svc))
	}()

	time.Sleep(2 * time.Second)
	c, err := client.NewGRPCClient(grpcListenAddr)
	if err != nil {
		fmt.Println(err)
	}
	if err = c.Aggregate(context.Background(), &types.AggregateRequest{
		ObuID: 1,
		Value: 22.45,
		Unix:  time.Now().Unix(),
	}); err != nil {
		fmt.Println(err)
	}
	log.Fatal(makeHTTPTransport(httpListenAddr, svc))
}

func makeGRPCTransport(listenAddr string, svc Aggregator) error {
	fmt.Println("GRPC transport running on port", listenAddr)
	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer func(ln net.Listener) {
		err = ln.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(ln)

	server := grpc.NewServer(grpc.EmptyServerOption{})
	types.RegisterAggregatorServer(server, NewGRPCAggregatorServer(svc))
	return server.Serve(ln)
}

func makeHTTPTransport(listenAddr string, svc Aggregator) error {
	fmt.Println("HTTP transport running on port", listenAddr)

	aggMetricHandler := newHTTPMetricHandler("aggregate")
	invMetricHandler := newHTTPMetricHandler("invoice")

	http.HandleFunc("/aggregate", makeHTTPHandlerFunc(aggMetricHandler.instrument(handleAggregate(svc))))
	http.HandleFunc("/invoice", makeHTTPHandlerFunc(invMetricHandler.instrument(handleGetInvoice(svc))))
	http.Handle("/metrics", promhttp.Handler())
	return http.ListenAndServe(listenAddr, nil)
}

func makeStore() Storer {
	storeType := os.Getenv("AGG_STORE_TYPE")
	switch storeType {
	case "memory":
		return NewMemoryStore()
	default:
		log.Fatalf("invalid store type given %s", storeType)
		return nil
	}
}
