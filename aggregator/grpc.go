package main

import (
	"context"
	"github.com/RockPigeon5985/carbon-tax-calculator/types"
)

type GRPCAggregatorServer struct {
	types.UnimplementedAggregatorServer
	svc Aggregator
}

func NewGRPCAggregatorServer(svc Aggregator) *GRPCAggregatorServer {
	return &GRPCAggregatorServer{
		svc: svc,
	}
}

func (s *GRPCAggregatorServer) Aggregate(ctx context.Context, req *types.AggregateRequest) (*types.None, error) {
	distance := types.Distance{
		Value: req.Value,
		OBUID: int(req.ObuID),
		Unix:  req.Unix,
	}

	return &types.None{}, s.svc.AggregateDistance(distance)
}
