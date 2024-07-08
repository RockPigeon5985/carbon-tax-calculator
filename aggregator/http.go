package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/RockPigeon5985/carbon-tax-calculator/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"time"
)

type HTTPFunc func(w http.ResponseWriter, r *http.Request) error

type APIError struct {
	Code int
	Err  error
}

func (e APIError) Error() string {
	return e.Err.Error()
}

func makeHTTPHandlerFunc(fn HTTPFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			var apiErr APIError
			if errors.As(err, &apiErr) {
				writeJSON(w, apiErr.Code, map[string]string{"error": apiErr.Error()})
			}
		}
	}
}

type HTTPMetricHandler struct {
	reqCounter prometheus.Counter
	errCounter prometheus.Counter
	reqLatency prometheus.Histogram
}

func newHTTPMetricHandler(reqName string) *HTTPMetricHandler {
	reqCounter := promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "http_aggregator_request_counter",
		Name:      reqName,
	})

	errCounter := promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "http_aggregator_error_counter",
		Name:      reqName,
	})

	reqLatency := promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "http_aggregator_request_latency",
		Name:      reqName,
		Buckets:   []float64{0.1, 0.5, 1},
	})

	return &HTTPMetricHandler{
		reqCounter: reqCounter,
		errCounter: errCounter,
		reqLatency: reqLatency,
	}
}

func (h *HTTPMetricHandler) instrument(next HTTPFunc) HTTPFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		h.reqCounter.Inc()
		var err error
		defer func(start time.Time) {
			latency := time.Since(start).Seconds()

			logrus.WithFields(logrus.Fields{
				"latency": latency,
				"request": r.RequestURI,
				"err":     err.Error(),
			}).Info()

			h.reqLatency.Observe(latency)

			if err != nil {
				h.errCounter.Inc()
			}
		}(time.Now())

		err = next(w, r)
		return err
	}
}

func handleAggregate(svc Aggregator) HTTPFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != "POST" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "method not supported"})
			return &APIError{
				Code: http.StatusBadRequest,
				Err:  fmt.Errorf("method %s not supported", r.Method),
			}
		}

		var distance types.Distance
		if err := json.NewDecoder(r.Body).Decode(&distance); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
			return &APIError{
				Code: http.StatusBadRequest,
				Err:  err,
			}
		}

		if err := svc.AggregateDistance(distance); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return &APIError{
				Code: http.StatusBadRequest,
				Err:  err,
			}
		}
		return writeJSON(w, http.StatusOK, map[string]string{"msg": "ok"})
	}
}

func handleGetInvoice(svc Aggregator) HTTPFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		if r.Method != "GET" {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "method not supported"})
			return &APIError{
				Code: http.StatusBadRequest,
				Err:  fmt.Errorf("method %s not supported", r.Method),
			}
		}

		values, ok := r.URL.Query()["obu"]
		if !ok {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "missing OBU ID"})
			return &APIError{
				Code: http.StatusBadRequest,
				Err:  fmt.Errorf("missing OBU ID"),
			}
		}

		obuID, err := strconv.Atoi(values[0])
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid OBU ID"})
			return &APIError{
				Code: http.StatusBadRequest,
				Err:  fmt.Errorf("invalid OBU ID %s", values[0]),
			}
		}

		invoice, err := svc.CalculateInvoice(obuID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
			return &APIError{
				Code: http.StatusInternalServerError,
				Err:  err,
			}
		}
		return writeJSON(w, http.StatusOK, invoice)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.WriteHeader(status)
	w.Header().Add("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}
