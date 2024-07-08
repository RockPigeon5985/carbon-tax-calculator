package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/RockPigeon5985/carbon-tax-calculator/types"
	"net/http"
)

type HTTPClient struct {
	Endpoint string
}

func NewHTTPClient(endpoint string) Client {
	return &HTTPClient{
		Endpoint: endpoint,
	}
}

func (c *HTTPClient) Aggregate(ctx context.Context, aggReq *types.AggregateRequest) error {
	b, err := json.Marshal(aggReq)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", c.Endpoint+"/aggregate", bytes.NewReader(b))
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("the service responded with a non 200 status code: %d", resp.StatusCode)
	}
	return nil
}

func (c *HTTPClient) GetInvoice(ctx context.Context, id int) (*types.Invoice, error) {
	invReq := types.GetInvoiceRequest{ObuID: int32(id)}

	b, err := json.Marshal(&invReq)
	if err != nil {
		return nil, err
	}
	endpoint := fmt.Sprintf("%s/%s?obu=%d", c.Endpoint, "invoice", id)
	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("the service responded with a non 200 status code: %d", resp.StatusCode)
	}

	var inv types.Invoice
	if err = json.NewDecoder(resp.Body).Decode(&inv); err != nil {
		return nil, err
	}

	return &inv, nil
}
