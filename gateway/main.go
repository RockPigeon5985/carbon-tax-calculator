package main

import (
	"context"
	"encoding/json"
	"flag"
	"github.com/RockPigeon5985/carbon-tax-calculator/aggregator/client"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"time"
)

type apiFunc func(w http.ResponseWriter, r *http.Request) error

type InvoiceHandler struct {
	client client.Client
}

func newInvoiceHandler(c client.Client) *InvoiceHandler {
	return &InvoiceHandler{client: c}
}

func main() {
	listenAddr := flag.String("listenAddr", ":6000", "")
	aggServiceAddr := flag.String("aggServiceAddr", "http://localhost:3000", "")
	flag.Parse()

	c := client.NewHTTPClient(*aggServiceAddr)
	invHandler := newInvoiceHandler(c)

	http.HandleFunc("/invoice", makeAPIFunc(invHandler.handleGetInvoice))

	logrus.Info("gateway running on port ", *listenAddr)
	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}

func (h *InvoiceHandler) handleGetInvoice(w http.ResponseWriter, r *http.Request) error {
	inv, err := h.client.GetInvoice(context.Background(), 423432)
	if err != nil {
		return err
	}

	_ = inv
	return writeJSON(w, http.StatusOK, map[string]string{"invoice": "some invoice"})
}

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.WriteHeader(status)
	w.Header().Add("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(v)
}

func makeAPIFunc(fn apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func(start time.Time) {
			logrus.WithFields(logrus.Fields{
				"took": time.Since(start),
				"uri":  r.RequestURI,
			}).Info("REQ")
		}(time.Now())

		if err := fn(w, r); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	}
}
