package main

import (
	"github.com/RockPigeon5985/carbon-tax-calculator/gokit-example/aggsvc/aggendpoint"
	"github.com/RockPigeon5985/carbon-tax-calculator/gokit-example/aggsvc/aggservice"
	"github.com/RockPigeon5985/carbon-tax-calculator/gokit-example/aggsvc/aggtransport"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net"
	"net/http"
	"os"
)

func main() {
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	var duration metrics.Histogram
	{
		duration = prometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "carbon_tax_calculator",
			Subsystem: "aggsvc",
			Name:      "request_duration_seconds",
			Help:      "Request duration in seconds.",
		}, []string{"method", "success"})
	}
	http.DefaultServeMux.Handle("/metrics", promhttp.Handler())

	var (
		service     = aggservice.New(logger)
		endpoints   = aggendpoint.New(service, duration, logger)
		httpHandler = aggtransport.NewHTTPHandler(endpoints, logger)
	)

	httpListener, err := net.Listen("tcp", ":4000")
	if err != nil {
		logger.Log("transport", "HTTP", "during", "Listen", "err", err)
		os.Exit(1)
	}

	logger.Log("transport", "HTTP", "addr", ":4000")
	err = http.Serve(httpListener, httpHandler)
	if err != nil {
		panic(err)
	}
}
