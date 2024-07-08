package main

import (
	"github.com/RockPigeon5985/carbon-tax-calculator/gokit-example/aggendpoint"
	"github.com/RockPigeon5985/carbon-tax-calculator/gokit-example/aggservice"
	"github.com/RockPigeon5985/carbon-tax-calculator/gokit-example/aggtransport"
	"github.com/go-kit/kit/log"
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

	var (
		service     = aggservice.New()
		endpoints   = aggendpoint.New(service, logger)
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
