package liststore

import (
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeRows(n int) []Row {
	rows := make([]Row, 0, n)
	for i := 1; i <= n; i++ {
		rows = append(rows, Row{
			RowID: fmt.Sprintf("produto_%d", i),
			Title: fmt.Sprintf("Produto %d", i),
		})
	}
	return rows
}

// TestPagination walks the page maths for the catalogue sizes discussed with
// the user: 30 items in pages of 9, and 15 items in pages of 5.
func TestPagination(t *testing.T) {
	tests := []struct {
		name          string
		rows          int
		pageSize      int
		wantPages     int
		wantPerPage   []int
		wantLastNoNav bool
	}{
		{
			name:        "30 rows in pages of 9",
			rows:        30,
			pageSize:    9,
			wantPages:   4,
			wantPerPage: []int{9, 9, 9, 3},
		},
		{
			name:        "15 rows in pages of 5",
			rows:        15,
			pageSize:    5,
			wantPages:   3,
			wantPerPage: []int{5, 5, 5},
		},
		{
			name:        "exactly one page",
			rows:        9,
			pageSize:    9,
			wantPages:   1,
			wantPerPage: []int{9},
		},
		{
			name:        "single leftover row joins the previous page",
			rows:        10,
			pageSize:    9,
			wantPages:   2,
			wantPerPage: []int{10},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list := PagedList{Rows: makeRows(tt.rows), PageSize: tt.pageSize}

			total := 0
			for i, want := range tt.wantPerPage {
				page := i + 1
				rows, hasMore := list.Page(page)
				assert.Len(t, rows, want, "page %d row count", page)
				total += len(rows)

				if page == len(tt.wantPerPage) {
					assert.False(t, hasMore, "last page must not offer navigation")
				}
			}

			assert.Equal(t, tt.rows, total, "every catalogue row must appear exactly once")
		})
	}
}

// TestPageSequenceIsContiguous makes sure no product is skipped or repeated.
func TestPageSequenceIsContiguous(t *testing.T) {
	list := PagedList{Rows: makeRows(30), PageSize: 9}

	seen := make([]string, 0, 30)
	for page := 1; ; page++ {
		rows, hasMore := list.Page(page)
		require.NotEmpty(t, rows, "page %d must not be empty", page)
		for _, row := range rows {
			seen = append(seen, row.RowID)
		}
		if !hasMore {
			break
		}
		require.Less(t, page, 10, "pagination must terminate")
	}

	require.Len(t, seen, 30)
	for i, rowID := range seen {
		assert.Equal(t, fmt.Sprintf("produto_%d", i+1), rowID, "order must be preserved")
	}
}

func TestPageOutOfRange(t *testing.T) {
	list := PagedList{Rows: makeRows(9), PageSize: 9}

	rows, hasMore := list.Page(2)
	assert.Empty(t, rows)
	assert.False(t, hasMore)

	// A page below one is clamped rather than panicking.
	rows, _ = list.Page(0)
	assert.Len(t, rows, 9)
}

func TestStoreSaveGetDelete(t *testing.T) {
	path := filepath.Join(t.TempDir(), "list_store.json")
	store := New(path, time.Hour)

	list := PagedList{
		Chat:     "5588999999999@s.whatsapp.net",
		Rows:     makeRows(30),
		PageSize: 9,
	}

	store.Save("MSG1", list)

	got, ok := store.Get("MSG1")
	require.True(t, ok)
	assert.Len(t, got.Rows, 30)
	assert.Equal(t, 4, got.TotalPages())

	store.Delete("MSG1")
	_, ok = store.Get("MSG1")
	assert.False(t, ok)
}

// TestStoreSurvivesRestart is the reason the store is on disk: a pending
// "see more" row must keep working after the process restarts.
func TestStoreSurvivesRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "list_store.json")

	first := New(path, time.Hour)
	first.Save("MSG1", PagedList{Rows: makeRows(30), PageSize: 9})

	reopened := New(path, time.Hour)
	got, ok := reopened.Get("MSG1")
	require.True(t, ok, "catalogue must outlive a restart")
	assert.Len(t, got.Rows, 30)
}

func TestStoreExpiry(t *testing.T) {
	path := filepath.Join(t.TempDir(), "list_store.json")
	store := New(path, time.Nanosecond)

	store.Save("MSG1", PagedList{Rows: makeRows(5), PageSize: 2})
	time.Sleep(2 * time.Millisecond)

	_, ok := store.Get("MSG1")
	assert.False(t, ok, "an expired catalogue must not be returned")
}

// TestGetLatestForChat covers the fallback used when a navigation tap arrives
// without a usable quoted id: the newest pending catalogue of that chat wins.
func TestGetLatestForChat(t *testing.T) {
	path := filepath.Join(t.TempDir(), "list_store.json")
	store := New(path, time.Hour)

	const chat = "5588999999999@s.whatsapp.net"

	store.Save("OLD", PagedList{Chat: chat, Rows: makeRows(30), PageSize: 9})
	time.Sleep(2 * time.Millisecond)
	store.Save("NEW", PagedList{Chat: chat, Rows: makeRows(12), PageSize: 5})
	store.Save("OTHER", PagedList{Chat: "5511888888888@s.whatsapp.net", Rows: makeRows(20), PageSize: 9})

	key, got, ok := store.GetLatestForChat(chat)
	require.True(t, ok)
	assert.Equal(t, "NEW", key, "the most recent catalogue must win")
	assert.Len(t, got.Rows, 12)

	_, _, ok = store.GetLatestForChat("5599777777777@s.whatsapp.net")
	assert.False(t, ok, "an unknown chat has no catalogue")
}

func TestGetLatestForChatIgnoresExpired(t *testing.T) {
	path := filepath.Join(t.TempDir(), "list_store.json")
	store := New(path, time.Nanosecond)

	store.Save("MSG1", PagedList{Chat: "5588999999999@s.whatsapp.net", Rows: makeRows(30), PageSize: 9})
	time.Sleep(2 * time.Millisecond)

	_, _, ok := store.GetLatestForChat("5588999999999@s.whatsapp.net")
	assert.False(t, ok, "expired catalogues must not be resolved")
}
