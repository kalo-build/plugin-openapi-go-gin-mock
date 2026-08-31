package mock

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"
)

// Store provides thread-safe in-memory storage for all API resources.
type Store struct {
	mu        sync.RWMutex
	customers map[string]Customer
	invoices  map[string]Invoice
	payments  map[string]Payment
}

// NewStore creates a new empty Store.
func NewStore() *Store {
	return &Store{
		customers: make(map[string]Customer),
		invoices:  make(map[string]Invoice),
		payments:  make(map[string]Payment),
	}
}

func generateID(prefix string) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	id := make([]byte, 14)
	for i := range id {
		id[i] = charset[rand.Intn(len(charset))]
	}
	return fmt.Sprintf("%s_%s", prefix, string(id))
}

func (s *Store) ListCustomers(limit, offset int) ([]Customer, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var all []Customer
	for _, item := range s.customers {
		all = append(all, item)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Id < all[j].Id })
	total := len(all)
	if offset >= total {
		return []Customer{}, total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return all[offset:end], total
}

func (s *Store) CreateCustomer(item Customer) Customer {
	s.mu.Lock()
	defer s.mu.Unlock()
	if item.Id == "" {
		item.Id = generateID("cus")
	}
	item.CreatedAt = time.Now().UTC()
	s.customers[item.Id] = item
	return item
}

func (s *Store) GetCustomer(id string) (Customer, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.customers[id]
	return item, ok
}

func (s *Store) UpdateCustomer(id string, updates Customer) (Customer, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.customers[id]
	if !ok {
		return Customer{}, false
	}
	updates.Id = existing.Id
	updates.CreatedAt = existing.CreatedAt
	s.customers[id] = updates
	return s.customers[id], true
}

func (s *Store) ListInvoices(limit, offset int, filterField, filterValue string) ([]Invoice, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var all []Invoice
	for _, item := range s.invoices {
		if filterValue != "" && item.CustomerId != filterValue {
			continue
		}
		all = append(all, item)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Id < all[j].Id })
	total := len(all)
	if offset >= total {
		return []Invoice{}, total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return all[offset:end], total
}

func (s *Store) CreateInvoice(item Invoice) Invoice {
	s.mu.Lock()
	defer s.mu.Unlock()
	if item.Id == "" {
		item.Id = generateID("inv")
	}
	if item.Status == "" {
		item.Status = "open"
	}
	item.CreatedAt = time.Now().UTC()
	s.invoices[item.Id] = item
	return item
}

func (s *Store) GetInvoice(id string) (Invoice, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.invoices[id]
	return item, ok
}

func (s *Store) FinalizeInvoice(id string) (Invoice, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.invoices[id]
	if !ok {
		return Invoice{}, false
	}
	item.Status = "finalized"
	s.invoices[id] = item
	return item, true
}

func (s *Store) VoidInvoice(id string) (Invoice, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	item, ok := s.invoices[id]
	if !ok {
		return Invoice{}, false
	}
	item.Status = "void"
	s.invoices[id] = item
	return item, true
}

func (s *Store) ListPayments(limit, offset int, filterField, filterValue string) ([]Payment, int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var all []Payment
	for _, item := range s.payments {
		if filterValue != "" && item.InvoiceId != filterValue {
			continue
		}
		all = append(all, item)
	}
	sort.Slice(all, func(i, j int) bool { return all[i].Id < all[j].Id })
	total := len(all)
	if offset >= total {
		return []Payment{}, total
	}
	end := offset + limit
	if end > total {
		end = total
	}
	return all[offset:end], total
}

func (s *Store) CreatePayment(item Payment) Payment {
	s.mu.Lock()
	defer s.mu.Unlock()
	if item.Id == "" {
		item.Id = generateID("pay")
	}
	if item.Status == "" {
		item.Status = "pending"
	}
	item.CreatedAt = time.Now().UTC()
	s.payments[item.Id] = item
	return item
}

func (s *Store) GetPayment(id string) (Payment, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.payments[id]
	return item, ok
}

// Reset clears all data from the store.
func (s *Store) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.customers = make(map[string]Customer)
	s.invoices = make(map[string]Invoice)
	s.payments = make(map[string]Payment)
}
