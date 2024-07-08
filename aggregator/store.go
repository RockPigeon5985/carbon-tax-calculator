package main

import (
	"fmt"
	"github.com/RockPigeon5985/carbon-tax-calculator/types"
)

type Storer interface {
	Insert(types.Distance) error
	Get(int) (float64, error)
}

type MemoryStore struct {
	data map[int]float64
}

func NewMemoryStore() Storer {
	return &MemoryStore{
		data: make(map[int]float64),
	}
}

func (m *MemoryStore) Insert(d types.Distance) error {
	m.data[d.OBUID] += d.Value
	return nil
}

func (m *MemoryStore) Get(id int) (float64, error) {
	dist, ok := m.data[id]
	if !ok {
		return 0.0, fmt.Errorf("could not find distance for obu id %d", id)
	}

	return dist, nil
}
