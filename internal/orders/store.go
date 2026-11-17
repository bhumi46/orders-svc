// Package orders implements the orders API with an in-memory store.
package orders

import (
	"errors"
	"sort"
	"strconv"
	"sync"
)

// Order is an order placed by a user.
type Order struct {
	ID       string `json:"id"`
	UserID   string `json:"user_id"`
	Item     string `json:"item"`
	Quantity int    `json:"quantity"`
}

// ErrNotFound is returned when an order does not exist.
var ErrNotFound = errors.New("order not found")

// Store is a concurrency-safe in-memory order store.
type Store struct {
	mu     sync.RWMutex
	orders map[string]Order
	nextID int
}

// NewStore returns an empty store.
func NewStore() *Store {
	return &Store{orders: map[string]Order{}, nextID: 1}
}

// List returns all orders sorted by numeric ID.
func (s *Store) List() []Order {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Order, 0, len(s.orders))
	for _, o := range s.orders {
		out = append(out, o)
	}
	sort.Slice(out, func(i, j int) bool {
		a, _ := strconv.Atoi(out[i].ID)
		b, _ := strconv.Atoi(out[j].ID)
		return a < b
	})
	return out
}

// Get returns one order or ErrNotFound.
func (s *Store) Get(id string) (Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, ok := s.orders[id]
	if !ok {
		return Order{}, ErrNotFound
	}
	return o, nil
}

// Create assigns an ID and stores the order.
func (s *Store) Create(o Order) Order {
	s.mu.Lock()
	defer s.mu.Unlock()
	o.ID = strconv.Itoa(s.nextID)
	s.nextID++
	s.orders[o.ID] = o
	return o
}
