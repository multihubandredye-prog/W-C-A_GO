// Package liststore persists paginated interactive lists so a "see more" tap
// can be answered with the next page long after the original request finished.
//
// It mirrors the pollstore pattern: an in-memory map guarded by a mutex and
// flushed to a JSON file, so pending catalogues survive a server restart.
// Without persistence a restart would turn every pending "see more" row into a
// dead end for the customer.
package liststore

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/aldinokemal/go-whatsapp-web-multidevice/config"
	"github.com/sirupsen/logrus"
)

// DefaultTTL is how long a paginated catalogue stays navigable.
const DefaultTTL = 7 * 24 * time.Hour

// Row is a single selectable item of a stored catalogue.
type Row struct {
	RowID       string `json:"row_id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

// PagedList holds everything needed to render any page of a catalogue after
// the original API call is gone.
type PagedList struct {
	// Chat is the recipient JID, used to send the follow-up pages.
	Chat string `json:"chat"`
	// DeviceID identifies which session must send the next page.
	DeviceID string `json:"device_id"`

	Title       string `json:"title,omitempty"`
	Description string `json:"description"`
	Footer      string `json:"footer,omitempty"`
	ButtonText  string `json:"button_text"`
	SectionName string `json:"section_name,omitempty"`

	// Rows is the full catalogue; pages are sliced from it on demand.
	Rows []Row `json:"rows"`

	PageSize          int    `json:"page_size"`
	PaginationLabel   string `json:"pagination_label"`
	ForwardPagination bool   `json:"forward_pagination"`

	CreatedAt time.Time `json:"created_at"`
}

// TotalPages reports how many pages the catalogue spans.
//
// Every page except the last spends one row on the navigation entry, so a page
// carries PageSize items plus that row. The last page needs no navigation and
// therefore holds up to PageSize+1 items.
func (p PagedList) TotalPages() int {
	if p.PageSize <= 0 || len(p.Rows) == 0 {
		return 1
	}
	total := len(p.Rows) / p.PageSize
	if len(p.Rows)%p.PageSize != 0 {
		total++
	}
	if total == 0 {
		total = 1
	}
	return total
}

// Page returns the catalogue rows for the given 1-based page, without the
// navigation entry. The second value reports whether further pages exist.
func (p PagedList) Page(page int) ([]Row, bool) {
	if page < 1 {
		page = 1
	}
	start := (page - 1) * p.PageSize
	if start >= len(p.Rows) {
		return nil, false
	}

	end := start + p.PageSize
	if end > len(p.Rows) {
		end = len(p.Rows)
	}

	// Absorb a trailing remainder that would otherwise need a page of its own
	// carrying a single item: the final page may hold one extra row because it
	// spends nothing on navigation.
	if remaining := len(p.Rows) - end; remaining == 1 {
		end = len(p.Rows)
	}

	return p.Rows[start:end], end < len(p.Rows)
}

type entry struct {
	Data      PagedList `json:"data"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Store keeps paginated catalogues keyed by the message ID that carried the
// page the customer is looking at.
type Store struct {
	filePath string
	ttl      time.Duration
	mu       sync.RWMutex
	data     map[string]entry
}

// New creates a store backed by filePath, loading any previous contents.
func New(filePath string, ttl time.Duration) *Store {
	if ttl <= 0 {
		ttl = DefaultTTL
	}
	s := &Store{
		filePath: filePath,
		ttl:      ttl,
		data:     make(map[string]entry),
	}
	s.load()
	return s
}

func (s *Store) load() {
	s.mu.Lock()
	defer s.mu.Unlock()

	file, err := os.ReadFile(s.filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			logrus.Errorf("failed to read list store file: %v", err)
		}
		return
	}

	if err := json.Unmarshal(file, &s.data); err != nil {
		logrus.Errorf("failed to unmarshal list store data: %v", err)
		return
	}

	// Drop anything that expired while the process was down.
	now := time.Now()
	for key, item := range s.data {
		if now.After(item.ExpiresAt) {
			delete(s.data, key)
		}
	}
}

func (s *Store) save() {
	s.mu.RLock()
	file, err := json.MarshalIndent(s.data, "", "  ")
	s.mu.RUnlock()
	if err != nil {
		logrus.Errorf("failed to marshal list store data: %v", err)
		return
	}

	if err := os.WriteFile(s.filePath, file, 0644); err != nil {
		logrus.Errorf("failed to write list store file: %v", err)
	}
}

// Save associates a catalogue with the message ID of the page just sent.
func (s *Store) Save(messageID string, data PagedList) {
	now := time.Now()
	// Stamp on every save so the newest page of a conversation is always the
	// one the chat fallback resolves to.
	data.CreatedAt = now

	s.mu.Lock()
	s.data[messageID] = entry{Data: data, ExpiresAt: now.Add(s.ttl)}
	s.pruneLocked()
	s.mu.Unlock()
	s.save()
}

// Get returns the catalogue linked to a message ID, if it has not expired.
func (s *Store) Get(messageID string) (PagedList, bool) {
	s.mu.RLock()
	item, ok := s.data[messageID]
	s.mu.RUnlock()

	if !ok {
		return PagedList{}, false
	}
	if time.Now().After(item.ExpiresAt) {
		s.Delete(messageID)
		return PagedList{}, false
	}
	return item.Data, true
}

// GetLatestForChat returns the most recent catalogue still pending for a chat.
//
// Used as a fallback when a navigation tap arrives without the quoted message
// id: some clients omit it, and without this the customer would tap "see more"
// and get nothing back. Matching by chat is safe because only one paginated
// catalogue is pending per conversation at a time.
func (s *Store) GetLatestForChat(chat string) (string, PagedList, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()
	var bestKey string
	var best entry
	found := false

	for key, item := range s.data {
		if item.Data.Chat != chat || now.After(item.ExpiresAt) {
			continue
		}
		if !found || item.Data.CreatedAt.After(best.Data.CreatedAt) {
			bestKey, best, found = key, item, true
		}
	}

	if !found {
		return "", PagedList{}, false
	}
	return bestKey, best.Data, true
}

// Delete removes a catalogue, used once the customer reaches the last page.
func (s *Store) Delete(messageID string) {
	s.mu.Lock()
	delete(s.data, messageID)
	s.mu.Unlock()
	s.save()
}

// pruneLocked drops expired entries. Callers must hold the write lock.
func (s *Store) pruneLocked() {
	now := time.Now()
	for key, item := range s.data {
		if now.After(item.ExpiresAt) {
			delete(s.data, key)
		}
	}
}

// Default is the process-wide store used by the send and event layers.
var Default *Store

func init() {
	if _, err := os.Stat(config.PathStorages); os.IsNotExist(err) {
		if err = os.MkdirAll(config.PathStorages, 0755); err != nil {
			logrus.Fatalf("Failed to create storage directory: %v", err)
		}
	}
	Default = New(config.PathStorages+"/list_store.json", DefaultTTL)
}
