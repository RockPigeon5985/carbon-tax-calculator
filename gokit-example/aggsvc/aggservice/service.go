package aggservice

import (
	"context"
	"github.com/RockPigeon5985/carbon-tax-calculator/types"
	"github.com/go-kit/log"
)

const basePrice = 3.15

type Service interface {
	Aggregate(context.Context, types.Distance) error
	Calculate(context.Context, int) (*types.Invoice, error)
}

type BasicService struct {
	store Storer
}

func newBasicService(store Storer) Service {
	return &BasicService{
		store: store,
	}
}

func (svc *BasicService) Aggregate(ctx context.Context, distance types.Distance) error {
	return svc.store.Insert(distance)
}

func (svc *BasicService) Calculate(ctx context.Context, obuID int) (*types.Invoice, error) {
	dist, err := svc.store.Get(obuID)
	if err != nil {
		return nil, err
	}

	return &types.Invoice{
		OBUID:         obuID,
		TotalDistance: dist,
		TotalAmount:   basePrice * dist,
	}, nil
}

// New will construct a complete microservice with logging and instrumentation middleware
func New(logger log.Logger) Service {
	var (
		store Storer
		svc   Service
	)
	{
		store = NewMemoryStore()
		svc = newBasicService(store)
		svc = newLoggingMiddleware(logger)(svc)
		svc = newInstrumentationMiddleware()(svc)
	}
	return svc
}
